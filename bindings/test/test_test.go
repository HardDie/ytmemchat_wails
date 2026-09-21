package test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/HardDie/ytmemchat_wails/internal/obs"
)

type testStub struct {
	overlay *obs.Server
	author  string
	text    string
}

func (s *testStub) OverlayServer() *obs.Server { return s.overlay }

func (s *testStub) DispatchChat(_ *obs.Server, author, text string) {
	s.author = author
	s.text = text
}

func TestSendTestMessage_empty(t *testing.T) {
	if err := New(&testStub{}).SendTestMessage("  "); !errors.Is(err, ErrMessageEmpty) {
		t.Fatalf("err = %v", err)
	}
}

func TestSendTestMessage_requiresHTTP(t *testing.T) {
	if err := New(&testStub{}).SendTestMessage("hi"); !errors.Is(err, ErrOBSNotListening) {
		t.Fatalf("err = %v", err)
	}
}

func TestSendTestMessage_dispatches(t *testing.T) {
	st := &testStub{overlay: &obs.Server{}}
	if err := New(st).SendTestMessage(" hello "); err != nil {
		t.Fatal(err)
	}
	if st.author != testAuthor || st.text != "hello" {
		t.Fatalf("author=%q text=%q", st.author, st.text)
	}
}

func TestFlushChat_requiresHTTP(t *testing.T) {
	if err := New(&testStub{}).FlushChat(); !errors.Is(err, ErrOBSNotListening) {
		t.Fatalf("err = %v", err)
	}
}

func TestFlushChat_publishes(t *testing.T) {
	srv := obs.New(obs.Config{})
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
	})
	if err := New(&testStub{overlay: srv}).FlushChat(); err != nil {
		t.Fatal(err)
	}
}
