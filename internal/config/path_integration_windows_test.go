//go:build integration && windows

package config

import (
	"strings"
	"testing"
)

func TestDefaultPath_windows(t *testing.T) {
	p, err := DefaultPath()
	if err != nil {
		t.Fatal(err)
	}
	lower := strings.ToLower(p)
	if !strings.Contains(lower, "appdata") {
		t.Fatalf("path = %q, want AppData", p)
	}
	if !strings.HasSuffix(strings.ReplaceAll(p, "\\", "/"), "ytmemchat/config.json") {
		t.Fatalf("path = %q", p)
	}
}
