package configuration

// SettingsForm is the settings window payload (stream, modules, and paths).
type SettingsForm struct {
	// StreamID is the live video ID (v= in the watch URL). Empty is allowed until Start.
	StreamID string `json:"streamId"`
	// APIKey is an optional YouTube Data API v3 key. Empty selects the no-key client.
	APIKey string `json:"apiKey"`
	// Port is the OBS HTTP listen port ("8080" or ":8080").
	Port string `json:"port"`
	// TTSEnabled sends non-alert chat to TTS when true.
	TTSEnabled bool `json:"ttsEnabled"`
	// TTSVoiceName is the OS voice identifier (empty uses the system default).
	TTSVoiceName string `json:"ttsVoiceName"`
	// AlertsEnabled matches chat against commands.yaml when true.
	AlertsEnabled bool `json:"alertsEnabled"`
	// AlertsToken is a single-character command prefix (for example "@").
	AlertsToken string `json:"alertsToken"`
	// AlertsMediaPath is the directory served at /obs/media/.
	AlertsMediaPath string `json:"alertsMediaPath"`
	// AlertsCommandsFilePath is the YAML command list.
	// GetSettings returns the effective path. Save stores a custom path only.
	AlertsCommandsFilePath string `json:"alertsCommandsFilePath"`
	// AlertsCommandsFileCustom shows the path field. Off stores an empty path.
	AlertsCommandsFileCustom bool `json:"alertsCommandsFileCustom"`
	// WebhookEnabled serves POST /api/webhook and /api/interrupt.
	WebhookEnabled bool `json:"webhookEnabled"`
	// InterruptHotkeyEnabled registers an OS-wide interrupt shortcut.
	InterruptHotkeyEnabled bool `json:"interruptHotkeyEnabled"`
	// InterruptHotkeyChord is the shortcut, for example "Ctrl+Shift+I".
	InterruptHotkeyChord string `json:"interruptHotkeyChord"`
	// InterruptHotkeyError is a register failure; empty when the shortcut is active or off.
	InterruptHotkeyError string `json:"interruptHotkeyError"`
	// APIKeyInKeychain is true when the OS vault is available for the API key.
	APIKeyInKeychain bool `json:"apiKeyInKeychain"`
	// Debug logs chat, command matching, and TTS. Off by default.
	// The same flag is sent on OBS WebSocket events.
	Debug bool `json:"debug"`
}

// TTSVoice is one installed OS voice for the settings list.
type TTSVoice struct {
	// Name is passed to the TTS engine (for example "Milena").
	Name string `json:"name"`
	// Languages is a readable list such as "Russian (ru_RU)".
	Languages string `json:"languages"`
	// Gender is Male, Female, or empty.
	Gender string `json:"gender"`
	// Details is extra OS description text.
	Details string `json:"details"`
}
