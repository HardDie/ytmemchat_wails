package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/HardDie/ytmemchat_wails/internal/obs"
	"github.com/HardDie/ytmemchat_wails/internal/youtube"
)

const testAuthor = "test"

var (
	errOBSNotListening  = fmt.Errorf("OBS HTTP is not listening")
	errTestMessageEmpty = fmt.Errorf("test message is empty")
)

// SendTestMessage fans a fake chat line through the same alerts/TTS path as
// live chat and POST /api/webhook. The HTTP webhook toggle is not required.
func (a *App) SendTestMessage(text string) error {
	text = strings.TrimSpace(text)
	if text == "" {
		return errTestMessageEmpty
	}
	a.mu.Lock()
	srv := a.httpSrv
	a.mu.Unlock()
	if srv == nil {
		return errOBSNotListening
	}
	a.dispatchLine(srv, &youtube.ChatMessage{
		Author:    testAuthor,
		Message:   text,
		Timestamp: time.Now(),
	})
	return nil
}

// InterruptTTS publishes overlay type tts_interrupt. The HTTP interrupt route
// is not required; this works whenever OBS HTTP is listening.
func (a *App) InterruptTTS() error {
	a.mu.Lock()
	srv := a.httpSrv
	a.mu.Unlock()
	if srv == nil {
		return errOBSNotListening
	}
	srv.PublishOverlay(obs.InterruptOverlay())
	return nil
}
