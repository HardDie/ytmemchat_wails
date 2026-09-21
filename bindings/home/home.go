// Package home is the Wails binding for the Home pane.
package home

import (
	"context"

	"github.com/HardDie/ytmemchat_wails/internal/obs"
	"github.com/HardDie/ytmemchat_wails/internal/youtube"
)

// Home exposes Start/Stop, OBS URLs, interrupt, and stream lookup.
type Home struct {
	app api
}

// api is the desktop App core this pane uses.
type api interface {
	Start() error
	Stop()
	GetRunStatus() RunStatus
	GetOBSStatus() OBSStatus
	OverlayServer() *obs.Server
	SavedYouTube() (streamID, apiKey string)
	DialogContext() context.Context
	LookupBroadcast(ctx context.Context, apiKey, videoID string) (youtube.LatestBroadcast, error)
	NotifyRun()
}

// New wraps the application core for the Home pane.
func New(app api) *Home {
	return &Home{app: app}
}

// GetOBSStatus returns whether OBS HTTP is listening and the Browser Source URLs.
func (h *Home) GetOBSStatus() OBSStatus {
	return h.app.GetOBSStatus()
}

// GetRunStatus returns whether YouTube chat ingest is running.
func (h *Home) GetRunStatus() RunStatus {
	return h.app.GetRunStatus()
}

// Start begins YouTube chat ingest into the OBS chat socket, then fans each
// line to alerts or TTS on the overlay. Uses last saved settings (save the
// form first). Connect work runs in a goroutine so the UI is not blocked.
func (h *Home) Start() error {
	return h.app.Start()
}

// Stop cancels the YouTube iterator. OBS HTTP stays up.
func (h *Home) Stop() {
	h.app.Stop()
}
