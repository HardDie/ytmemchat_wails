package commands

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteCommandsExport(t *testing.T) {
	media := t.TempDir()
	mustWrite(t, filepath.Join(media, "jump.webm"), "jump-bytes")
	mustWrite(t, filepath.Join(media, "unused.mp3"), "unused-bytes")
	mustWrite(t, filepath.Join(media, "clips", "dance.mp4"), "dance-bytes")
	mustWrite(t, filepath.Join(media, "clips", "spare.mp4"), "spare-bytes")

	custom := filepath.Join(t.TempDir(), "my-alerts.yml")
	const yamlBody = "# keep me\ncommands:\n  - name: jump\n    file: jump.webm\n  - name: dance\n    file: clips/dance.mp4\n  - name: again\n    file: jump.webm\n"
	mustWrite(t, custom, yamlBody)

	dest := filepath.Join(t.TempDir(), "share.zip")
	if err := writeCommandsExport(dest, custom, media); err != nil {
		t.Fatal(err)
	}
	got := zipNames(t, dest)
	want := []string{"commands.yaml", "jump.webm", "clips/dance.mp4"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("names = %v", got)
	}
	if body := zipFile(t, dest, "commands.yaml"); body != yamlBody {
		t.Fatalf("yaml = %q", body)
	}
	if body := zipFile(t, dest, "jump.webm"); body != "jump-bytes" {
		t.Fatalf("jump = %q", body)
	}
}

func TestWriteCommandsExport_emptyCollectionOmitsMedia(t *testing.T) {
	media := t.TempDir()
	mustWrite(t, filepath.Join(media, "unused.mp3"), "nope")
	yamlPath := filepath.Join(media, "commands.yaml")
	const yamlBody = "commands: []\n"
	mustWrite(t, yamlPath, yamlBody)
	dest := filepath.Join(t.TempDir(), "share.zip")
	if err := writeCommandsExport(dest, yamlPath, media); err != nil {
		t.Fatal(err)
	}
	got := zipNames(t, dest)
	if len(got) != 1 || got[0] != "commands.yaml" {
		t.Fatalf("names = %v", got)
	}
	if body := zipFile(t, dest, "commands.yaml"); body != yamlBody {
		t.Fatalf("yaml = %q", body)
	}
}

func TestWriteCommandsExport_missingFile(t *testing.T) {
	media := t.TempDir()
	yamlPath := filepath.Join(t.TempDir(), "commands.yaml")
	mustWrite(t, yamlPath, "commands:\n  - name: jump\n    file: jump.webm\n")
	dest := filepath.Join(t.TempDir(), "share.zip")
	err := writeCommandsExport(dest, yamlPath, media)
	if err == nil || !strings.Contains(err.Error(), "jump.webm") || !strings.Contains(err.Error(), "missing") {
		t.Fatalf("err = %v", err)
	}
	if _, statErr := os.Stat(dest); !os.IsNotExist(statErr) {
		t.Fatalf("dest stat = %v", statErr)
	}
}

func TestWriteCommandsExport_rejectsOutside(t *testing.T) {
	media := t.TempDir()
	outside := filepath.Join(t.TempDir(), "secret.mp3")
	mustWrite(t, outside, "secret-bytes")
	yamlPath := filepath.Join(t.TempDir(), "commands.yaml")
	mustWrite(t, yamlPath, "commands:\n  - name: jump\n    file: ../secret.mp3\n")
	dest := filepath.Join(t.TempDir(), "share.zip")
	err := writeCommandsExport(dest, yamlPath, media)
	if err == nil || !strings.Contains(err.Error(), "inside the media folder") {
		t.Fatalf("err = %v", err)
	}
	if _, statErr := os.Stat(dest); !os.IsNotExist(statErr) {
		t.Fatalf("dest stat = %v", statErr)
	}
}

func TestWriteCommandsExport_rejectsCommandsYamlName(t *testing.T) {
	media := t.TempDir()
	mustWrite(t, filepath.Join(media, "commands.yaml"), "not-the-doc")
	yamlPath := filepath.Join(t.TempDir(), "custom.yml")
	mustWrite(t, yamlPath, "commands:\n  - name: jump\n    file: commands.yaml\n")
	err := writeCommandsExport(filepath.Join(t.TempDir(), "share.zip"), yamlPath, media)
	if err == nil || !strings.Contains(err.Error(), "reserved") {
		t.Fatalf("err = %v", err)
	}
}

func TestWriteCommandsExport_missingYAML(t *testing.T) {
	media := t.TempDir()
	err := writeCommandsExport(filepath.Join(t.TempDir(), "share.zip"), filepath.Join(t.TempDir(), "commands.yaml"), media)
	if err == nil || !strings.Contains(err.Error(), "save commands before export") {
		t.Fatalf("err = %v", err)
	}
	if err := writeCommandsExport(filepath.Join(t.TempDir(), "share.zip"), "", media); err == nil || !strings.Contains(err.Error(), "commands.yaml path") {
		t.Fatalf("path err = %v", err)
	}
	yamlPath := filepath.Join(t.TempDir(), "commands.yaml")
	mustWrite(t, yamlPath, "commands: []\n")
	if err := writeCommandsExport(filepath.Join(t.TempDir(), "share.zip"), yamlPath, "  "); err == nil || !strings.Contains(err.Error(), "media folder") {
		t.Fatalf("media err = %v", err)
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
