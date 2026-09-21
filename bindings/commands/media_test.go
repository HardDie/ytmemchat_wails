package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMediaRelativePath(t *testing.T) {
	root := t.TempDir()
	sub := filepath.Join(root, "videos")
	if err := os.Mkdir(sub, 0o700); err != nil {
		t.Fatal(err)
	}
	inside := filepath.Join(sub, "huh.webm")
	if err := os.WriteFile(inside, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := mediaRelativePath(root, inside)
	if err != nil {
		t.Fatal(err)
	}
	if got != "videos/huh.webm" {
		t.Fatalf("got %q", got)
	}
	outside := filepath.Join(t.TempDir(), "nope.mp3")
	if err := os.WriteFile(outside, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := mediaRelativePath(root, outside); err == nil || !strings.Contains(err.Error(), "inside the media folder") {
		t.Fatalf("outside err = %v", err)
	}
	if _, err := mediaRelativePath("", inside); err == nil {
		t.Fatal("empty root")
	}
}
