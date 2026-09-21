package test

import (
	"errors"
	"testing"

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
