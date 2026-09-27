package archive

import (
	"archive/zip"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWrite_selectedFilesOnly(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "jump.webm"), "jump-bytes")
	mustWrite(t, filepath.Join(root, "unused.mp3"), "unused-bytes")
	mustWrite(t, filepath.Join(root, "clips", "dance.mp4"), "dance-bytes")
	mustWrite(t, filepath.Join(root, "clips", "spare.mp4"), "spare-bytes")

	src := filepath.Join(t.TempDir(), "my-alerts.yml")
	const body = "# keep me\n"
	mustWrite(t, src, body)

	dest := filepath.Join(t.TempDir(), "share.zip")
	err := Write(Spec{
		Dest:         dest,
		Document:     src,
		DocumentName: "notes.txt",
		Root:         root,
		Files:        []string{"jump.webm", "clips/dance.mp4", "jump.webm", ""},
	})
	if err != nil {
		t.Fatal(err)
	}
	got := zipNames(t, dest)
	if strings.Join(got, ",") != "notes.txt,jump.webm,clips/dance.mp4" {
		t.Fatalf("names = %v", got)
	}
	if zipFile(t, dest, "notes.txt") != body {
		t.Fatalf("doc = %q", zipFile(t, dest, "notes.txt"))
	}
	if zipFile(t, dest, "jump.webm") != "jump-bytes" {
		t.Fatal("jump bytes")
	}
	if zipFile(t, dest, "clips/dance.mp4") != "dance-bytes" {
		t.Fatal("dance bytes")
	}
}

func TestWrite_documentOnly(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "unused.mp3"), "nope")
	src := filepath.Join(root, "doc.yml")
	mustWrite(t, src, "commands: []\n")
	dest := filepath.Join(t.TempDir(), "share.zip")
	if err := Write(Spec{Dest: dest, Document: src, DocumentName: "commands.yaml", Root: root}); err != nil {
		t.Fatal(err)
	}
	got := zipNames(t, dest)
	if len(got) != 1 || got[0] != "commands.yaml" {
		t.Fatalf("names = %v", got)
	}
}

func TestWrite_absoluteInsideIsRelative(t *testing.T) {
	root := t.TempDir()
	abs := filepath.Join(root, "clips", "dance.mp4")
	mustWrite(t, abs, "dance-bytes")
	src := filepath.Join(t.TempDir(), "doc.txt")
	mustWrite(t, src, "doc")
	dest := filepath.Join(t.TempDir(), "share.zip")
	err := Write(Spec{Dest: dest, Document: src, DocumentName: "doc.txt", Root: root, Files: []string{abs}})
	if err != nil {
		t.Fatal(err)
	}
	if got := zipNames(t, dest); strings.Join(got, ",") != "doc.txt,clips/dance.mp4" {
		t.Fatalf("names = %v", got)
	}
}

func TestWrite_rejectsOutside(t *testing.T) {
	root := t.TempDir()
	outside := filepath.Join(t.TempDir(), "secret.mp3")
	mustWrite(t, outside, "secret-bytes")
	src := filepath.Join(t.TempDir(), "doc.txt")
	mustWrite(t, src, "doc")
	dest := filepath.Join(t.TempDir(), "share.zip")
	spec := Spec{Dest: dest, Document: src, DocumentName: "doc.txt", Root: root, Files: []string{"../secret.mp3"}}
	err := Write(spec)
	if !errors.Is(err, ErrOutside) {
		t.Fatalf("err = %v", err)
	}
	if _, statErr := os.Stat(dest); !os.IsNotExist(statErr) {
		t.Fatalf("dest stat = %v", statErr)
	}
	spec.Files = []string{outside}
	if err := Write(spec); !errors.Is(err, ErrOutside) {
		t.Fatalf("abs err = %v", err)
	}
}

func TestWrite_rejectsSymlinkOutside(t *testing.T) {
	root := t.TempDir()
	outside := filepath.Join(t.TempDir(), "secret.mp3")
	mustWrite(t, outside, "secret-bytes")
	if err := os.Symlink(outside, filepath.Join(root, "link.mp3")); err != nil {
		t.Fatal(err)
	}
	src := filepath.Join(t.TempDir(), "doc.txt")
	mustWrite(t, src, "doc")
	dest := filepath.Join(t.TempDir(), "share.zip")
	err := Write(Spec{Dest: dest, Document: src, DocumentName: "doc.txt", Root: root, Files: []string{"link.mp3"}})
	if !errors.Is(err, ErrOutside) {
		t.Fatalf("err = %v", err)
	}
	if _, statErr := os.Stat(dest); !os.IsNotExist(statErr) {
		t.Fatalf("dest stat = %v", statErr)
	}
}

func TestWrite_keepsSymlinkInside(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "real.mp3"), "real-bytes")
	if err := os.Symlink(filepath.Join(root, "real.mp3"), filepath.Join(root, "alias.mp3")); err != nil {
		t.Fatal(err)
	}
	src := filepath.Join(t.TempDir(), "doc.txt")
	mustWrite(t, src, "doc")
	dest := filepath.Join(t.TempDir(), "share.zip")
	err := Write(Spec{Dest: dest, Document: src, DocumentName: "doc.txt", Root: root, Files: []string{"alias.mp3"}})
	if err != nil {
		t.Fatal(err)
	}
	if got := zipNames(t, dest); strings.Join(got, ",") != "doc.txt,alias.mp3" {
		t.Fatalf("names = %v", got)
	}
	if zipFile(t, dest, "alias.mp3") != "real-bytes" {
		t.Fatal("alias bytes")
	}
}

func TestWrite_rejectsDocumentName(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "notes.txt"), "not-the-doc")
	src := filepath.Join(t.TempDir(), "custom.yml")
	mustWrite(t, src, "doc")
	err := Write(Spec{
		Dest:         filepath.Join(t.TempDir(), "share.zip"),
		Document:     src,
		DocumentName: "notes.txt",
		Root:         root,
		Files:        []string{"notes.txt"},
	})
	if !errors.Is(err, ErrReserved) {
		t.Fatalf("err = %v", err)
	}
}

func TestWrite_missingFile(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(t.TempDir(), "doc.txt")
	mustWrite(t, src, "doc")
	dest := filepath.Join(t.TempDir(), "share.zip")
	err := Write(Spec{Dest: dest, Document: src, DocumentName: "doc.txt", Root: root, Files: []string{"missing.mp3"}})
	if !errors.Is(err, ErrMissing) {
		t.Fatalf("err = %v", err)
	}
	var pe *PathError
	if !errors.As(err, &pe) || pe.Name != "missing.mp3" {
		t.Fatalf("path err = %#v", err)
	}
	if _, statErr := os.Stat(dest); !os.IsNotExist(statErr) {
		t.Fatalf("dest stat = %v", statErr)
	}
}

func TestWrite_missingDocument(t *testing.T) {
	err := Write(Spec{
		Dest:         filepath.Join(t.TempDir(), "share.zip"),
		Document:     filepath.Join(t.TempDir(), "missing.txt"),
		DocumentName: "doc.txt",
	})
	if !errors.Is(err, ErrDocumentMissing) {
		t.Fatalf("err = %v", err)
	}
}

func TestWrite_addsZipSuffix(t *testing.T) {
	src := filepath.Join(t.TempDir(), "doc.txt")
	mustWrite(t, src, "doc")
	dest := filepath.Join(t.TempDir(), "share")
	if err := Write(Spec{Dest: dest, Document: src, DocumentName: "doc.txt"}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(dest + ".zip"); err != nil {
		t.Fatal(err)
	}
}

func TestValidate_doesNotWrite(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(t.TempDir(), "doc.txt")
	mustWrite(t, src, "doc")
	dest := filepath.Join(t.TempDir(), "share.zip")
	spec := Spec{Dest: dest, Document: src, DocumentName: "doc.txt", Root: root, Files: []string{"nope.mp3"}}
	if err := Validate(spec); !errors.Is(err, ErrMissing) {
		t.Fatalf("err = %v", err)
	}
	if _, err := os.Stat(dest); !os.IsNotExist(err) {
		t.Fatal("validate wrote a zip")
	}
}

func mustWrite(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}

func zipNames(t *testing.T, path string) []string {
	t.Helper()
	r, err := zip.OpenReader(path)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	names := make([]string, 0, len(r.File))
	for _, f := range r.File {
		names = append(names, f.Name)
	}
	return names
}

func zipFile(t *testing.T, path, name string) string {
	t.Helper()
	r, err := zip.OpenReader(path)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	for _, f := range r.File {
		if f.Name != name {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			t.Fatal(err)
		}
		defer rc.Close()
		buf, err := io.ReadAll(rc)
		if err != nil {
			t.Fatal(err)
		}
		return string(buf)
	}
	t.Fatalf("missing %s", name)
	return ""
}
