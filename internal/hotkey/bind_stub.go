//go:build nomain || integration

package hotkey

func newSession(_ Chord, _ func()) (stoppable, error) {
	return noopSession{}, nil
}

type noopSession struct{}

func (noopSession) unbind() {}
