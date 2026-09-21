// Package test is the Wails binding for the Test pane.
package test

import (
	"fmt"
	"strings"

	"github.com/HardDie/ytmemchat_wails/internal/obs"
)

const testAuthor = "test"

// ErrOBSNotListening is returned when overlay HTTP is down.
var ErrOBSNotListening = fmt.Errorf("OBS HTTP is not listening")

// ErrMessageEmpty is returned when the test payload is blank.
var ErrMessageEmpty = fmt.Errorf("test message is empty")

// Test exposes injecting a fake chat line into alerts and TTS.
type Test struct {
	app api
}

// api is the desktop App core this pane uses.
type api interface {
	OverlayServer() *obs.Server
	DispatchChat(fallback *obs.Server, author, text string)
}

// New wraps the application core for the Test pane.
func New(app api) *Test {
	return &Test{app: app}
}

// SendTestMessage fans a fake chat line through the same alerts/TTS path as
// live chat and POST /api/webhook. The HTTP webhook toggle is not required.
func (t *Test) SendTestMessage(text string) error {
	text = strings.TrimSpace(text)
	if text == "" {
		return ErrMessageEmpty
	}
	srv := t.app.OverlayServer()
	if srv == nil {
		return ErrOBSNotListening
	}
	t.app.DispatchChat(srv, testAuthor, text)
	return nil
}
