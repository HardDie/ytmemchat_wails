package main

import (
	"strings"
	"testing"
	"time"

	testpane "github.com/HardDie/ytmemchat_wails/bindings/test"
	"github.com/HardDie/ytmemchat_wails/internal/config"
	"github.com/HardDie/ytmemchat_wails/internal/obs"
	"github.com/gorilla/websocket"
)

func TestSendTestMessage_usesOverlayPath(t *testing.T) {
	a := listenApp(t)
	gotText := make(chan string, 1)
	a.newSynth = func(_ config.Settings, srv *obs.Server) (synthesizer, error) {
		return synthFunc(func(text string) error {
			srv.PublishOverlay(obs.TTSOverlay([]byte("wav"), 1))
			gotText <- text
			return nil
		}), nil
	}
	savePatched(t, a, func(f *SettingsForm) {
		f.WebhookEnabled = false
		f.AlertsEnabled = false
		f.TTSEnabled = true
	})
	obsSt := a.GetOBSStatus()
	chatURL := "ws" + strings.TrimPrefix(obsSt.ChatURL, "http") + "/ws"
	overURL := "ws" + strings.TrimPrefix(obsSt.OverlayURL, "http") + "/ws"
	chat, _, err := websocket.DefaultDialer.Dial(chatURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer chat.Close()
	over, _, err := websocket.DefaultDialer.Dial(overURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer over.Close()
	time.Sleep(30 * time.Millisecond)
	if err := testpane.New(a).SendTestMessage(" hello "); err != nil {
		t.Fatal(err)
	}
	_ = chat.SetReadDeadline(time.Now().Add(2 * time.Second))
	var ce obs.ChatEvent
	if err := chat.ReadJSON(&ce); err != nil {
		t.Fatal(err)
	}
	if ce.AuthorName != "test" || ce.MessageText != "hello" {
		t.Fatalf("chat %+v", ce)
	}
	_ = over.SetReadDeadline(time.Now().Add(2 * time.Second))
	var ev obs.OverlayEvent
	if err := over.ReadJSON(&ev); err != nil {
		t.Fatal(err)
	}
	if ev.Type != obs.PayloadTypeTTS {
		t.Fatalf("overlay %+v", ev)
	}
	select {
	case text := <-gotText:
		if text != "hello" {
			t.Fatalf("text %q", text)
		}
	case <-time.After(time.Second):
		t.Fatal("no tts")
	}
}
