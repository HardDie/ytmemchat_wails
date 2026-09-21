package hotkey

import "testing"

func TestBinding_disabled(t *testing.T) {
	var b Binding
	b.Sync(false, DefaultChord, func() { t.Fatal("press") })
	if b.Err() != "" {
		t.Fatal(b.Err())
	}
	b.Stop()
}

func TestBinding_parseError(t *testing.T) {
	var b Binding
	b.Sync(true, "I", nil)
	if b.Err() == "" {
		t.Fatal("want parse error")
	}
	b.Stop()
	if b.Err() != "" {
		t.Fatalf("after Stop: %q", b.Err())
	}
}

func TestBinding_enabled(t *testing.T) {
	var b Binding
	b.Sync(true, DefaultChord, func() {})
	b.Stop()
}
