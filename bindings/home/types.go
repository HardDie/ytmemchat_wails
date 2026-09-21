package home

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

// RunStatus is Start/Stop state for the settings window.
type RunStatus struct {
	// Running is true while the YouTube iterator is live.
	Running bool `json:"running"`
	// Connecting is true after Start until the iterator is ready or failed.
	Connecting bool `json:"connecting"`
	// UsingAPIKey is true when the v3 client was selected (key present).
	UsingAPIKey bool `json:"usingApiKey"`
	// QuotaUnits is this app’s estimated spend in the default Data API bucket today.
	QuotaUnits int `json:"quotaUnits"`
	// QuotaUnitsLimit is the documented default daily unit budget (not Cloud-approved quota).
	QuotaUnitsLimit int `json:"quotaUnitsLimit"`
	// QuotaSearch is this process’s estimated search.list spend.
	QuotaSearch int `json:"quotaSearch"`
	// QuotaSearchLimit is the documented default daily search.list budget.
	QuotaSearchLimit int `json:"quotaSearchLimit"`
	// Error is a user-facing failure; empty when ok. Never includes the API key.
	Error string `json:"error"`
}

// StreamLookup is a live or upcoming video resolved from a known stream ID.
type StreamLookup struct {
	// StreamID is the watch URL v= value to use for Start.
	StreamID string `json:"streamId"`
	// ChannelID is the UC… channel that owns the video.
	ChannelID string `json:"channelId"`
	// Kind is "live" or "upcoming".
	Kind string `json:"kind"`
}
