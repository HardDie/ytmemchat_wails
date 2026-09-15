package main

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/HardDie/ytmemchat_wails/internal/config"
	"github.com/HardDie/ytmemchat_wails/internal/obs"
	"github.com/HardDie/ytmemchat_wails/internal/tts"
)

// App is the Wails bindings façade (settings, OBS HTTP, YouTube chat Start/Stop).
type App struct {
	ctx            context.Context
	mu             sync.Mutex
	store          *config.Store
	storeErr       error
	settings       config.Settings
	httpSrv        *obs.Server
	httpErr        error
	httpAddr       string
	listenOverride string
	skipHTTP       bool
	clientFn       clientFactory
	newMatcher     matcherFactory
	newSynth       synthFactory
	runCancel      context.CancelFunc
	runWG          *sync.WaitGroup
	runGen         int
	runConnecting  bool
	runRunning     bool
	runUsingKey    bool
	runError       string
	emit           func(string, any)
}

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
	AlertsCommandsFilePath string `json:"alertsCommandsFilePath"`
	// WebhookEnabled serves POST /api/webhook and /api/interrupt.
	WebhookEnabled bool `json:"webhookEnabled"`
}

// NewApp returns the bound application struct with the default config store.
func NewApp() *App {
	a := &App{settings: config.Defaults()}
	st, err := config.NewDefaultStore()
	if err != nil {
		a.storeErr = err
		return a
	}
	a.store = st
	return a
}

func newAppWithStore(st *config.Store) *App {
	return &App{store: st, settings: config.Defaults(), skipHTTP: true}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.setupWailsHooks()
	a.mu.Lock()
	defer a.mu.Unlock()
	a.loadLocked()
	if err := a.startHTTPLocked(); err != nil {
		slog.Error("obs listen failed", "err", err)
	}
}

func (a *App) loadLocked() {
	if a.store == nil {
		return
	}
	s, err := a.store.Load()
	if err != nil {
		slog.Error("config load failed", "err", err)
		return
	}
	a.settings = s
}

func formFrom(s config.Settings) SettingsForm {
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
}

// GetSettings returns the current settings for the window.
func (a *App) GetSettings() SettingsForm {
	a.mu.Lock()
	defer a.mu.Unlock()
	return formFrom(a.settings)
}

// SaveSettings writes the settings form. Empty stream ID is allowed.
// Changing listen port, media path, or the webhook flag restarts OBS HTTP.
// Errors never include the API key.
func (a *App) SaveSettings(in SettingsForm) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.store == nil {
		if a.storeErr != nil {
			return fmt.Errorf("config: %w", a.storeErr)
		}
		return fmt.Errorf("config: no store")
	}
	next := a.settings
	applyForm(&next, in)
	if err := a.store.Save(next); err != nil {
		return err
	}
	old := a.settings
	a.loadLocked()
	if a.skipHTTP {
		return nil
	}
	if a.httpSrv == nil || !sameOBSListen(old, a.settings) {
		if err := a.startHTTPLocked(); err != nil {
			return err
		}
	}
	return nil
}

// ConfigPath is the JSON file path, or empty if the store could not be created.
func (a *App) ConfigPath() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.store == nil {
		return ""
	}
	return a.store.Path()
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

// GetTTSVoices lists installed voices with language and gender. Empty if the engine is missing.
func (a *App) GetTTSVoices() []TTSVoice {
	voices, err := tts.GetAvailableVoices()
	if err != nil {
		return []TTSVoice{}
	}
	return ttsVoicesFrom(voices)
}

func ttsVoicesFrom(list []tts.VoiceInfo) []TTSVoice {
	seen := make(map[string]int)
	out := make([]TTSVoice, 0, len(list))
	for _, v := range list {
		if v.Name == "" {
			continue
		}
		langs := tts.FormatLanguages(v.Language)
		if i, ok := seen[v.Name]; ok {
			out[i].Languages = mergeLanguageLabels(out[i].Languages, langs)
			if out[i].Gender == "" {
				out[i].Gender = v.Gender
			}
			if out[i].Details == "" {
				out[i].Details = v.Details
			}
			continue
		}
		seen[v.Name] = len(out)
		out = append(out, TTSVoice{
			Name:      v.Name,
			Languages: langs,
			Gender:    v.Gender,
			Details:   v.Details,
		})
	}
	return out
}

func mergeLanguageLabels(existing, next string) string {
	if next == "" {
		return existing
	}
	if existing == "" {
		return next
	}
	seen := map[string]struct{}{}
	var parts []string
	for _, p := range strings.Split(existing+", "+next, ", ") {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		parts = append(parts, p)
	}
	return strings.Join(parts, ", ")
}
