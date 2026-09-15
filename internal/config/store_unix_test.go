//go:build !windows

package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSave_unixFileMode(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	st := NewStore(path)
	if err := st.Save(Defaults()); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("perm = %o, want 0600", info.Mode().Perm())
	}
}
