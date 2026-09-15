package obs

import (
	_ "embed"
	"net/http"
)

var (
	//go:embed index.html
	indexHTML []byte
	//go:embed chat.html
	chatHTML []byte
	//go:embed overlay.html
	overlayHTML []byte
	//go:embed favicon.png
	faviconPNG []byte
)

func writeHTML(w http.ResponseWriter, body []byte) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(body)
}

func (s *Server) serveIndex(w http.ResponseWriter, _ *http.Request) {
	writeHTML(w, indexHTML)
}

func (s *Server) serveChatHTML(w http.ResponseWriter, _ *http.Request) {
	writeHTML(w, chatHTML)
}

func (s *Server) serveOverlayHTML(w http.ResponseWriter, _ *http.Request) {
	writeHTML(w, overlayHTML)
}

func (s *Server) serveFavicon(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "image/png")
	_, _ = w.Write(faviconPNG)
}
