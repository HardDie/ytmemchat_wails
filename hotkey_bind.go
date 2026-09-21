//go:build !nomain

package main

import (
	"log/slog"
	"time"

	oshotkey "golang.design/x/hotkey"

	"github.com/HardDie/ytmemchat_wails/bindings/home"
	"github.com/HardDie/ytmemchat_wails/internal/hotkey"
)

type interruptHotkey struct {
	hk   *oshotkey.Hotkey
	stop chan struct{}
	done chan struct{}
}

func (a *App) syncInterruptHotkey() {
	a.stopInterruptHotkey()
	a.mu.Lock()
	s := a.settings.InterruptHotkey
	a.hotkeyErr = ""
	a.mu.Unlock()
	if !s.IsEnabled() {
		return
	}
	chord, err := hotkey.Parse(s.Chord)
	if err != nil {
		a.setHotkeyErr(err.Error())
		slog.Error("interrupt hotkey", "err", err)
		return
	}
	key, ok := osHotkeyKey(chord.Key)
	if !ok {
		a.setHotkeyErr("unsupported key " + chord.Key)
		slog.Error("interrupt hotkey", "key", chord.Key)
		return
	}
	hk := oshotkey.New(osHotkeyMods(chord), key)
	if err := hk.Register(); err != nil {
		a.setHotkeyErr(err.Error())
		slog.Error("interrupt hotkey register", "err", err, "chord", chord.String())
		return
	}
	stop := make(chan struct{})
	done := make(chan struct{})
	a.mu.Lock()
	still := a.settings.InterruptHotkey.IsEnabled() && a.settings.InterruptHotkey.Chord == chord.String()
	if !still {
		a.mu.Unlock()
		_ = hk.Unregister()
		return
	}
	a.hk = interruptHotkey{hk: hk, stop: stop, done: done}
	a.mu.Unlock()
	go listenInterruptHotkey(a, hk, stop, done)
}

func (a *App) setHotkeyErr(msg string) {
	a.mu.Lock()
	a.hotkeyErr = msg
	a.mu.Unlock()
}

func (a *App) stopInterruptHotkey() {
	a.mu.Lock()
	stop := a.hk.stop
	done := a.hk.done
	hk := a.hk.hk
	a.hk = interruptHotkey{}
	a.hotkeyErr = ""
	a.mu.Unlock()
	if stop != nil {
		close(stop)
	}
	if done != nil {
		<-done
	}
	if hk != nil {
		_ = hk.Unregister()
	}
}

func listenInterruptHotkey(a *App, hk *oshotkey.Hotkey, stop <-chan struct{}, done chan struct{}) {
	defer close(done)
	tick := time.NewTicker(200 * time.Millisecond)
	defer tick.Stop()
	var last time.Time
	for {
		select {
		case <-stop:
			return
		case <-tick.C:
		case <-hk.Keydown():
			select {
			case <-stop:
				return
			default:
			}
			if !last.IsZero() && time.Since(last) < 300*time.Millisecond {
				continue
			}
			last = time.Now()
			if err := home.New(a).InterruptTTS(); err != nil {
				slog.Debug("interrupt hotkey", "err", err)
			}
		}
	}
}

func osHotkeyMods(c hotkey.Chord) []oshotkey.Modifier {
	var m []oshotkey.Modifier
	if c.Ctrl {
		m = append(m, oshotkey.ModCtrl)
	}
	if c.Shift {
		m = append(m, oshotkey.ModShift)
	}
	m = append(m, osHotkeyExtraMods(c)...)
	return m
}

func osHotkeyKey(name string) (oshotkey.Key, bool) {
	k, ok := osHotkeyKeys[name]
	return k, ok
}

var osHotkeyKeys = map[string]oshotkey.Key{
	"Space":  oshotkey.KeySpace,
	"Escape": oshotkey.KeyEscape,
	"Tab":    oshotkey.KeyTab,
	"Enter":  oshotkey.KeyReturn,
	"Delete": oshotkey.KeyDelete,
	"Left":   oshotkey.KeyLeft,
	"Right":  oshotkey.KeyRight,
	"Up":     oshotkey.KeyUp,
	"Down":   oshotkey.KeyDown,
	"A":      oshotkey.KeyA,
	"B":      oshotkey.KeyB,
	"C":      oshotkey.KeyC,
	"D":      oshotkey.KeyD,
	"E":      oshotkey.KeyE,
	"F":      oshotkey.KeyF,
	"G":      oshotkey.KeyG,
	"H":      oshotkey.KeyH,
	"I":      oshotkey.KeyI,
	"J":      oshotkey.KeyJ,
	"K":      oshotkey.KeyK,
	"L":      oshotkey.KeyL,
	"M":      oshotkey.KeyM,
	"N":      oshotkey.KeyN,
	"O":      oshotkey.KeyO,
	"P":      oshotkey.KeyP,
	"Q":      oshotkey.KeyQ,
	"R":      oshotkey.KeyR,
	"S":      oshotkey.KeyS,
	"T":      oshotkey.KeyT,
	"U":      oshotkey.KeyU,
	"V":      oshotkey.KeyV,
	"W":      oshotkey.KeyW,
	"X":      oshotkey.KeyX,
	"Y":      oshotkey.KeyY,
	"Z":      oshotkey.KeyZ,
	"0":      oshotkey.Key0,
	"1":      oshotkey.Key1,
	"2":      oshotkey.Key2,
	"3":      oshotkey.Key3,
	"4":      oshotkey.Key4,
	"5":      oshotkey.Key5,
	"6":      oshotkey.Key6,
	"7":      oshotkey.Key7,
	"8":      oshotkey.Key8,
	"9":      oshotkey.Key9,
	"F1":     oshotkey.KeyF1,
	"F2":     oshotkey.KeyF2,
	"F3":     oshotkey.KeyF3,
	"F4":     oshotkey.KeyF4,
	"F5":     oshotkey.KeyF5,
	"F6":     oshotkey.KeyF6,
	"F7":     oshotkey.KeyF7,
	"F8":     oshotkey.KeyF8,
	"F9":     oshotkey.KeyF9,
	"F10":    oshotkey.KeyF10,
	"F11":    oshotkey.KeyF11,
	"F12":    oshotkey.KeyF12,
}
