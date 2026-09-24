package commands

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

func TestPreviewAlert_requiresHTTP(t *testing.T) {
	if err := New(&cmdStub{}).PreviewAlert("a.mp3", 1, 1); !errors.Is(err, ErrOBSNotListening) {
		t.Fatalf("err = %v", err)
	}
}

func TestPreviewAlert_publishesOverlay(t *testing.T) {
	c := New(&cmdStub{})
	if err := c.PreviewAlert("  ", 1, 1); !errors.Is(err, ErrAlertFileEmpty) {
		t.Fatalf("empty %v", err)
	}
	srv, addr := listenOBS(t)
	c = New(&cmdStub{overlay: srv})
	u := "ws" + strings.TrimPrefix(obs.OverlaySourceURL(addr), "http") + "/ws" + obs.SocketQuery("")
	conn, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	time.Sleep(30 * time.Millisecond)
	if err := c.PreviewAlert(" videos/huh.webm ", 0.5, 1.2); err != nil {
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
