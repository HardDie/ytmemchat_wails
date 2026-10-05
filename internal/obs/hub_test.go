package obs

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// A settings save can shut a server down while a TTS line is still in flight.
// Publishing to it then used to panic with "send on closed channel".
func TestPublishAfterShutdown_drops(t *testing.T) {
	s := New(Config{})
	if err := s.Shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}
	s.PublishChat(NewChatEvent("a", "", "hi", time.Time{}))
	s.PublishOverlay(InterruptOverlay())
}

func TestShutdownTwice(t *testing.T) {
	s := New(Config{})
	_ = s.Shutdown(context.Background())
	_ = s.Shutdown(context.Background())
}

func TestPublishDuringShutdown_race(t *testing.T) {
	for range 50 {
		s := New(Config{})
		var wg sync.WaitGroup
		for range 8 {
			wg.Go(func() {
				for range 200 {
					s.PublishOverlay(InterruptOverlay())
					s.PublishChat(FlushChat())
				}
			})
		}
		_ = s.Shutdown(context.Background())
		wg.Wait()
	}
}

// A socket that registers after shutdown must not stay open.
func TestServeWSAfterShutdown_closes(t *testing.T) {
	h := newHub[ChatEvent](slog.Default())
	ts := httptest.NewServer(http.HandlerFunc(h.serveWS))
	defer ts.Close()
	h.shutdown()

	u := "ws" + strings.TrimPrefix(ts.URL, "http")
	c, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	_ = c.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, _, err = c.ReadMessage()
	if err == nil {
		t.Fatal("want closed socket")
	}
	if ne, ok := err.(interface{ Timeout() bool }); ok && ne.Timeout() {
		t.Fatal("socket stayed open after shutdown")
	}
	h.mu.Lock()
	n := len(h.clients)
	h.mu.Unlock()
	if n != 0 {
		t.Fatalf("clients = %d, want 0", n)
	}
}
