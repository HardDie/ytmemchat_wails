package obs

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"
)

const injectBuffer = 32

// Server is the local HTTP + WebSocket surface for OBS and curl.
type Server struct {
	cfg      Config
	log      *slog.Logger
	mux      *http.ServeMux
	http     *http.Server
	chat     *hub[ChatEvent]
	overlay  *hub[OverlayEvent]
	injected chan InjectedMessage
}

// New registers routes and starts WebSocket broadcasters. It does not listen.
func New(cfg Config) *Server {
	log := slog.Default().With("component", "obs")
	s := &Server{
		cfg:      cfg,
		log:      log,
		mux:      http.NewServeMux(),
		chat:     newHub[ChatEvent](log),
		overlay:  newHub[OverlayEvent](log),
		injected: make(chan InjectedMessage, injectBuffer),
	}
	s.http = &http.Server{
		Addr:              cfg.Addr,
		Handler:           s.mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	s.routes()
	return s
}

// Handler is the HTTP mux (for tests and custom listeners).
func (s *Server) Handler() http.Handler {
	return s.mux
}

// ListenAndServe binds cfg.Addr and serves until [Server.Shutdown] or an error.
func (s *Server) ListenAndServe() error {
	if s.cfg.Addr == "" {
		return fmt.Errorf("obs: empty listen address")
	}
	s.log.Info("http listen", "addr", s.cfg.Addr)
	err := s.http.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("obs: listen: %w", err)
	}
	return nil
}

// Serve serves on an existing listener (tests, port 0).
func (s *Server) Serve(l net.Listener) error {
	s.log.Info("http serve", "addr", l.Addr().String())
	err := s.http.Serve(l)
	if err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("obs: serve: %w", err)
	}
	return nil
}

// Shutdown stops the HTTP server and closes WebSocket clients.
func (s *Server) Shutdown(ctx context.Context) error {
	err := s.http.Shutdown(ctx)
	s.chat.shutdown()
	s.overlay.shutdown()
	return err
}

// PublishChat sends a message to all /obs/chat/ws clients.
func (s *Server) PublishChat(msg ChatEvent) {
	s.chat.publish(msg)
}

// PublishOverlay sends an event to all /obs/overlay/ws clients.
func (s *Server) PublishOverlay(msg OverlayEvent) {
	s.overlay.publish(msg)
}

// NotifyAppClosed writes an app_closed event to chat and overlay sockets
// before the connections are torn down. Callers should wait briefly so the
// browser can handle the message, then [Server.Shutdown].
func (s *Server) NotifyAppClosed() {
	s.chat.broadcast(AppClosedChat())
	s.overlay.broadcast(AppClosedOverlay())
}

func (s *Server) serveChatWS(w http.ResponseWriter, r *http.Request) {
	if s.redirectStaleSocket(w, r) {
		return
	}
	s.chat.serveWS(w, r)
}

func (s *Server) serveOverlayWS(w http.ResponseWriter, r *http.Request) {
	if s.redirectStaleSocket(w, r) {
		return
	}
	s.overlay.serveWS(w, r)
}

// Injected is fake chat lines from POST /api/webhook. Receive from app.go.
func (s *Server) Injected() <-chan InjectedMessage {
	return s.injected
}
