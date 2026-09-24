package obs

import "time"

// HTTP paths for OBS pages, sockets, media, and operator APIs.
const (
	PathIndex     = "/"
	PathChat      = "/obs/chat"
	PathChatWS    = "/obs/chat/ws"
	PathOverlay   = "/obs/overlay"
	PathOverlayWS = "/obs/overlay/ws"
	PathScript    = "/obs/script.js"
	PathMedia     = "/obs/media/"
	PathWebhook   = "/api/webhook"
	PathInterrupt = "/api/interrupt"
	PathFavicon   = "/favicon.ico"
)

// PayloadType is a WebSocket "type" value for overlay or chat control.
type PayloadType string

const (
	// PayloadTypeAlert plays a media file from PathMedia.
	PayloadTypeAlert PayloadType = "alert"
	// PayloadTypeTTS plays synthesized WAV audio in payload.
	PayloadTypeTTS PayloadType = "tts"
	// PayloadTypeTTSInterrupt stops the current TTS utterance.
	PayloadTypeTTSInterrupt PayloadType = "tts_interrupt"
	// PayloadTypeAppClosed is sent on graceful process exit so OBS pages can
	// show that ytmemchat closed instead of a generic socket drop.
	PayloadTypeAppClosed PayloadType = "app_closed"
	// PayloadTypeChatFlush tells the chat page to remove every on-screen line.
	PayloadTypeChatFlush PayloadType = "chat_flush"
)

// OverlayEvent is JSON sent to /obs/overlay/ws (same fields as the console overlay).
type OverlayEvent struct {
	// Type is alert, tts, tts_interrupt, or app_closed.
	Type PayloadType `json:"type"`
	// Payload is raw WAV bytes for TTS (JSON base64). Unused for alerts.
	Payload []byte `json:"payload"`
	// Filename is a file under PathMedia for alerts.
	Filename string `json:"filename"`
	// Volume is playback gain from 0 to 1.
	Volume float64 `json:"volume"`
	// Scale is the visual size multiplier for alerts.
	Scale float64 `json:"scale"`
}

// AlertOverlay builds an alert event for filename inside the media directory.
func AlertOverlay(filename string, volume, scale float64) OverlayEvent {
	return OverlayEvent{Type: PayloadTypeAlert, Filename: filename, Volume: volume, Scale: scale}
}

// TTSOverlay builds a TTS event. wav is WAVE bytes (JSON-encoded as base64).
func TTSOverlay(wav []byte, volume float64) OverlayEvent {
	return OverlayEvent{Type: PayloadTypeTTS, Payload: wav, Volume: volume}
}

// InterruptOverlay builds a tts_interrupt event.
func InterruptOverlay() OverlayEvent {
	return OverlayEvent{Type: PayloadTypeTTSInterrupt}
}

// AppClosedOverlay tells overlay clients the Wails process is exiting.
func AppClosedOverlay() OverlayEvent {
	return OverlayEvent{Type: PayloadTypeAppClosed}
}

// ChatEvent is JSON sent to /obs/chat/ws (same fields as the console chat overlay).
type ChatEvent struct {
	// Type is empty for chat lines.
	// "app_closed" is a graceful process exit.
	// "chat_flush" clears every on-screen line.
	Type string `json:"type,omitempty"`
	// AuthorName is the sender display name.
	AuthorName string `json:"authorName"`
	// AuthorPicture is the sender avatar URL.
	AuthorPicture string `json:"authorPicture"`
	// MessageText is the chat line.
	MessageText string `json:"messageText"`
	// PublishedAt is when the message was published (RFC3339).
	PublishedAt string `json:"publishedAt"`
	// IsModerator is a moderator badge flag.
	IsModerator bool `json:"isModerator"`
	// IsOwner is a host badge flag.
	IsOwner bool `json:"isOwner"`
}

// NewChatEvent fills [ChatEvent] with RFC3339 PublishedAt. Zero time uses now.
func NewChatEvent(author, picture, text string, published time.Time) ChatEvent {
	if published.IsZero() {
		published = time.Now()
	}
	return ChatEvent{
		AuthorName:    author,
		AuthorPicture: picture,
		MessageText:   text,
		PublishedAt:   published.UTC().Format(time.RFC3339),
	}
}

// AppClosedChat tells chat clients the Wails process is exiting.
func AppClosedChat() ChatEvent {
	return ChatEvent{Type: string(PayloadTypeAppClosed)}
}

// FlushChat tells chat clients to remove every on-screen line.
func FlushChat() ChatEvent {
	return ChatEvent{Type: string(PayloadTypeChatFlush)}
}

// InjectedMessage is a fake chat line from POST /api/webhook.
type InjectedMessage struct {
	// Author is the synthetic sender (always "webhook").
	Author string
	// Text is the JSON "message" field.
	Text string
	// SentAt is when the handler accepted the request.
	SentAt time.Time
}

// Config is HTTP listen options for [Server].
type Config struct {
	// Addr is the listen address (for example ":8080"). Unused when tests call Handler only.
	Addr string
	// MediaPath is the directory served at PathMedia. Empty disables that route.
	MediaPath string
	// Webhooks serves PathWebhook and PathInterrupt when true.
	Webhooks bool
	// Version is the running app version (release tag or commit).
	// Empty means "dev".
	Version string
}

// QueryVersion is the cache-bust query on OBS pages, sockets, and script.js.
const QueryVersion = "v"
