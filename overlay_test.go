package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/HardDie/ytmemchat_wails/internal/alerts"
	"github.com/HardDie/ytmemchat_wails/internal/config"
	"github.com/HardDie/ytmemchat_wails/internal/obs"
	"github.com/HardDie/ytmemchat_wails/internal/youtube"
	"github.com/gorilla/websocket"
)

type recSink struct {
	chat []obs.ChatEvent
	over []obs.OverlayEvent
}

func (r *recSink) PublishChat(ev obs.ChatEvent) { r.chat = append(r.chat, ev) }
func (r *recSink) PublishOverlay(ev obs.OverlayEvent) {
	r.over = append(r.over, ev)
}

type matchStub struct {
	hit bool
	n   int
}

func (m *matchStub) Alert(string) bool {
	m.n++
	return m.hit
}

type synthStub struct {
	n   int
	err error
}

func (s *synthStub) SynthesizeAudio(string) error {
	s.n++
	return s.err
}

func overlayOff(a *App) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.settings.Alerts.Enabled = false
	a.settings.TTS.Enabled = false
}

func TestDispatchChat_alertSkipsTTS(t *testing.T) {
	sink := &recSink{}
	match := &matchStub{hit: true}
	speak := &synthStub{}
	dispatchChat(sink, match, speak, &youtube.ChatMessage{Author: "A", Message: "@jump"})
	if len(sink.chat) != 1 || speak.n != 0 || match.n != 1 {
		t.Fatalf("chat %d speak %d match %d", len(sink.chat), speak.n, match.n)
	}
}

func TestDispatchChat_plainTextTTS(t *testing.T) {
	sink := &recSink{}
	match := &matchStub{hit: false}
	speak := &synthStub{}
	dispatchChat(sink, match, speak, &youtube.ChatMessage{Author: "A", Message: "hi"})
	if speak.n != 1 {
		t.Fatalf("speak %d", speak.n)
	}
}

func TestDispatchChat_alertsOffTTS(t *testing.T) {
	sink := &recSink{}
	speak := &synthStub{}
	dispatchChat(sink, nil, speak, &youtube.ChatMessage{Message: "hi"})
	if speak.n != 1 {
		t.Fatalf("speak %d", speak.n)
	}
}

func TestDispatchChat_bothOff(t *testing.T) {
	sink := &recSink{}
	dispatchChat(sink, nil, nil, &youtube.ChatMessage{Message: "hi"})
	if len(sink.chat) != 1 || len(sink.over) != 0 {
		t.Fatalf("%+v", sink)
	}
}

func TestDispatchChat_nilMessage(t *testing.T) {
	sink := &recSink{}
	dispatchChat(sink, &matchStub{hit: true}, &synthStub{}, nil)
	if len(sink.chat) != 0 {
		t.Fatal("expected skip")
	}
}

func TestStart_missingDefaultCommandsFile(t *testing.T) {
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
	media := t.TempDir()
	savePatched(t, a, func(f *SettingsForm) {
		f.StreamID = "vid"
		f.Port = "8080"
		f.AlertsEnabled = true
		f.AlertsToken = "@"
		f.AlertsMediaPath = media
		f.AlertsCommandsFileCustom = false
	})
	a.newSynth = func(config.Settings, *obs.Server) (synthesizer, error) {
		return synthFunc(func(string) error { return nil }), nil
	}
	ch := make(chan *youtube.ChatMessage, 1)
	a.clientFn = func(config.Settings) (youtube.Client, error) {
		return &fakeClient{it: &fakeIterator{ch: ch}}, nil
	}
	if err := a.Start(); err != nil {
		t.Fatal(err)
	}
	a.Stop()
}

func TestStart_missingCommandsFile(t *testing.T) {
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
	savePatched(t, a, func(f *SettingsForm) {
		f.StreamID = "vid"
		f.Port = "8080"
	})
	a.mu.Lock()
	a.settings.Alerts.Enabled = true
	a.settings.Alerts.CommandsFilePath = filepath.Join(t.TempDir(), "nope.yaml")
	a.settings.TTS.Enabled = false
	a.mu.Unlock()
	err := a.Start()
	if err == nil || !strings.Contains(err.Error(), "alerts commands file") {
		t.Fatalf("err = %v", err)
	}
	st := a.GetRunStatus()
	if st.Running || st.Connecting || st.Error == "" {
		t.Fatalf("%+v", st)
	}
}

func TestStart_alertPublishesOverlayNotTTS(t *testing.T) {
	yaml := filepath.Join(t.TempDir(), "commands.yaml")
	if err := os.WriteFile(yaml, []byte("commands:\n  - name: jump\n    file: jump.mp3\n    volume: 0.5\n    scale: 1.2\n"), 0o600); err != nil {
		t.Fatal(err)
	}
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
	savePatched(t, a, func(f *SettingsForm) {
		f.StreamID = "live1"
		f.Port = "8080"
	})
	spoke := 0
	a.mu.Lock()
	a.settings.Alerts.Enabled = true
	a.settings.Alerts.Token = "@"
	a.settings.Alerts.CommandsFilePath = yaml
	a.settings.TTS.Enabled = true
	a.mu.Unlock()
	a.newSynth = func(config.Settings, *obs.Server) (synthesizer, error) {
		return synthFunc(func(string) error {
			spoke++
			return nil
		}), nil
	}
	ch := make(chan *youtube.ChatMessage, 1)
	a.clientFn = func(config.Settings) (youtube.Client, error) {
		return &fakeClient{it: &fakeIterator{ch: ch}}, nil
	}
	u := "ws" + strings.TrimPrefix(a.GetOBSStatus().OverlayURL, "http") + "/ws" + obs.SocketQuery("")
	conn, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	time.Sleep(30 * time.Millisecond)
	if err := a.Start(); err != nil {
		t.Fatal(err)
	}
	waitRun(t, a, func(s RunStatus) bool { return s.Running && !s.Connecting })
	ch <- &youtube.ChatMessage{Author: "Ada", Message: "go @jump now"}
	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	var ev obs.OverlayEvent
	if err := conn.ReadJSON(&ev); err != nil {
		t.Fatal(err)
	}
	if ev.Type != obs.PayloadTypeAlert || ev.Filename != "jump.mp3" || ev.Volume != 0.5 || ev.Scale != 1.2 {
		t.Fatalf("%+v", ev)
	}
	if spoke != 0 {
		t.Fatalf("tts called %d", spoke)
	}
}

func TestStart_plainChatUsesTTS(t *testing.T) {
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
	savePatched(t, a, func(f *SettingsForm) {
		f.StreamID = "live1"
		f.Port = "8080"
	})
	a.mu.Lock()
	a.settings.Alerts.Enabled = false
	a.settings.TTS.Enabled = true
	a.mu.Unlock()
	gotText := make(chan string, 1)
	a.newSynth = func(_ config.Settings, srv *obs.Server) (synthesizer, error) {
		return synthFunc(func(text string) error {
			srv.PublishOverlay(obs.TTSOverlay([]byte("wav"), 1))
			gotText <- text
			return nil
		}), nil
	}
	ch := make(chan *youtube.ChatMessage, 1)
	a.clientFn = func(config.Settings) (youtube.Client, error) {
		return &fakeClient{it: &fakeIterator{ch: ch}}, nil
	}
	u := "ws" + strings.TrimPrefix(a.GetOBSStatus().OverlayURL, "http") + "/ws" + obs.SocketQuery("")
	conn, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	time.Sleep(30 * time.Millisecond)
	if err := a.Start(); err != nil {
		t.Fatal(err)
	}
	waitRun(t, a, func(s RunStatus) bool { return s.Running && !s.Connecting })
	ch <- &youtube.ChatMessage{Author: "Ada", Message: "hello"}
	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	var ev obs.OverlayEvent
	if err := conn.ReadJSON(&ev); err != nil {
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

type synthFunc func(string) error

func (f synthFunc) SynthesizeAudio(text string) error { return f(text) }

func TestTTSDebugMessage(t *testing.T) {
	missing := fmt.Errorf("synthesis failed: %w", &exec.Error{Name: "say", Err: exec.ErrNotFound})
	if got := ttsDebugMessage(missing); got != "debug: app for tts not found" {
		t.Fatalf("missing app %q", got)
	}
	if got := ttsDebugMessage(errors.New("tts: no speakable text")); got != "debug: error text to message" {
		t.Fatalf("other %q", got)
	}
}

func TestAlertPublisher_debugLogs(t *testing.T) {
	media := t.TempDir()
	if err := os.WriteFile(filepath.Join(media, "jump.mp3"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	yamlPath := filepath.Join(t.TempDir(), "commands.yaml")
	body := "commands:\n  - name: jump\n    file: jump.mp3\n  - name: gone\n    file: missing.mp3\n"
	if err := os.WriteFile(yamlPath, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	out := make(chan alerts.Clip, 1)
	inner, err := alerts.New(alerts.Config{
		Token:            "@",
		MediaPath:        media,
		CommandsFilePath: yamlPath,
		Out:              out,
	})
	if err != nil {
		t.Fatal(err)
	}
	sink := &recSink{}
	p := &alertPublisher{inner: inner, out: out, srv: sink, debug: true}
	var lines []string
	prev := debugLog
	debugLog = func(msg string, _ ...any) {
		lines = append(lines, msg)
	}
	t.Cleanup(func() { debugLog = prev })

	if !p.Alert("please @jump") {
		t.Fatal("expected jump")
	}
	if !p.Alert("@gone") {
		t.Fatal("expected gone")
	}
	if p.Alert("hello") {
		t.Fatal("expected no command")
	}
	want := []string{
		"debug: started parsing command",
		"debug: found command",
		"debug: success",
		"debug: started parsing command",
		"debug: found command",
		"debug: error path to the file",
		"debug: started parsing command",
		"debug: command not found",
	}
	if strings.Join(lines, "\n") != strings.Join(want, "\n") {
		t.Fatalf("logs:\n%s", strings.Join(lines, "\n"))
	}
	if len(sink.over) != 2 || sink.over[0].Filename != "jump.mp3" || sink.over[1].Filename != "missing.mp3" {
		t.Fatalf("overlay %+v", sink.over)
	}
}
