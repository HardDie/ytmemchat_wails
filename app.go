package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/HardDie/ytmemchat_wails/internal/alerts"
	"github.com/HardDie/ytmemchat_wails/internal/config"
	"github.com/HardDie/ytmemchat_wails/internal/obs"
	"github.com/HardDie/ytmemchat_wails/internal/tts"
	"github.com/HardDie/ytmemchat_wails/internal/youtube"
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
	overlay        atomic.Pointer[overlayState]
	injectCancel   context.CancelFunc
	injectWG       *sync.WaitGroup
	runCancel      context.CancelFunc
	runWG          *sync.WaitGroup
	runGen         int
	runConnecting  bool
	runRunning     bool
	runUsingKey    bool
	runError       string
	quotaStreamID  string
	emit           func(string, any)
	lookupLatest   func(context.Context, string, string) (youtube.LatestBroadcast, error)
	hk             interruptHotkey
	hotkeyErr      string
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
	// InterruptHotkeyEnabled registers an OS-wide interrupt shortcut.
	InterruptHotkeyEnabled bool `json:"interruptHotkeyEnabled"`
	// InterruptHotkeyChord is the shortcut, for example "Ctrl+Shift+I".
	InterruptHotkeyChord string `json:"interruptHotkeyChord"`
	// InterruptHotkeyError is a register failure; empty when the shortcut is active or off.
	InterruptHotkeyError string `json:"interruptHotkeyError"`
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
	a.loadLocked()
	if err := a.startHTTPLocked(); err != nil {
		slog.Error("obs listen failed", "err", err)
	}
	a.quotaStreamID = strings.TrimSpace(a.settings.Youtube.StreamID)
	a.mu.Unlock()
	a.syncInterruptHotkey()
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

// GetSettings returns the current settings for the window.
func (a *App) GetSettings() SettingsForm {
	a.mu.Lock()
	defer a.mu.Unlock()
	return formFrom(a.settings, a.hotkeyErr)
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

// LookupLatestStream finds the channel from streamID, then the current live
// stream or newest upcoming stream (not a VOD). Requires a Data API key.
// apiKey may be empty to use the saved key. Does not write settings.
func (a *App) LookupLatestStream(streamID, apiKey string) (StreamLookup, error) {
	vid := strings.TrimSpace(streamID)
	key := strings.TrimSpace(apiKey)
	a.mu.Lock()
	if vid == "" {
		vid = strings.TrimSpace(a.settings.Youtube.StreamID)
	}
	if key == "" {
		key = strings.TrimSpace(a.settings.Youtube.APIKey)
	}
	ctx := a.ctx
	fn := a.lookupLatest
	a.mu.Unlock()
	if ctx == nil {
		ctx = context.Background()
	}
	if vid == "" {
		return StreamLookup{}, fmt.Errorf("a previous stream ID is required to find the channel")
	}
	if key == "" {
		return StreamLookup{}, fmt.Errorf("finding the latest stream requires a YouTube API key")
	}
	if fn == nil {
		fn = youtube.LookupLatestBroadcast
	}
	got, err := fn(ctx, key, vid)
	if err != nil {
		return StreamLookup{}, err
	}
	a.mu.Lock()
	a.resetQuotaIfStreamIDChangedLocked(got.VideoID)
	a.emitRunLocked()
	a.mu.Unlock()
	return StreamLookup{StreamID: got.VideoID, ChannelID: got.ChannelID, Kind: string(got.Kind)}, nil
}

// SaveSettings writes the settings form. Empty stream ID is allowed.
// Changing listen port, media path, or the webhook flag restarts OBS HTTP.
// Errors never include the API key.
func (a *App) SaveSettings(in SettingsForm) error {
	a.mu.Lock()
	if a.store == nil {
		err := a.storeErr
		a.mu.Unlock()
		if err != nil {
			return fmt.Errorf("config: %w", err)
		}
		return fmt.Errorf("config: no store")
	}
	next := a.settings
	applyForm(&next, in)
	if err := a.store.Save(next); err != nil {
		a.mu.Unlock()
		return err
	}
	old := a.settings
	a.loadLocked()
	if strings.TrimSpace(old.Youtube.StreamID) != strings.TrimSpace(a.settings.Youtube.StreamID) {
		a.resetQuotaIfStreamIDChangedLocked(a.settings.Youtube.StreamID)
		a.emitRunLocked()
	}
	skip := a.skipHTTP
	var httpErr error
	if !skip {
		if a.httpSrv == nil || !sameOBSListen(old, a.settings) {
			httpErr = a.startHTTPLocked()
		} else {
			a.installOverlayLocked(false)
		}
	}
	a.mu.Unlock()
	a.syncInterruptHotkey()
	return httpErr
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

// AlertCommand is one row in commands.yaml for the editor.
type AlertCommand struct {
	// Name is the token word (for example "jump" in @jump).
	Name string `json:"name"`
	// File is a media filename inside the alerts media folder.
	File string `json:"file"`
	// Volume is overlay gain 0–1. Nil omits the YAML key (runtime default 1).
	Volume *float64 `json:"volume,omitempty"`
	// Scale is visual size. Nil omits the YAML key (runtime default 1).
	Scale *float64 `json:"scale,omitempty"`
}

// AlertCommandsFile is the commands.yaml editor payload.
type AlertCommandsFile struct {
	// Path is the saved commands.yaml location.
	Path string `json:"path"`
	// Commands is the list to show and write.
	Commands []AlertCommand `json:"commands"`
}

// GetAlertCommands loads the saved commands.yaml. A missing file yields an empty list.
func (a *App) GetAlertCommands() (AlertCommandsFile, error) {
	a.mu.Lock()
	path := strings.TrimSpace(a.settings.Alerts.CommandsFilePath)
	a.mu.Unlock()
	if path == "" {
		return AlertCommandsFile{}, fmt.Errorf("set a commands.yaml path in Configuration")
	}
	parsed, err := alerts.LoadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return AlertCommandsFile{Path: path, Commands: []AlertCommand{}}, nil
		}
		return AlertCommandsFile{}, err
	}
	return AlertCommandsFile{Path: path, Commands: toAlertCommands(parsed.Commands)}, nil
}

// SaveAlertCommands writes commands.yaml. Empty volume and scale are omitted.
func (a *App) SaveAlertCommands(in AlertCommandsFile) error {
	a.mu.Lock()
	path := strings.TrimSpace(a.settings.Alerts.CommandsFilePath)
	skip := a.skipHTTP
	a.mu.Unlock()
	if path == "" {
		return fmt.Errorf("set a commands.yaml path in Configuration")
	}
	if err := alerts.SaveFile(path, alerts.File{Commands: fromAlertCommands(in.Commands)}); err != nil {
		return err
	}
	if skip {
		return nil
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.installOverlayLocked(false)
	return nil
}

func toAlertCommands(in []alerts.Command) []AlertCommand {
	out := make([]AlertCommand, len(in))
	for i, c := range in {
		out[i] = AlertCommand{Name: c.Name, File: c.File, Volume: c.Volume, Scale: c.Scale}
	}
	return out
}

func fromAlertCommands(in []AlertCommand) []alerts.Command {
	out := make([]alerts.Command, len(in))
	for i, c := range in {
		out[i] = alerts.Command{Name: strings.TrimSpace(c.Name), File: strings.TrimSpace(c.File), Volume: c.Volume, Scale: c.Scale}
	}
	return out
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
