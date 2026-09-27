package archive

import (
	"archive/zip"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestExtract_roundTrip(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "jump.webm"), "jump-bytes")
	mustWrite(t, filepath.Join(root, "unused.mp3"), "leave-me")
	mustWrite(t, filepath.Join(root, "clips", "dance.mp4"), "dance-bytes")
	srcDoc := filepath.Join(t.TempDir(), "custom.yml")
	const body = "commands:\n  - name: jump\n    file: jump.webm\n  - name: dance\n    file: clips/dance.mp4\n"
	mustWrite(t, srcDoc, body)
	zipPath := filepath.Join(t.TempDir(), "share.zip")
	err := Write(Spec{
		Dest:         zipPath,
		Document:     srcDoc,
		DocumentName: "commands.yaml",
		Root:         root,
		Files:        []string{"jump.webm", "clips/dance.mp4"},
	})
	if err != nil {
		t.Fatal(err)
	}
	got, err := ReadRootFile(zipPath, "commands.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != body {
		t.Fatalf("root = %q", got)
	}
	dest := t.TempDir()
	if err := Extract(zipPath, dest, "commands.yaml"); err != nil {
		t.Fatal(err)
	}
	if fileText(t, filepath.Join(dest, "commands.yaml")) != body {
		t.Fatal("yaml")
	}
	if fileText(t, filepath.Join(dest, "jump.webm")) != "jump-bytes" {
		t.Fatal("jump")
	}
	if fileText(t, filepath.Join(dest, "clips", "dance.mp4")) != "dance-bytes" {
		t.Fatal("dance")
	}
	if _, err := os.Stat(filepath.Join(dest, "unused.mp3")); !os.IsNotExist(err) {
		t.Fatal("unused file was unpacked")
	}
}

func TestExtract_requiresRootDocument(t *testing.T) {
	zipPath := writeRawZip(t, map[string]string{
		"clips/commands.yaml": "nope\n",
		"jump.webm":           "x",
	})
	dest := t.TempDir()
	err := Extract(zipPath, dest, "commands.yaml")
	if !errors.Is(err, ErrRootDocument) {
		t.Fatalf("err = %v", err)
	}
	if _, statErr := os.Stat(filepath.Join(dest, "jump.webm")); !os.IsNotExist(statErr) {
		t.Fatal("extracted without commands.yaml")
	}
	if _, err := ReadRootFile(zipPath, "commands.yaml"); !errors.Is(err, ErrRootDocument) {
		t.Fatalf("read err = %v", err)
	}
}

func TestExtract_rejectsEscape(t *testing.T) {
	zipPath := writeRawZip(t, map[string]string{
		"commands.yaml": "commands: []\n",
		"../secret.mp3": "secret",
	})
	dest := t.TempDir()
	err := Extract(zipPath, dest, "commands.yaml")
	if !errors.Is(err, ErrOutside) {
		t.Fatalf("err = %v", err)
	}
	if _, statErr := os.Stat(filepath.Join(dest, "commands.yaml")); !os.IsNotExist(statErr) {
		t.Fatal("partial extract")
	}
}

func TestExtract_rejectsSymlinkEntry(t *testing.T) {
	zipPath := filepath.Join(t.TempDir(), "bad.zip")
	f, err := os.Create(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	if _, err := zw.Create("commands.yaml"); err != nil {
		t.Fatal(err)
	}
	hdr := &zip.FileHeader{Name: "link.mp3"}
	hdr.SetMode(os.ModeSymlink | 0o777)
	w, err := zw.CreateHeader(hdr)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte("../secret")); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	dest := t.TempDir()
	err = Extract(zipPath, dest, "commands.yaml")
	if !errors.Is(err, ErrNotFile) {
		t.Fatalf("err = %v", err)
	}
	if _, statErr := os.Stat(filepath.Join(dest, "commands.yaml")); !os.IsNotExist(statErr) {
		t.Fatal("partial extract")
	}
}

func writeRawZip(t *testing.T, files map[string]string) string {
	t.Helper()
	dest := filepath.Join(t.TempDir(), "raw.zip")
	f, err := os.Create(dest)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	for name, body := range files {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(body)); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	return dest
}

func fileText(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
