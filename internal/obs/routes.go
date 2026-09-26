package obs

import (
	"net/http"
	"strings"
)

func (s *Server) routes() {
	s.mux.HandleFunc("GET /{$}", s.serveIndex)
	s.mux.HandleFunc("GET "+PathChat, s.serveChatHTML)
	s.mux.HandleFunc("GET "+PathChatWS, s.serveChatWS)
	s.mux.HandleFunc("GET "+PathOverlay, s.serveOverlayHTML)
	s.mux.HandleFunc("GET "+PathOverlayWS, s.serveOverlayWS)
	s.mux.HandleFunc("GET "+PathScript, s.serveScript)
	s.mux.HandleFunc("GET "+PathFavicon, s.serveFavicon)
	if strings.TrimSpace(s.cfg.MediaPath) != "" {
		fs := http.StripPrefix(PathMedia, http.FileServer(http.Dir(s.cfg.MediaPath)))
		s.mux.Handle("GET "+PathMedia, noStore(fs))
	}
	if s.cfg.Webhooks {
		s.mux.HandleFunc("POST "+PathWebhook, s.handleWebhook)
		s.mux.HandleFunc("POST "+PathInterrupt, s.handleInterrupt)
	}
}

// noStore tells the browser not to keep the response.
// FileServer drops Cache-Control before it writes an error status.
func noStore(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(noStoreWriter{w}, r)
	})
}

type noStoreWriter struct {
	http.ResponseWriter
}

func (w noStoreWriter) WriteHeader(code int) {
	w.Header().Set("Cache-Control", "no-store")
	w.ResponseWriter.WriteHeader(code)
}

func (w noStoreWriter) Write(b []byte) (int, error) {
	w.Header().Set("Cache-Control", "no-store")
	return w.ResponseWriter.Write(b)
}

func (w noStoreWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}
