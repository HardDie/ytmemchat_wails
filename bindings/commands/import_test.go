package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/HardDie/ytmemchat_wails/pkg/archive"
)

func TestExtractImportArchive(t *testing.T) {
	media := t.TempDir()
	mustWrite(t, filepath.Join(media, "jump.webm"), "jump-bytes")
	mustWrite(t, filepath.Join(media, "clips", "dance.mp4"), "dance-bytes")
	yamlPath := filepath.Join(t.TempDir(), "mine.yml")
	const body = "commands:\n  - name: jump\n    file: jump.webm\n  - name: dance\n    file: clips/dance.mp4\n"
	mustWrite(t, yamlPath, body)
	zipPath := filepath.Join(t.TempDir(), "share.zip")
	err := archive.Write(archive.Spec{
		Dest:         zipPath,
		Document:     yamlPath,
		DocumentName: "commands.yaml",
		Root:         media,
		Files:        []string{"jump.webm", "clips/dance.mp4"},
	})
	if err != nil {
		t.Fatal(err)
	}
	dest := t.TempDir()
	if err := extractImportArchive(zipPath, dest); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(dest, "commands.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != body {
		t.Fatalf("yaml = %q", got)
	}
	if _, err := os.Stat(filepath.Join(dest, "clips", "dance.mp4")); err != nil {
		t.Fatal(err)
	}
}

func TestValidateImportArchive_requiresRootYAML(t *testing.T) {
	zipPath := filepath.Join(t.TempDir(), "bad.zip")
	err := archive.Write(archive.Spec{
		Dest:         zipPath,
		Document:     writeTemp(t, "commands: []\n"),
		DocumentName: "notes.txt",
		Root:         t.TempDir(),
	})
	if err != nil {
		t.Fatal(err)
	}
	err = validateImportArchive(zipPath)
	if err == nil || !strings.Contains(err.Error(), "root of the archive") {
		t.Fatalf("err = %v", err)
	}
	broken := filepath.Join(t.TempDir(), "broken.zip")
	err = archive.Write(archive.Spec{
		Dest:         broken,
		Document:     writeTemp(t, "commands: [\n"),
		DocumentName: "commands.yaml",
		Root:         t.TempDir(),
	})
	if err != nil {
		t.Fatal(err)
	}
	err = validateImportArchive(broken)
	if err == nil || !strings.Contains(err.Error(), "commands.yaml") {
		t.Fatalf("yaml err = %v", err)
	}
}

func writeTemp(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "doc.yml")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}
