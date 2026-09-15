package obs

import (
	"log/slog"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(*http.Request) bool { return true },
}

type hub[T any] struct {
	mu        sync.Mutex
	clients   map[*websocket.Conn]struct{}
	send      chan T
	log       *slog.Logger
	closeOnce sync.Once
}

func newHub[T any](log *slog.Logger) *hub[T] {
	h := &hub[T]{
		clients: make(map[*websocket.Conn]struct{}),
		send:    make(chan T, 64),
		log:     log,
	}
	go h.loop()
	return h
}

func (h *hub[T]) loop() {
	for msg := range h.send {
		h.broadcast(msg)
	}
}

func (h *hub[T]) publish(msg T) {
	h.send <- msg
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
		h.closeAll()
		close(h.send)
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
