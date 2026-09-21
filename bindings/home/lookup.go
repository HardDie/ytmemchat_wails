package home

import (
	"context"
	"fmt"
	"strings"

	"github.com/HardDie/ytmemchat_wails/internal/obs"
)

// ErrOBSNotListening is returned when overlay HTTP is down.
var ErrOBSNotListening = fmt.Errorf("OBS HTTP is not listening")

// InterruptTTS publishes overlay type tts_interrupt. The HTTP interrupt route
// is not required; this works whenever OBS HTTP is listening.
func (h *Home) InterruptTTS() error {
	srv := h.app.OverlayServer()
	if srv == nil {
		return ErrOBSNotListening
	}
	srv.PublishOverlay(obs.InterruptOverlay())
	return nil
}

// LookupLatestStream finds the channel from streamID, then the current live
// stream or newest upcoming stream (not a VOD). Requires a Data API key.
// apiKey may be empty to use the saved key. Does not write settings.
func (h *Home) LookupLatestStream(streamID, apiKey string) (StreamLookup, error) {
	vid := strings.TrimSpace(streamID)
	key := strings.TrimSpace(apiKey)
	savedID, savedKey := h.app.SavedYouTube()
	if vid == "" {
		vid = strings.TrimSpace(savedID)
	}
	if key == "" {
		key = strings.TrimSpace(savedKey)
	}
	ctx := h.app.DialogContext()
	if ctx == nil {
		ctx = context.Background()
	}
	if vid == "" {
		return StreamLookup{}, fmt.Errorf("a previous stream ID is required to find the channel")
	}
	if key == "" {
		return StreamLookup{}, fmt.Errorf("finding the latest stream requires a YouTube API key")
	}
	got, err := h.app.LookupBroadcast(ctx, key, vid)
	if err != nil {
		return StreamLookup{}, err
	}
	h.app.NotifyRun()
	return StreamLookup{StreamID: got.VideoID, ChannelID: got.ChannelID, Kind: string(got.Kind)}, nil
}
