//go:build !nomain && !integration

package hotkey

import (
	"fmt"
	"time"

	oshotkey "golang.design/x/hotkey"
)

type osSession struct {
	hk   *oshotkey.Hotkey
	stop chan struct{}
	done chan struct{}
}

func newSession(c Chord, onPress func()) (stoppable, error) {
	key, ok := osHotkeyKey(c.Key)
	if !ok {
		return nil, fmt.Errorf("unsupported key %s", c.Key)
	}
	hk := oshotkey.New(osHotkeyMods(c), key)
	if err := hk.Register(); err != nil {
		return nil, err
	}
	stop := make(chan struct{})
	done := make(chan struct{})
	go listenOSHotkey(hk, onPress, stop, done)
	return &osSession{hk: hk, stop: stop, done: done}, nil
}

func (s *osSession) unbind() {
	close(s.stop)
	<-s.done
	_ = s.hk.Unregister()
}

func listenOSHotkey(hk *oshotkey.Hotkey, onPress func(), stop <-chan struct{}, done chan struct{}) {
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
			if onPress != nil {
				onPress()
			}
		}
	}
}

func osHotkeyMods(c Chord) []oshotkey.Modifier {
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
