//go:build integration && linux

package config

import (
	"strings"
	"testing"
)

func TestDefaultPath_linux(t *testing.T) {
	p, err := DefaultPath()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(p, ".config") {
		t.Fatalf("path = %q, want .config", p)
	}
	if !strings.HasSuffix(p, "ytmemchat/config.json") {
		t.Fatalf("path = %q", p)
	}
}
