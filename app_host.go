package main

import (
	"context"
	"fmt"
	"time"

	"github.com/HardDie/ytmemchat_wails/internal/config"
	"github.com/HardDie/ytmemchat_wails/internal/obs"
	"github.com/HardDie/ytmemchat_wails/internal/youtube"
)

// SettingsSnapshot is the saved settings plus the current hotkey error.
func (a *App) SettingsSnapshot() (config.Settings, string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.settings, a.hotkeyErr
}

// ConfigFilePath is the JSON file path, or empty if the store could not be created.
func (a *App) ConfigFilePath() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.store == nil {
		return ""
	}
	return a.store.Path()
}

// PersistSettings writes next to disk, reloads memory, and may restart OBS HTTP.
func (a *App) PersistSettings(next config.Settings) error {
	a.mu.Lock()
	if a.store == nil {
		err := a.storeErr
		a.mu.Unlock()
		if err != nil {
			return fmt.Errorf("config: %w", err)
		}
		return fmt.Errorf("config: no store")
	}
	if err := a.store.Save(next); err != nil {
		a.mu.Unlock()
		return err
	}
	old := a.settings
	a.loadLocked()
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

// DialogContext is the Wails window context, or nil before startup.
func (a *App) DialogContext() context.Context {
	return a.ctx
}

// CommandsPath is the saved commands.yaml path.
func (a *App) CommandsPath() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.settings.Alerts.CommandsFilePath
}

// MediaPath is the saved alerts media folder.
func (a *App) MediaPath() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.settings.Alerts.MediaPath
}

// SkipHTTP is true in tests that do not start the OBS listener.
func (a *App) SkipHTTP() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.skipHTTP
}

// ReloadOverlay rebuilds the alerts/TTS overlay path from current settings.
func (a *App) ReloadOverlay() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.installOverlayLocked(false)
}

// OverlayServer is the OBS HTTP server, or nil if it is not listening.
func (a *App) OverlayServer() *obs.Server {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.httpSrv
}

// DispatchChat fans one line through the same overlay path as live chat.
func (a *App) DispatchChat(fallback *obs.Server, author, text string) {
	a.dispatchLine(fallback, &youtube.ChatMessage{
		Author:    author,
		Message:   text,
		Timestamp: time.Now(),
	})
}

// SavedYouTube returns the persisted stream ID and API key.
func (a *App) SavedYouTube() (streamID, apiKey string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.settings.Youtube.StreamID, a.settings.Youtube.APIKey
}

// LookupBroadcast resolves a live or upcoming video. Tests may stub lookupLatest.
func (a *App) LookupBroadcast(ctx context.Context, apiKey, videoID string) (youtube.LatestBroadcast, error) {
	a.mu.Lock()
	fn := a.lookupLatest
	a.mu.Unlock()
	if fn == nil {
		fn = youtube.LookupLatestBroadcast
	}
	return fn(ctx, apiKey, videoID)
}

// NotifyRun emits the current pipeline status to the window.
func (a *App) NotifyRun() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.emitRunLocked()
}
