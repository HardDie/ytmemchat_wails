package main

import (
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"github.com/HardDie/ytmemchat_wails/bindings/configuration"
	"github.com/HardDie/ytmemchat_wails/internal/config"
	"github.com/HardDie/ytmemchat_wails/internal/obs"
	"github.com/HardDie/ytmemchat_wails/internal/youtube/quota"
)

func savePatched(t *testing.T, a *App, patch func(*SettingsForm)) {
	t.Helper()
	c := configuration.New(a)
	f := c.GetSettings()
	if patch != nil {
		patch(&f)
	}
	if err := c.SaveSettings(f); err != nil {
		t.Fatal(err)
	}
}

func TestSaveSettings_webhookRestartsHTTP(t *testing.T) {
	st := config.NewStore(filepath.Join(t.TempDir(), "config.json"))
	a := newAppWithStore(st)
	a.listenOverride = "127.0.0.1:0"
	a.skipHTTP = false
	a.mu.Lock()
	if err := a.startHTTPLocked(); err != nil {
		a.mu.Unlock()
		t.Fatal(err)
	}
	a.mu.Unlock()
	t.Cleanup(func() {
		a.mu.Lock()
		a.stopHTTPLocked()
		a.mu.Unlock()
	})
	base := strings.TrimSuffix(a.GetOBSStatus().ChatURL, obs.PathChat)
	res, err := http.Post(base+obs.PathWebhook, "application/json", strings.NewReader(`{"message":"x"}`))
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("webhook off status %d", res.StatusCode)
	}
	savePatched(t, a, func(f *SettingsForm) {
		f.WebhookEnabled = true
		f.TTSEnabled = false
		f.AlertsEnabled = false
	})
	base = strings.TrimSuffix(a.GetOBSStatus().ChatURL, obs.PathChat)
	res, err = http.Post(base+obs.PathWebhook, "application/json", strings.NewReader(`{"message":"x"}`))
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("webhook on status %d", res.StatusCode)
	}
}

func TestOBSHTTP_servesChatPage(t *testing.T) {
	st := config.NewStore(filepath.Join(t.TempDir(), "config.json"))
	a := newAppWithStore(st)
	a.listenOverride = "127.0.0.1:0"
	a.skipHTTP = false
	a.mu.Lock()
	if err := a.startHTTPLocked(); err != nil {
		a.mu.Unlock()
		t.Fatal(err)
	}
	a.mu.Unlock()
	t.Cleanup(func() {
		a.mu.Lock()
		a.stopHTTPLocked()
		a.mu.Unlock()
	})
	stt := a.GetOBSStatus()
	if !stt.Listening || stt.ChatURL == "" || stt.OverlayURL == "" || stt.IndexURL == "" {
		t.Fatalf("%+v", stt)
	}
	res, err := http.Get(stt.ChatURL)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status %d", res.StatusCode)
	}
	if !strings.Contains(string(body), obs.PathScript) {
		t.Fatalf("chat page must load %s", obs.PathScript)
	}
	base := strings.TrimSuffix(stt.ChatURL, obs.PathChat)
	jsRes, err := http.Get(base + obs.PathScript)
	if err != nil {
		t.Fatal(err)
	}
	jsBody, _ := io.ReadAll(jsRes.Body)
	jsRes.Body.Close()
	if jsRes.StatusCode != http.StatusOK {
		t.Fatalf("script status %d", jsRes.StatusCode)
	}
	if !strings.Contains(string(jsBody), "location.pathname") {
		t.Fatal("shared script must derive websocket from location")
	}
	res, err = http.Get(stt.IndexURL)
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("index %d", res.StatusCode)
	}
	res, err = http.Get(stt.OverlayURL)
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("overlay %d", res.StatusCode)
	}
}

func TestGetRunStatus_includesQuotaSnapshot(t *testing.T) {
	quota.Default.Reset()
	t.Cleanup(quota.Default.Reset)
	a := newAppWithStore(config.NewStore(filepath.Join(t.TempDir(), "config.json")))
	quota.Default.Record(quota.VideosList)
	quota.Default.Record(quota.SearchList)
	st := a.GetRunStatus()
	if st.QuotaUnits != 1 || st.QuotaSearch != 1 {
		t.Fatalf("spend %+v", st)
	}
	if st.QuotaUnitsLimit != quota.DefaultUnitsPerDay || st.QuotaSearchLimit != quota.DefaultSearchPerDay {
		t.Fatalf("limits %+v", st)
	}
}

func TestSetupQuota_restoresFile(t *testing.T) {
	quota.Default.Reset()
	t.Cleanup(func() {
		quota.Default.SetPersist(nil)
		quota.Default.Reset()
	})
	dir := t.TempDir()
	a := newAppWithStore(config.NewStore(filepath.Join(dir, "config.json")))
	day := quota.Default.Snapshot().Day
	if err := quota.WriteFile(filepath.Join(dir, quota.FileName), quota.Snapshot{Day: day, Units: 126, Search: 2}); err != nil {
		t.Fatal(err)
	}
	a.setupQuotaLocked()
	got := quota.Default.Snapshot()
	if got.Units != 126 || got.Search != 2 {
		t.Fatalf("restored %+v", got)
	}
	quota.Default.Record(quota.VideosList)
	saved, err := quota.ReadFile(filepath.Join(dir, quota.FileName))
	if err != nil {
		t.Fatal(err)
	}
	if saved.Units != 127 {
		t.Fatalf("persisted %+v", saved)
	}
}
