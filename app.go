package main

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/HardDie/ytmemchat_wails/internal/config"
	"github.com/HardDie/ytmemchat_wails/internal/obs"
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
	runCancel      context.CancelFunc
	runWG          *sync.WaitGroup
	runGen         int
	runConnecting  bool
	runRunning     bool
	runUsingKey    bool
	runError       string
	emit           func(string, any)
}

// SettingsForm is the settings window payload. Other JSON fields stay on disk
// unchanged when saving.
type SettingsForm struct {
	// StreamID is the live video ID (v= in the watch URL). Empty is allowed until Start.
	StreamID string `json:"streamId"`
	// APIKey is an optional YouTube Data API v3 key. Empty selects the no-key client.
	APIKey string `json:"apiKey"`
	// Port is the OBS HTTP listen port ("8080" or ":8080").
	Port string `json:"port"`
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
		StreamID: s.Youtube.StreamID,
		APIKey:   s.Youtube.APIKey,
		Port:     s.Server.Port,
	}
}

// GetSettings returns the current stream ID, API key, and listen port.
func (a *App) GetSettings() SettingsForm {
	a.mu.Lock()
	defer a.mu.Unlock()
	return formFrom(a.settings)
}

// SaveSettings writes stream ID, API key, and port. Empty stream ID is allowed.
// Other persisted fields (TTS, alerts, webhook) are left as last loaded.
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
	next.Youtube.StreamID = in.StreamID
	next.Youtube.APIKey = in.APIKey
	next.Server.Port = in.Port
	if err := a.store.Save(next); err != nil {
		return err
	}
	old := a.settings
	a.loadLocked()
	if a.skipHTTP {
		return nil
	}
	if a.httpSrv == nil || !sameListenAddr(old, a.settings) {
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
