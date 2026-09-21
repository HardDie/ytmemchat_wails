package main

import (
	"log/slog"

	"github.com/HardDie/ytmemchat_wails/bindings/home"
)

func (a *App) syncInterruptHotkey() {
	a.mu.Lock()
	s := a.settings.InterruptHotkey
	a.mu.Unlock()
	a.hk.Sync(s.IsEnabled(), s.Chord, func() {
		if err := home.New(a).InterruptTTS(); err != nil {
			slog.Debug("interrupt hotkey", "err", err)
		}
	})
	a.mu.Lock()
	a.hotkeyErr = a.hk.Err()
	a.mu.Unlock()
}

func (a *App) stopInterruptHotkey() {
	a.hk.Stop()
	a.mu.Lock()
	a.hotkeyErr = ""
	a.mu.Unlock()
}
