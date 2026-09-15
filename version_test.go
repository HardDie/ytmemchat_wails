package main

import "testing"

func TestVersionString_emptyIsDev(t *testing.T) {
	old := buildVersion
	t.Cleanup(func() { buildVersion = old })
	buildVersion = "  "
	if versionString() != "dev" {
		t.Fatalf("got %q", versionString())
	}
	buildVersion = "v1.2.3"
	if versionString() != "v1.2.3" {
		t.Fatalf("got %q", versionString())
	}
}
