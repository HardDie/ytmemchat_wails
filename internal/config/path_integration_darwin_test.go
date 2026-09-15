//go:build integration && darwin

package config

import (
	"strings"
	"testing"
)

func TestDefaultPath_darwin(t *testing.T) {
	p, err := DefaultPath()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(p, "Application Support") {
		t.Fatalf("path = %q, want Application Support", p)
	}
	if !strings.HasSuffix(p, "ytmemchat/config.json") {
		t.Fatalf("path = %q", p)
	}
}
