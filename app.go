package main

import (
	"context"
	"log/slog"
	"path/filepath"
	"sync"
	"sync/atomic"

	"github.com/HardDie/ytmemchat_wails/bindings/configuration"
	"github.com/HardDie/ytmemchat_wails/bindings/home"
	"github.com/HardDie/ytmemchat_wails/internal/config"
	"github.com/HardDie/ytmemchat_wails/internal/hotkey"
	"github.com/HardDie/ytmemchat_wails/internal/obs"
	"github.com/HardDie/ytmemchat_wails/internal/youtube"
	"github.com/HardDie/ytmemchat_wails/internal/youtube/quota"
)

type (
	SettingsForm = configuration.SettingsForm
	RunStatus    = home.RunStatus
	OBSStatus    = home.OBSStatus
)

// App is the desktop core (settings, OBS HTTP, YouTube chat Start/Stop).
// Pane bindings in bindings/ own Wails methods that are unique to one pane.
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
	emit           func(string, any)
	lookupLatest   func(context.Context, string, string) (youtube.LatestBroadcast, error)
	hk             hotkey.Binding
	hotkeyErr      string
}

// NewApp returns the application core with the default config store.
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
	a.setupQuotaLocked()
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

func (a *App) setupQuotaLocked() {
	if a.store == nil {
		return
	}
	path := filepath.Join(filepath.Dir(a.store.Path()), quota.FileName)
	if snap, err := quota.ReadFile(path); err != nil {
		slog.Error("quota load failed", "err", err)
	} else {
		quota.Default.Restore(snap)
	}
	quota.Default.SetPersist(func(s quota.Snapshot) {
		if err := quota.WriteFile(path, s); err != nil {
			slog.Error("quota persist failed", "err", err)
		}
	})
}
