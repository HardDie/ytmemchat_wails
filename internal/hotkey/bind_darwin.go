//go:build !nomain && !integration && darwin

package hotkey

import oshotkey "golang.design/x/hotkey"

func osHotkeyExtraMods(c Chord) []oshotkey.Modifier {
	var m []oshotkey.Modifier
	if c.Alt {
		m = append(m, oshotkey.ModOption)
	}
	if c.Super {
		m = append(m, oshotkey.ModCmd)
	}
	return m
}
