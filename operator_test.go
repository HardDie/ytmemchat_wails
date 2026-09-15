package main

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/HardDie/ytmemchat_wails/internal/config"
	"github.com/HardDie/ytmemchat_wails/internal/obs"
	"github.com/gorilla/websocket"
)

func TestSendTestMessage_empty(t *testing.T) {
	a := newAppWithStore(config.NewStore(t.TempDir() + "/c.json"))
	if err := a.SendTestMessage("  "); !errors.Is(err, errTestMessageEmpty) {
		t.Fatalf("err = %v", err)
	}
}

func TestSendTestMessage_requiresHTTP(t *testing.T) {
	a := newAppWithStore(config.NewStore(t.TempDir() + "/c.json"))
	if err := a.SendTestMessage("hi"); !errors.Is(err, errOBSNotListening) {
		t.Fatalf("err = %v", err)
	}
	if err := a.InterruptTTS(); !errors.Is(err, errOBSNotListening) {
		t.Fatalf("interrupt %v", err)
	}
	if err := a.PreviewAlert("a.mp3", 1, 1); !errors.Is(err, errOBSNotListening) {
		t.Fatalf("preview %v", err)
	}
}

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
	if err := a.SendTestMessage(" hello "); err != nil {
		t.Fatal(err)
	}
	_ = chat.SetReadDeadline(time.Now().Add(2 * time.Second))
	var ce obs.ChatEvent
	if err := chat.ReadJSON(&ce); err != nil {
		t.Fatal(err)
	}
	if ce.AuthorName != testAuthor || ce.MessageText != "hello" {
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

func TestInterruptTTS_publishesOverlay(t *testing.T) {
	a := listenApp(t)
	u := "ws" + strings.TrimPrefix(a.GetOBSStatus().OverlayURL, "http") + "/ws"
	conn, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	time.Sleep(30 * time.Millisecond)
	if err := a.InterruptTTS(); err != nil {
		t.Fatal(err)
	}
	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	var ev obs.OverlayEvent
	if err := conn.ReadJSON(&ev); err != nil {
		t.Fatal(err)
	}
	if ev.Type != obs.PayloadTypeTTSInterrupt {
		t.Fatalf("%+v", ev)
	}
}

func TestPreviewAlert_publishesOverlay(t *testing.T) {
	a := listenApp(t)
	if err := a.PreviewAlert("  ", 1, 1); !errors.Is(err, errAlertFileEmpty) {
		t.Fatalf("empty %v", err)
	}
	u := "ws" + strings.TrimPrefix(a.GetOBSStatus().OverlayURL, "http") + "/ws"
	conn, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	time.Sleep(30 * time.Millisecond)
	if err := a.PreviewAlert(" videos/huh.webm ", 0.5, 1.2); err != nil {
		t.Fatal(err)
	}
	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	var ev obs.OverlayEvent
	if err := conn.ReadJSON(&ev); err != nil {
		t.Fatal(err)
	}
	if ev.Type != obs.PayloadTypeAlert || ev.Filename != "videos/huh.webm" || ev.Volume != 0.5 || ev.Scale != 1.2 {
		t.Fatalf("%+v", ev)
	}
}
