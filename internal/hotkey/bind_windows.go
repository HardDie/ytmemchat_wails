//go:build !nomain && !integration && windows

package hotkey

import oshotkey "golang.design/x/hotkey"

func osHotkeyExtraMods(c Chord) []oshotkey.Modifier {
	var m []oshotkey.Modifier
	if c.Alt {
		m = append(m, oshotkey.ModAlt)
	}
	if c.Super {
		m = append(m, oshotkey.ModWin)
	}
	return m
}
