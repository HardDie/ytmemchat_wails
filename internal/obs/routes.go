package obs

import (
	"net/http"
	"strings"
)

func (s *Server) routes() {
	s.mux.HandleFunc("GET /{$}", s.serveIndex)
	s.mux.HandleFunc("GET "+PathChat, s.serveChatHTML)
	s.mux.HandleFunc("GET "+PathChatWS, s.chat.serveWS)
	s.mux.HandleFunc("GET "+PathOverlay, s.serveOverlayHTML)
	s.mux.HandleFunc("GET "+PathOverlayWS, s.overlay.serveWS)
	s.mux.HandleFunc("GET "+PathScript, s.serveScript)
	s.mux.HandleFunc("GET "+PathFavicon, s.serveFavicon)
	if strings.TrimSpace(s.cfg.MediaPath) != "" {
		fs := http.StripPrefix(PathMedia, http.FileServer(http.Dir(s.cfg.MediaPath)))
		s.mux.Handle("GET "+PathMedia, fs)
	}
	if s.cfg.Webhooks {
		s.mux.HandleFunc("POST "+PathWebhook, s.handleWebhook)
		s.mux.HandleFunc("POST "+PathInterrupt, s.handleInterrupt)
	}
}
