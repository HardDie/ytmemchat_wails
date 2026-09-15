package main

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/HardDie/ytmemchat_wails/internal/config"
	"github.com/HardDie/ytmemchat_wails/internal/obs"
	"github.com/gorilla/websocket"
)

func listenApp(t *testing.T) *App {
	t.Helper()
	store := config.NewStore(t.TempDir() + "/c.json")
	a := newAppWithStore(store)
	a.listenOverride = "127.0.0.1:0"
	a.skipHTTP = false
	a.mu.Lock()
	if err := a.startHTTPLocked(); err != nil {
		a.mu.Unlock()
		t.Fatal(err)
	}
	a.mu.Unlock()
	t.Cleanup(func() {
		a.Stop()
		a.mu.Lock()
		a.stopHTTPLocked()
		a.mu.Unlock()
	})
	return a
}

func TestInject_alertWithoutStart(t *testing.T) {
	yaml := filepath.Join(t.TempDir(), "commands.yaml")
	if err := os.WriteFile(yaml, []byte("commands:\n  - name: jump\n    file: jump.mp3\n    volume: 0.5\n    scale: 1.2\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	a := listenApp(t)
	spoke := 0
	a.newSynth = func(config.Settings, *obs.Server) (synthesizer, error) {
		return synthFunc(func(string) error {
			spoke++
			return nil
		}), nil
	}
	savePatched(t, a, func(f *SettingsForm) {
		f.WebhookEnabled = true
		f.AlertsEnabled = true
		f.AlertsToken = "@"
		f.AlertsCommandsFilePath = yaml
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

	base := strings.TrimSuffix(obsSt.ChatURL, obs.PathChat)
	res, err := http.Post(base+obs.PathWebhook, "application/json", strings.NewReader(`{"message":"go @jump now"}`))
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("status %d", res.StatusCode)
	}

	_ = chat.SetReadDeadline(time.Now().Add(2 * time.Second))
	var ce obs.ChatEvent
	if err := chat.ReadJSON(&ce); err != nil {
		t.Fatal(err)
	}
	if ce.AuthorName != "webhook" || ce.MessageText != "go @jump now" {
		t.Fatalf("chat %+v", ce)
	}
	_ = over.SetReadDeadline(time.Now().Add(2 * time.Second))
	var ev obs.OverlayEvent
	if err := over.ReadJSON(&ev); err != nil {
		t.Fatal(err)
	}
	if ev.Type != obs.PayloadTypeAlert || ev.Filename != "jump.mp3" {
		t.Fatalf("overlay %+v", ev)
	}
	if spoke != 0 {
		t.Fatalf("tts called %d", spoke)
	}
	if a.GetRunStatus().Running {
		t.Fatal("youtube should not be running")
	}
}

func TestInject_plainTextTTS(t *testing.T) {
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
		f.WebhookEnabled = true
		f.AlertsEnabled = false
		f.TTSEnabled = true
	})
	obsSt := a.GetOBSStatus()
	overURL := "ws" + strings.TrimPrefix(obsSt.OverlayURL, "http") + "/ws"
	over, _, err := websocket.DefaultDialer.Dial(overURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer over.Close()
	time.Sleep(30 * time.Millisecond)

	base := strings.TrimSuffix(obsSt.ChatURL, obs.PathChat)
	res, err := http.Post(base+obs.PathWebhook, "application/json", strings.NewReader(`{"message":"hello"}`))
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("status %d", res.StatusCode)
	}
	_ = over.SetReadDeadline(time.Now().Add(2 * time.Second))
	var ev obs.OverlayEvent
	if err := over.ReadJSON(&ev); err != nil {
		t.Fatal(err)
	}
	if ev.Type != obs.PayloadTypeTTS || string(ev.Payload) != "wav" {
		t.Fatalf("%+v", ev)
	}
	select {
	case text := <-gotText:
		if text != "hello" {
			t.Fatalf("text %q", text)
		}
	case <-time.After(time.Second):
		t.Fatal("no tts text")
	}
}
