package obs

import (
	"encoding/json"
	"net/http"
	"time"
)

const webhookAuthor = "webhook"

type webhookBody struct {
	Message string `json:"message"`
}

func (s *Server) handleWebhook(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
	var body webhookBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	msg := InjectedMessage{
		Author: webhookAuthor,
		Text:   body.Message,
		SentAt: time.Now(),
	}
	select {
	case s.injected <- msg:
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, "inject queue full", http.StatusServiceUnavailable)
	}
}

func (s *Server) handleInterrupt(w http.ResponseWriter, _ *http.Request) {
	s.PublishOverlay(InterruptOverlay())
	w.WriteHeader(http.StatusNoContent)
}
