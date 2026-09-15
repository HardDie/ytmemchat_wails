package main

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/HardDie/ytmemchat_wails/internal/config"
	"github.com/HardDie/ytmemchat_wails/internal/obs"
	"github.com/HardDie/ytmemchat_wails/internal/youtube"
	"github.com/gorilla/websocket"
)

type fakeIterator struct {
	ch chan *youtube.ChatMessage
}

func (f *fakeIterator) Next() (*youtube.ChatMessage, bool) {
	m, ok := <-f.ch
	return m, ok
}

func (f *fakeIterator) GetChan() <-chan *youtube.ChatMessage {
	return f.ch
}

type fakeClient struct {
	it      youtube.MessageIterator
	err     error
	videoID string
	calls   int
}

func (f *fakeClient) GetMessageIterator(ctx context.Context, liveVideoID string) (youtube.MessageIterator, error) {
	f.calls++
	f.videoID = liveVideoID
	if f.err != nil {
		return nil, f.err
	}
	return f.it, nil
}

func waitRun(t *testing.T, a *App, pred func(RunStatus) bool) RunStatus {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	var st RunStatus
	for time.Now().Before(deadline) {
		st = a.GetRunStatus()
		if pred(st) {
			return st
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timeout status %+v", st)
	return st
}

func TestStart_requiresStreamID(t *testing.T) {
	a := newAppWithStore(config.NewStore(t.TempDir() + "/c.json"))
	err := a.Start()
	if !errors.Is(err, config.ErrStreamIDRequired) {
		t.Fatalf("err = %v", err)
	}
	st := a.GetRunStatus()
	if st.Running || st.Error == "" {
		t.Fatalf("%+v", st)
	}
}

func TestStart_requiresHTTP(t *testing.T) {
	st := config.NewStore(t.TempDir() + "/c.json")
	a := newAppWithStore(st)
	if err := a.SaveSettings(SettingsForm{StreamID: "vid", Port: "8080"}); err != nil {
		t.Fatal(err)
	}
	err := a.Start()
	if err == nil || !strings.Contains(err.Error(), "HTTP") {
		t.Fatalf("err = %v", err)
	}
}

func TestStart_publishesChatAndStop(t *testing.T) {
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
	if err := a.SaveSettings(SettingsForm{StreamID: "live1", Port: "8080"}); err != nil {
		t.Fatal(err)
	}
	ch := make(chan *youtube.ChatMessage, 1)
	fc := &fakeClient{it: &fakeIterator{ch: ch}}
	a.clientFn = func(s config.Settings) (youtube.Client, error) {
		if s.HasAPIKey() {
			t.Fatal("expected nokey path")
		}
		return fc, nil
	}
	u := "ws" + strings.TrimPrefix(a.GetOBSStatus().ChatURL, "http") + "/ws"
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
	if fc.videoID != "live1" {
		t.Fatalf("video %q", fc.videoID)
	}
	ch <- &youtube.ChatMessage{Author: "Ada", ImgURL: "http://pic", Message: "hi", Timestamp: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)}
	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	var ev obs.ChatEvent
	if err := conn.ReadJSON(&ev); err != nil {
		t.Fatal(err)
	}
	if ev.AuthorName != "Ada" || ev.MessageText != "hi" {
		t.Fatalf("%+v", ev)
	}
	a.Stop()
	st := a.GetRunStatus()
	if st.Running || st.Connecting {
		t.Fatalf("after stop %+v", st)
	}
}

func TestStart_invalidAPIKeyNoNokey(t *testing.T) {
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
	if err := a.SaveSettings(SettingsForm{StreamID: "vid", APIKey: "bad", Port: "8080"}); err != nil {
		t.Fatal(err)
	}
	calls := 0
	a.clientFn = func(s config.Settings) (youtube.Client, error) {
		calls++
		if !s.HasAPIKey() {
			t.Fatal("fell back to nokey")
		}
		return &fakeClient{err: youtube.ErrInvalidAPIKey}, nil
	}
	if err := a.Start(); err != nil {
		t.Fatal(err)
	}
	st := waitRun(t, a, func(s RunStatus) bool { return s.Error != "" })
	if !strings.Contains(st.Error, "invalid") {
		t.Fatalf("%+v", st)
	}
	if st.Running || calls != 1 {
		t.Fatalf("calls %d status %+v", calls, st)
	}
}

func TestChooseYouTubeClient_emptyKey(t *testing.T) {
	c, err := chooseYouTubeClient(config.Defaults())
	if err != nil || c == nil {
		t.Fatalf("%v %v", c, err)
	}
}
