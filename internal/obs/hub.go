package obs

import (
	"log/slog"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// upgrader keeps gorilla's default origin check.
// A socket opened by a page on another site is refused.
// OBS passes: its page is served by this server, so Origin matches Host.
var upgrader = websocket.Upgrader{}

type hub[T any] struct {
	mu      sync.Mutex
	clients map[*websocket.Conn]struct{}
	// send is never closed: many goroutines publish, so none may close it.
	send chan T
	// done is closed by shutdown. publish, loop, and serveWS stop on it.
	done      chan struct{}
	log       *slog.Logger
	closeOnce sync.Once
}

func newHub[T any](log *slog.Logger) *hub[T] {
	h := &hub[T]{
		clients: make(map[*websocket.Conn]struct{}),
		send:    make(chan T, 64),
		done:    make(chan struct{}),
		log:     log,
	}
	go h.loop()
	return h
}

func (h *hub[T]) loop() {
	for {
		select {
		case msg := <-h.send:
			h.broadcast(msg)
		case <-h.done:
			return
		}
	}
}

// publish queues msg for every client. After shutdown it drops msg.
// A caller may still hold a server that a settings save just replaced.
func (h *hub[T]) publish(msg T) {
	select {
	case <-h.done:
		return
	default:
	}
	select {
	case h.send <- msg:
	case <-h.done:
	}
}

func (h *hub[T]) broadcast(msg T) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for c := range h.clients {
		if err := c.WriteJSON(msg); err != nil {
			h.log.Debug("obs websocket write", "err", err)
			c.Close()
			delete(h.clients, c)
		}
	}
}

func (h *hub[T]) serveWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.log.Debug("obs websocket upgrade", "err", err)
		return
	}
	defer conn.Close()

	h.mu.Lock()
	select {
	case <-h.done:
		// shutdown started; closeAll may have run already and would miss this socket.
		h.mu.Unlock()
		return
	default:
	}
	h.clients[conn] = struct{}{}
	h.mu.Unlock()
	defer func() {
		h.mu.Lock()
		delete(h.clients, conn)
		h.mu.Unlock()
	}()

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}
	}
}

func (h *hub[T]) shutdown() {
	h.closeOnce.Do(func() {
		close(h.done)
		h.closeAll()
	})
}

func (h *hub[T]) closeAll() {
	h.mu.Lock()
	defer h.mu.Unlock()
	for c := range h.clients {
		c.Close()
		delete(h.clients, c)
	}
}
