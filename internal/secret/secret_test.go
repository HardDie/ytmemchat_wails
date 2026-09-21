package secret

import (
	"errors"
	"testing"
)

func TestMemory_roundTrip(t *testing.T) {
	var m Memory
	if !m.Available() {
		t.Fatal("memory vault should be available")
	}
	if _, err := m.Get(); !errors.Is(err, ErrNotFound) {
		t.Fatalf("get = %v", err)
	}
	if err := m.Set("secret"); err != nil {
		t.Fatal(err)
	}
	got, err := m.Get()
	if err != nil || got != "secret" {
		t.Fatalf("get %q %v", got, err)
	}
	if err := m.Delete(); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Get(); !errors.Is(err, ErrNotFound) {
		t.Fatalf("after delete %v", err)
	}
}

func TestUnavailable(t *testing.T) {
	var u Unavailable
	if u.Available() {
		t.Fatal("unavailable")
	}
	if _, err := u.Get(); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("get %v", err)
	}
	if err := u.Set("x"); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("set %v", err)
	}
}

func TestMemory_down(t *testing.T) {
	var m Memory
	m.Down()
	if m.Available() {
		t.Fatal("down")
	}
	if err := m.Set("x"); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("set %v", err)
	}
}

func TestMemory_failSet(t *testing.T) {
	var m Memory
	m.FailSet(ErrUnavailable)
	if !m.Available() {
		t.Fatal("still available")
	}
	if err := m.Set("x"); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("set %v", err)
	}
}
