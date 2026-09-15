package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"strings"
	"time"

	"github.com/HardDie/ytmemchat_wails/internal/config"
	"github.com/HardDie/ytmemchat_wails/internal/obs"
)

const httpShutdownTimeout = 5 * time.Second

// OBSStatus is the live HTTP listener state for the settings window.
type OBSStatus struct {
	// Listening is true when the OBS HTTP server accepted a bind.
	Listening bool `json:"listening"`
	// Error is a bind/serve problem; empty when ok.
	Error string `json:"error"`
	// ChatURL is the OBS Browser Source URL for chat.
	ChatURL string `json:"chatUrl"`
	// OverlayURL is the OBS Browser Source URL for alerts and TTS.
	OverlayURL string `json:"overlayUrl"`
	// IndexURL lists those URLs in a browser (not for OBS).
	IndexURL string `json:"indexUrl"`
}

func (a *App) shutdown(_ context.Context) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.stopHTTPLocked()
}

func (a *App) startHTTPLocked() error {
	a.stopHTTPLocked()
	addr := a.settings.Server.ListenAddr()
	if a.listenOverride != "" {
		addr = a.listenOverride
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		a.httpErr = err
		a.httpAddr = ""
		return fmt.Errorf("obs listen %s: %w", addr, err)
	}
	media := ""
	if a.settings.Alerts.Enabled {
		media = strings.TrimSpace(a.settings.Alerts.MediaPath)
	}
	srv := obs.New(obs.Config{
		Addr:      addr,
		MediaPath: media,
		Webhooks:  a.settings.Webhook.Enabled,
	})
	a.httpSrv = srv
	a.httpAddr = ln.Addr().String()
	a.httpErr = nil
	go func() {
		if err := srv.Serve(ln); err != nil {
			slog.Error("obs serve ended", "err", err)
		}
	}()
	return nil
}

func (a *App) stopHTTPLocked() {
	if a.httpSrv == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), httpShutdownTimeout)
	defer cancel()
	if err := a.httpSrv.Shutdown(ctx); err != nil {
		slog.Error("obs shutdown", "err", err)
	}
	a.httpSrv = nil
	a.httpAddr = ""
}

func (a *App) obsStatusLocked() OBSStatus {
	display := a.settings.Server.ListenAddr()
	if a.httpAddr != "" {
		display = a.httpAddr
	}
	st := OBSStatus{
		Listening:  a.httpSrv != nil && a.httpErr == nil,
		ChatURL:    obs.ChatSourceURL(display),
		OverlayURL: obs.OverlaySourceURL(display),
		IndexURL:   obsIndexURL(display),
	}
	if a.httpErr != nil {
		st.Error = a.httpErr.Error()
		st.Listening = false
	}
	return st
}

func obsIndexURL(addr string) string {
	chat := obs.ChatSourceURL(addr)
	return strings.TrimSuffix(chat, obs.PathChat) + "/"
}

// GetOBSStatus returns whether OBS HTTP is listening and the Browser Source URLs.
func (a *App) GetOBSStatus() OBSStatus {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.obsStatusLocked()
}

func sameListenAddr(old, next config.Settings) bool {
	return old.Server.ListenAddr() == next.Server.ListenAddr()
}
