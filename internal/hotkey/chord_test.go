package hotkey

import (
	"errors"
	"testing"
)

func TestParse_default(t *testing.T) {
	c, err := Parse(DefaultChord)
	if err != nil {
		t.Fatal(err)
	}
	if !c.Ctrl || !c.Shift || c.Alt || c.Super || c.Key != "I" {
		t.Fatalf("%+v", c)
	}
	if c.String() != DefaultChord {
		t.Fatalf("String = %q", c.String())
	}
}

func TestParse_aliases(t *testing.T) {
	c, err := Parse("control + shift + i")
	if err != nil || c.String() != DefaultChord {
		t.Fatalf("%+v %v", c, err)
	}
	c, err = Parse("Cmd+Alt+F8")
	if err != nil || !c.Super || !c.Alt || c.Key != "F8" {
		t.Fatalf("%+v %v", c, err)
	}
	if c.String() != "Cmd+Alt+F8" {
		t.Fatalf("String = %q", c.String())
	}
}

func TestParse_rejectsBareKey(t *testing.T) {
	_, err := Parse("I")
	if !errors.Is(err, ErrNeedModifier) {
		t.Fatalf("err = %v", err)
	}
}

func TestParse_unknown(t *testing.T) {
	_, err := Parse("Ctrl+Period")
	if !errors.Is(err, ErrUnknownPart) {
		t.Fatalf("err = %v", err)
	}
	_, err = Parse("")
	if !errors.Is(err, ErrEmpty) {
		t.Fatalf("err = %v", err)
	}
}
