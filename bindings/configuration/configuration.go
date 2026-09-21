// Package configuration is the Wails binding for the Configuration pane.
package configuration

import (
	"context"

	"github.com/HardDie/ytmemchat_wails/internal/config"
)

// Configuration exposes settings, TTS voices, and path pickers.
type Configuration struct {
	app api
}

// api is the desktop App core this pane uses.
type api interface {
	SettingsSnapshot() (config.Settings, string)
	ConfigFilePath() string
	PersistSettings(next config.Settings) error
	DialogContext() context.Context
}

// New wraps the application core for the Configuration pane.
func New(app api) *Configuration {
	return &Configuration{app: app}
}

// ConfigPath is the JSON file path, or empty if the store could not be created.
func (c *Configuration) ConfigPath() string {
	return c.app.ConfigFilePath()
}

// GetSettings returns the current settings for the window.
func (c *Configuration) GetSettings() SettingsForm {
	s, hotkeyErr := c.app.SettingsSnapshot()
	return formFrom(s, hotkeyErr)
}

// SaveSettings writes the settings form. Empty stream ID is allowed.
// Changing listen port, media path, or the webhook flag restarts OBS HTTP.
// Errors never include the API key.
func (c *Configuration) SaveSettings(in SettingsForm) error {
	cur, _ := c.app.SettingsSnapshot()
	next := cur
	applyForm(&next, in)
	return c.app.PersistSettings(next)
}

func formFrom(s config.Settings, hotkeyErr string) SettingsForm {
	return SettingsForm{
		StreamID:               s.Youtube.StreamID,
		APIKey:                 s.Youtube.APIKey,
		Port:                   s.Server.Port,
		TTSEnabled:             s.TTS.Enabled,
		TTSVoiceName:           s.TTS.VoiceName,
		AlertsEnabled:          s.Alerts.Enabled,
		AlertsToken:            s.Alerts.Token,
		AlertsMediaPath:        s.Alerts.MediaPath,
		AlertsCommandsFilePath: s.Alerts.CommandsFilePath,
		WebhookEnabled:         s.Webhook.Enabled,
		InterruptHotkeyEnabled: s.InterruptHotkey.IsEnabled(),
		InterruptHotkeyChord:   s.InterruptHotkey.Chord,
		InterruptHotkeyError:   hotkeyErr,
	}
}

func applyForm(dst *config.Settings, in SettingsForm) {
	dst.Youtube.StreamID = in.StreamID
	dst.Youtube.APIKey = in.APIKey
	dst.Server.Port = in.Port
	dst.TTS.Enabled = in.TTSEnabled
	dst.TTS.VoiceName = in.TTSVoiceName
	dst.Alerts.Enabled = in.AlertsEnabled
	dst.Alerts.Token = in.AlertsToken
	dst.Alerts.MediaPath = in.AlertsMediaPath
	dst.Alerts.CommandsFilePath = in.AlertsCommandsFilePath
	dst.Webhook.Enabled = in.WebhookEnabled
	en := in.InterruptHotkeyEnabled
	dst.InterruptHotkey.Enabled = &en
	dst.InterruptHotkey.Chord = in.InterruptHotkeyChord
}
