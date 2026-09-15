package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/HardDie/ytmemchat_wails/internal/config"
	"github.com/HardDie/ytmemchat_wails/internal/obs"
	"github.com/HardDie/ytmemchat_wails/internal/youtube"
	"github.com/HardDie/ytmemchat_wails/internal/youtube/nokey"
)

type clientFactory func(config.Settings) (youtube.Client, error)

// RunStatus is Start/Stop state for the settings window.
type RunStatus struct {
	// Running is true while the YouTube iterator is live.
	Running bool `json:"running"`
	// Connecting is true after Start until the iterator is ready or failed.
	Connecting bool `json:"connecting"`
	// UsingAPIKey is true when the v3 client was selected (key present).
	UsingAPIKey bool `json:"usingApiKey"`
	// Error is a user-facing failure; empty when ok. Never includes the API key.
	Error string `json:"error"`
}

func chooseYouTubeClient(s config.Settings) (youtube.Client, error) {
	if s.HasAPIKey() {
		return youtube.New(s.Youtube.APIKey)
	}
	return nokey.New(), nil
}

func publicPipelineError(err error) string {
	if err == nil {
		return ""
	}
	switch {
	case errors.Is(err, config.ErrStreamIDRequired):
		return "Stream ID is required to Start"
	case youtube.IsInvalidAPIKey(err):
		return "YouTube API key is invalid"
	case errors.Is(err, youtube.ErrNotLive):
		return "This video is not a current live stream with chat"
	case errors.Is(err, youtube.ErrQuotaExceeded):
		return "YouTube API quota exceeded"
	case errors.Is(err, youtube.ErrEmptyAPIKey):
		return "YouTube API key is empty"
	default:
		return err.Error()
	}
}

func (a *App) factory() clientFactory {
	if a.clientFn != nil {
		return a.clientFn
	}
	return chooseYouTubeClient
}

func (a *App) stopRunLocked() (cancel context.CancelFunc, wg *sync.WaitGroup) {
	cancel = a.runCancel
	a.runCancel = nil
	a.runConnecting = false
	a.runRunning = false
	wg = a.runWG
	return cancel, wg
}

func (a *App) emitRunLocked() {
	if a.emit == nil {
		return
	}
	a.emit("pipeline", a.runStatusLocked())
}

func (a *App) runStatusLocked() RunStatus {
	return RunStatus{
		Running:     a.runRunning,
		Connecting:  a.runConnecting,
		UsingAPIKey: a.runUsingKey,
		Error:       a.runError,
	}
}

// GetRunStatus returns whether YouTube chat ingest is running.
func (a *App) GetRunStatus() RunStatus {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.runStatusLocked()
}

// Start begins YouTube chat ingest into the OBS chat socket, then fans each
// line to alerts or TTS on the overlay. Uses last saved settings (save the
// form first). Connect work runs in a goroutine so the UI is not blocked.
func (a *App) Start() error {
	a.Stop()

	a.mu.Lock()
	if err := a.settings.CanStart(); err != nil {
		a.runError = publicPipelineError(err)
		a.emitRunLocked()
		a.mu.Unlock()
		return err
	}
	if a.httpSrv == nil {
		err := fmt.Errorf("OBS HTTP is not listening")
		a.runError = err.Error()
		a.emitRunLocked()
		a.mu.Unlock()
		return err
	}
	settings := a.settings
	srv := a.httpSrv
	match, speak, err := a.overlayFor(settings, srv)
	if err != nil {
		a.runError = publicPipelineError(err)
		a.emitRunLocked()
		a.mu.Unlock()
		return err
	}
	newClient := a.factory()
	ctx, cancel := context.WithCancel(context.Background())
	a.runCancel = cancel
	a.runConnecting = true
	a.runRunning = false
	a.runUsingKey = settings.HasAPIKey()
	a.runError = ""
	a.runGen++
	gen := a.runGen
	wg := &sync.WaitGroup{}
	wg.Add(1)
	a.runWG = wg
	a.emitRunLocked()
	a.mu.Unlock()

	go a.ingestChat(ctx, gen, wg, settings, srv, newClient, match, speak)
	return nil
}

// Stop cancels the YouTube iterator. OBS HTTP stays up.
func (a *App) Stop() {
	a.mu.Lock()
	cancel, wg := a.stopRunLocked()
	a.runError = ""
	a.emitRunLocked()
	a.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	if wg != nil {
		wg.Wait()
	}
}

func (a *App) ingestChat(ctx context.Context, gen int, wg *sync.WaitGroup, settings config.Settings, srv *obs.Server, newClient clientFactory, match alertMatcher, speak synthesizer) {
	defer wg.Done()
	client, err := newClient(settings)
	if err != nil {
		if ctx.Err() != nil {
			a.finishRun(gen)
			return
		}
		a.failRun(gen, err)
		return
	}
	it, err := client.GetMessageIterator(ctx, settings.Youtube.StreamID)
	if err != nil {
		if ctx.Err() != nil {
			a.finishRun(gen)
			return
		}
		a.failRun(gen, err)
		return
	}
	a.mu.Lock()
	if gen != a.runGen {
		a.mu.Unlock()
		return
	}
	a.runConnecting = false
	a.runRunning = true
	a.emitRunLocked()
	a.mu.Unlock()

	for {
		select {
		case <-ctx.Done():
			a.finishRun(gen)
			return
		case msg, ok := <-it.GetChan():
			if !ok {
				a.finishRun(gen)
				return
			}
			dispatchChat(srv, match, speak, msg)
		}
	}
}

func (a *App) failRun(gen int, err error) {
	msg := publicPipelineError(err)
	slog.Error("youtube ingest failed", "err", msg)
	a.mu.Lock()
	defer a.mu.Unlock()
	if gen != a.runGen {
		return
	}
	a.runConnecting = false
	a.runRunning = false
	a.runError = msg
	a.emitRunLocked()
}

func (a *App) finishRun(gen int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if gen != a.runGen {
		return
	}
	a.runConnecting = false
	a.runRunning = false
	a.emitRunLocked()
}
