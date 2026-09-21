package hotkey

import (
	"log/slog"
	"sync"
)

// Binding is one OS-global hotkey. The zero value is idle.
type Binding struct {
	mu      sync.Mutex
	err     string
	session stoppable
}

type stoppable interface {
	unbind()
}

// Sync unregisters any previous chord, then registers chord when enabled.
// onPress runs on key down (debounced). Parse errors and OS register errors
// are stored on [Binding.Err].
func (b *Binding) Sync(enabled bool, chord string, onPress func()) {
	b.Stop()
	if !enabled {
		return
	}
	parsed, err := Parse(chord)
	if err != nil {
		b.setErr(err.Error())
		slog.Error("interrupt hotkey", "err", err)
		return
	}
	s, err := newSession(parsed, onPress)
	if err != nil {
		b.setErr(err.Error())
		slog.Error("interrupt hotkey register", "err", err, "chord", parsed.String())
		return
	}
	b.mu.Lock()
	b.session = s
	b.mu.Unlock()
}

// Stop unregisters the OS hotkey if one is active.
func (b *Binding) Stop() {
	b.mu.Lock()
	s := b.session
	b.session = nil
	b.err = ""
	b.mu.Unlock()
	if s != nil {
		s.unbind()
	}
}

// Err is the last Sync failure, or empty after a successful Sync or Stop.
func (b *Binding) Err() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.err
}

func (b *Binding) setErr(msg string) {
	b.mu.Lock()
	b.err = msg
	b.mu.Unlock()
}
