package hotkey

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// DefaultChord is the first-launch interrupt shortcut (Control+Shift+I).
const DefaultChord = "Ctrl+Shift+I"

var (
	// ErrEmpty is returned when the chord has no key.
	ErrEmpty = fmt.Errorf("hotkey: empty shortcut")
	// ErrNeedModifier is returned when the key has no Ctrl/Cmd/Alt/Shift.
	ErrNeedModifier = fmt.Errorf("hotkey: add Ctrl, Cmd, Alt, or Shift so a single key is not grabbed globally")
	// ErrUnknownPart is returned for an unrecognized token.
	ErrUnknownPart = fmt.Errorf("hotkey: unknown key or modifier")
)

// Chord is a modifier set plus one key name (A–Z, 0–9, F1–F12, Space, …).
type Chord struct {
	Ctrl  bool
	Super bool // Cmd on macOS, Win/Super elsewhere
	Alt   bool
	Shift bool
	Key   string
}

// Parse accepts strings like "ctrl+shift+i" or "Cmd+Alt+F8".
func Parse(s string) (Chord, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Chord{}, ErrEmpty
	}
	var c Chord
	seenKey := false
	for _, raw := range strings.Split(s, "+") {
		p := strings.TrimSpace(raw)
		if p == "" {
			return Chord{}, ErrUnknownPart
		}
		switch normalizePart(p) {
		case "CTRL":
			c.Ctrl = true
		case "SUPER":
			c.Super = true
		case "ALT":
			c.Alt = true
		case "SHIFT":
			c.Shift = true
		default:
			key, ok := canonicalKey(p)
			if !ok {
				return Chord{}, fmt.Errorf("%w: %s", ErrUnknownPart, p)
			}
			if seenKey {
				return Chord{}, fmt.Errorf("%w: extra key %s", ErrUnknownPart, p)
			}
			c.Key = key
			seenKey = true
		}
	}
	if c.Key == "" {
		return Chord{}, ErrEmpty
	}
	if !c.Ctrl && !c.Super && !c.Alt && !c.Shift {
		return Chord{}, ErrNeedModifier
	}
	return c, nil
}

// String is a stable form such as "Ctrl+Shift+I".
func (c Chord) String() string {
	var parts []string
	if c.Ctrl {
		parts = append(parts, "Ctrl")
	}
	if c.Super {
		parts = append(parts, "Cmd")
	}
	if c.Alt {
		parts = append(parts, "Alt")
	}
	if c.Shift {
		parts = append(parts, "Shift")
	}
	if c.Key != "" {
		parts = append(parts, c.Key)
	}
	return strings.Join(parts, "+")
}

func normalizePart(p string) string {
	u := strings.ToUpper(strings.TrimSpace(p))
	switch u {
	case "CTRL", "CONTROL", "CONTROLLEFT", "CONTROLRIGHT":
		return "CTRL"
	case "CMD", "COMMAND", "SUPER", "WIN", "WINDOWS", "META", "METALEFT", "METARIGHT":
		return "SUPER"
	case "ALT", "OPTION", "OPT", "ALTLEFT", "ALTRIGHT":
		return "ALT"
	case "SHIFT", "SHIFTLEFT", "SHIFTRIGHT":
		return "SHIFT"
	default:
		return u
	}
}

func canonicalKey(p string) (string, bool) {
	u := strings.ToUpper(strings.TrimSpace(p))
	u = strings.TrimPrefix(u, "KEY")
	u = strings.TrimPrefix(u, "DIGIT")
	u = strings.TrimPrefix(u, "ARROW")
	switch u {
	case "SPACE", " ":
		return "Space", true
	case "ESC", "ESCAPE":
		return "Escape", true
	case "TAB":
		return "Tab", true
	case "ENTER", "RETURN":
		return "Enter", true
	case "DELETE", "DEL", "BACKSPACE":
		return "Delete", true
	case "LEFT":
		return "Left", true
	case "RIGHT":
		return "Right", true
	case "UP":
		return "Up", true
	case "DOWN":
		return "Down", true
	}
	if len(u) >= 2 && u[0] == 'F' {
		n := u[1:]
		ok := true
		for _, r := range n {
			if r < '0' || r > '9' {
				ok = false
				break
			}
		}
		if ok {
			switch n {
			case "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12":
				return "F" + n, true
			}
		}
	}
	if utf8.RuneCountInString(u) == 1 {
		r, _ := utf8.DecodeRuneInString(u)
		if r >= 'A' && r <= 'Z' {
			return string(r), true
		}
		if r >= '0' && r <= '9' {
			return string(r), true
		}
		if unicode.IsLetter(r) {
			return "", false
		}
	}
	return "", false
}
