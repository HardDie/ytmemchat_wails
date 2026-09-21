//go:build !nomain && !integration && linux

package hotkey

import oshotkey "golang.design/x/hotkey"

func osHotkeyExtraMods(c Chord) []oshotkey.Modifier {
	var m []oshotkey.Modifier
	if c.Alt {
		m = append(m, oshotkey.Mod1)
	}
	if c.Super {
		m = append(m, oshotkey.Mod4)
	}
	return m
}
