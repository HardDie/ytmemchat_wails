package home

import (
	"context"
	"errors"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/HardDie/ytmemchat_wails/internal/obs"
	"github.com/gorilla/websocket"
)

func TestInterruptTTS_requiresHTTP(t *testing.T) {
	if err := New(&homeStub{}).InterruptTTS(); !errors.Is(err, ErrOBSNotListening) {
		t.Fatalf("err = %v", err)
	}
}

func TestInterruptTTS_publishesOverlay(t *testing.T) {
	srv, addr := listenOBS(t)
	u := "ws" + strings.TrimPrefix(obs.OverlaySourceURL(addr), "http") + "/ws" + obs.SocketQuery("")
	conn, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	time.Sleep(30 * time.Millisecond)
	if err := New(&homeStub{overlay: srv}).InterruptTTS(); err != nil {
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

func listenOBS(t *testing.T) (*obs.Server, string) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	srv := obs.New(obs.Config{Addr: ln.Addr().String()})
	go func() { _ = srv.Serve(ln) }()
	t.Cleanup(func() {
		_ = srv.Shutdown(context.Background())
	})
	return srv, ln.Addr().String()
}
