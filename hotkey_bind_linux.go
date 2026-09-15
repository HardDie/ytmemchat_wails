//go:build !nomain && linux

package main

import (
	oshotkey "golang.design/x/hotkey"

	"github.com/HardDie/ytmemchat_wails/internal/hotkey"
)

func osHotkeyExtraMods(c hotkey.Chord) []oshotkey.Modifier {
	var m []oshotkey.Modifier
	if c.Alt {
		m = append(m, oshotkey.Mod1)
	}
	if c.Super {
		m = append(m, oshotkey.Mod4)
	}
	return m
}
