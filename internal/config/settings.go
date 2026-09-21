package config

import (
	"strings"

	"github.com/HardDie/ytmemchat_wails/internal/hotkey"
)

// CurrentVersion is stored in config.json for future migrations.
const CurrentVersion = 1

// AppDirName is the folder under os.UserConfigDir used for this app.
const AppDirName = "ytmemchat"

// FileName is the settings file inside AppDirName.
const FileName = "config.json"

// Settings is the persisted document. Unknown JSON fields are ignored.
type Settings struct {
	// Version is the settings schema version.
	Version int `json:"version"`
	// Youtube holds stream ID and optional API key.
	Youtube Youtube `json:"youtube"`
	// Server is the local HTTP listen port for OBS.
	Server Server `json:"server"`
	// TTS controls overlay speech for non-alert messages.
	TTS TTS `json:"tts"`
	// Alerts controls command-token media playback.
	Alerts Alerts `json:"alerts"`
	// Webhook enables POST /api/webhook and /api/interrupt.
	Webhook Webhook `json:"webhook"`
	// InterruptHotkey is the OS-wide shortcut that stops overlay TTS.
	InterruptHotkey InterruptHotkey `json:"interruptHotkey"`
	// Window is optional saved Wails window bounds.
	Window *Window `json:"window,omitempty"`
	// APIKeyInKeychain is true when the OS vault is in use.
	// It is filled by [Store.Load] and is not written to JSON.
	APIKeyInKeychain bool `json:"-"`
}

// Youtube is YouTube live chat connection settings.
type Youtube struct {
	// APIKey is an optional Data API v3 key. Empty selects youtube/nokey.
	APIKey string `json:"apiKey"`
	// StreamID is the live video ID (v= in the watch URL).
	StreamID string `json:"streamId"`
}

// Server is the OBS HTTP server listen settings.
type Server struct {
	// Port is a TCP port ("8080") or address (":8080").
	Port string `json:"port"`
}

// TTS is text-to-speech settings.
type TTS struct {
	// Enabled sends non-alert chat to the TTS engine when true.
	Enabled bool `json:"enabled"`
	// VoiceName is the OS voice identifier (for example "Milena").
	VoiceName string `json:"voiceName"`
}

// Alerts is meme-command settings.
type Alerts struct {
	// Enabled matches chat against commands.yaml when true.
	Enabled bool `json:"enabled"`
	// Token is a single-character command prefix (for example "@").
	Token string `json:"token"`
	// MediaPath is the directory of alert media files.
	MediaPath string `json:"mediaPath"`
	// CommandsFilePath is the YAML file of command names to files.
	CommandsFilePath string `json:"commandsFilePath"`
}

// Webhook toggles the local operator HTTP API.
type Webhook struct {
	// Enabled serves /api/webhook and /api/interrupt when true.
	Enabled bool `json:"enabled"`
}

// InterruptHotkey is a global keyboard shortcut to interrupt overlay speech.
type InterruptHotkey struct {
	// Enabled is nil in old files (treated as on). Pointer so JSON can disable it.
	Enabled *bool `json:"enabled"`
	// Chord is a string such as "Ctrl+Shift+I". Empty uses the default.
	Chord string `json:"chord"`
}

// IsEnabled reports whether the global interrupt shortcut should be registered.
func (h InterruptHotkey) IsEnabled() bool {
	if h.Enabled == nil {
		return true
	}
	return *h.Enabled
}

func optBool(v bool) *bool {
	b := v
	return &b
}

// Window is saved desktop window geometry.
type Window struct {
	// Width is the window width in pixels.
	Width int `json:"width"`
	// Height is the window height in pixels.
	Height int `json:"height"`
	// X is the window origin X in pixels.
	X int `json:"x"`
	// Y is the window origin Y in pixels.
	Y int `json:"y"`
}

// Defaults returns first-launch settings (empty stream ID and API key).
// Alerts and TTS stay off until the operator enables them.
func Defaults() Settings {
	return Settings{
		Version: CurrentVersion,
		Server:  Server{Port: "8080"},
		TTS:     TTS{Enabled: false},
		Alerts: Alerts{
			Enabled: false,
			Token:   "@",
		},
		Webhook: Webhook{Enabled: false},
		InterruptHotkey: InterruptHotkey{
			Enabled: optBool(true),
			Chord:   hotkey.DefaultChord,
		},
	}
}

// TrimSpace trims user-facing string fields. It does not log the API key.
func (s Settings) TrimSpace() Settings {
	s.Youtube.APIKey = strings.TrimSpace(s.Youtube.APIKey)
	s.Youtube.StreamID = strings.TrimSpace(s.Youtube.StreamID)
	s.Server.Port = strings.TrimSpace(s.Server.Port)
	s.TTS.VoiceName = strings.TrimSpace(s.TTS.VoiceName)
	s.Alerts.Token = strings.TrimSpace(s.Alerts.Token)
	s.Alerts.MediaPath = strings.TrimSpace(s.Alerts.MediaPath)
	s.Alerts.CommandsFilePath = strings.TrimSpace(s.Alerts.CommandsFilePath)
	s.InterruptHotkey.Chord = strings.TrimSpace(s.InterruptHotkey.Chord)
	if s.InterruptHotkey.Chord == "" {
		s.InterruptHotkey.Chord = hotkey.DefaultChord
	}
	if c, err := hotkey.Parse(s.InterruptHotkey.Chord); err == nil {
		s.InterruptHotkey.Chord = c.String()
	}
	return s
}

// HasAPIKey reports whether a non-empty YouTube API key is set (after trim).
func (s Settings) HasAPIKey() bool {
	return strings.TrimSpace(s.Youtube.APIKey) != ""
}
