package main

import (
	"context"
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"github.com/HardDie/ytmemchat_wails/internal/alerts"
	"github.com/HardDie/ytmemchat_wails/internal/config"
	"github.com/HardDie/ytmemchat_wails/internal/obs"
	"github.com/HardDie/ytmemchat_wails/internal/tts"
	"github.com/HardDie/ytmemchat_wails/internal/youtube"
)

func savePatched(t *testing.T, a *App, patch func(*SettingsForm)) {
	t.Helper()
	f := a.GetSettings()
	if patch != nil {
		patch(&f)
	}
	if err := a.SaveSettings(f); err != nil {
		t.Fatal(err)
	}
}

func TestGetSaveSettings_roundTrip(t *testing.T) {
	st := config.NewStore(filepath.Join(t.TempDir(), "config.json"))
	a := newAppWithStore(st)
	got := a.GetSettings()
	if got.Port != "8080" || got.StreamID != "" || got.APIKey != "" {
		t.Fatalf("%+v", got)
	}
	if !got.TTSEnabled || !got.AlertsEnabled || got.AlertsToken != "@" || got.WebhookEnabled {
		t.Fatalf("defaults %+v", got)
	}
	savePatched(t, a, func(f *SettingsForm) {
		f.StreamID = "  liveid  "
		f.APIKey = " secret "
		f.Port = ":9090"
		f.TTSEnabled = false
		f.TTSVoiceName = " Milena "
		f.AlertsEnabled = true
		f.AlertsToken = "#"
		f.AlertsMediaPath = " /tmp/media "
		f.AlertsCommandsFilePath = " /tmp/commands.yaml "
		f.WebhookEnabled = true
	})
	got = a.GetSettings()
	if got.StreamID != "liveid" || got.APIKey != "secret" || got.Port != ":9090" {
		t.Fatalf("after save %+v", got)
	}
	if got.TTSEnabled || got.TTSVoiceName != "Milena" {
		t.Fatalf("tts %+v", got)
	}
	if !got.AlertsEnabled || got.AlertsToken != "#" || got.AlertsMediaPath != "/tmp/media" || got.AlertsCommandsFilePath != "/tmp/commands.yaml" {
		t.Fatalf("alerts %+v", got)
	}
	if !got.WebhookEnabled {
		t.Fatal("webhook")
	}
	if a.ConfigPath() != st.Path() {
		t.Fatalf("path %q", a.ConfigPath())
	}
}

func TestSaveSettings_invalidPortLeavesMemory(t *testing.T) {
	st := config.NewStore(filepath.Join(t.TempDir(), "config.json"))
	a := newAppWithStore(st)
	savePatched(t, a, func(f *SettingsForm) {
		f.StreamID = "vid"
		f.Port = "8080"
	})
	f := a.GetSettings()
	f.Port = "nope"
	err := a.SaveSettings(f)
	if !errors.Is(err, config.ErrInvalidPort) {
		t.Fatalf("err = %v", err)
	}
	got := a.GetSettings()
	if got.Port != "8080" {
		t.Fatalf("port mutated to %q", got.Port)
	}
}

func TestSaveSettings_alertToken(t *testing.T) {
	st := config.NewStore(filepath.Join(t.TempDir(), "config.json"))
	a := newAppWithStore(st)
	f := a.GetSettings()
	f.AlertsEnabled = true
	f.AlertsToken = ""
	f.Port = "8080"
	if err := a.SaveSettings(f); !errors.Is(err, config.ErrAlertToken) {
		t.Fatalf("err = %v", err)
	}
}

func TestSaveSettings_emptyStreamIDAllowed(t *testing.T) {
	st := config.NewStore(filepath.Join(t.TempDir(), "config.json"))
	a := newAppWithStore(st)
	savePatched(t, a, func(f *SettingsForm) {
		f.StreamID = ""
		f.Port = "8080"
	})
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

func TestTTSVoicesFrom_mergesLanguages(t *testing.T) {
	got := ttsVoicesFrom([]tts.VoiceInfo{
		{Name: "Alex", Language: "en_US", Gender: "Male", Details: "hi"},
		{Name: "Alex", Language: "en_GB"},
		{Name: "", Language: "xx"},
	})
	if len(got) != 1 {
		t.Fatalf("%+v", got)
	}
	if got[0].Name != "Alex" || got[0].Gender != "Male" {
		t.Fatalf("%+v", got[0])
	}
	if !strings.Contains(got[0].Languages, "en_US") || !strings.Contains(got[0].Languages, "en_GB") {
		t.Fatalf("languages %q", got[0].Languages)
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
	if !strings.Contains(string(body), "location.pathname") {
		t.Fatalf("body %s", body)
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

func TestLookupLatestStream_usesSavedKeyAndFormID(t *testing.T) {
	st := config.NewStore(filepath.Join(t.TempDir(), "config.json"))
	a := newAppWithStore(st)
	savePatched(t, a, func(f *SettingsForm) {
		f.StreamID = "oldvid"
		f.APIKey = "secret"
	})
	a.lookupLatest = func(_ context.Context, key, vid string) (youtube.LatestBroadcast, error) {
		if key != "secret" || vid != "oldvid" {
			t.Fatalf("key=%q vid=%q", key, vid)
		}
		return youtube.LatestBroadcast{VideoID: "newlive", ChannelID: "UCabc", Kind: youtube.BroadcastLive}, nil
	}
	got, err := a.LookupLatestStream("", "")
	if err != nil {
		t.Fatal(err)
	}
	if got.StreamID != "newlive" || got.Kind != "live" || got.ChannelID != "UCabc" {
		t.Fatalf("%+v", got)
	}
	if a.GetSettings().StreamID != "oldvid" {
		t.Fatal("lookup must not save")
	}
}

func TestLookupLatestStream_prefersCallArgs(t *testing.T) {
	st := config.NewStore(filepath.Join(t.TempDir(), "config.json"))
	a := newAppWithStore(st)
	a.lookupLatest = func(_ context.Context, key, vid string) (youtube.LatestBroadcast, error) {
		if key != "formkey" || vid != "formvid" {
			t.Fatalf("key=%q vid=%q", key, vid)
		}
		return youtube.LatestBroadcast{VideoID: "soon", Kind: youtube.BroadcastUpcoming}, nil
	}
	got, err := a.LookupLatestStream(" formvid ", " formkey ")
	if err != nil {
		t.Fatal(err)
	}
	if got.StreamID != "soon" || got.Kind != "upcoming" {
		t.Fatalf("%+v", got)
	}
}

func TestLookupLatestStream_requiresIDAndKey(t *testing.T) {
	st := config.NewStore(filepath.Join(t.TempDir(), "config.json"))
	a := newAppWithStore(st)
	if _, err := a.LookupLatestStream("", "k"); err == nil || !strings.Contains(err.Error(), "stream ID") {
		t.Fatalf("id err = %v", err)
	}
	if _, err := a.LookupLatestStream("vid", ""); err == nil || !strings.Contains(err.Error(), "API key") {
		t.Fatalf("key err = %v", err)
	}
}

func TestGetSaveAlertCommands(t *testing.T) {
	st := config.NewStore(filepath.Join(t.TempDir(), "config.json"))
	a := newAppWithStore(st)
	if _, err := a.GetAlertCommands(); err == nil || !strings.Contains(err.Error(), "commands.yaml") {
		t.Fatalf("path err = %v", err)
	}
	yamlPath := filepath.Join(t.TempDir(), "commands.yaml")
	savePatched(t, a, func(f *SettingsForm) {
		f.AlertsCommandsFilePath = yamlPath
	})
	got, err := a.GetAlertCommands()
	if err != nil {
		t.Fatal(err)
	}
	if got.Path != yamlPath || len(got.Commands) != 0 {
		t.Fatalf("%+v", got)
	}
	vol := 0.5
	if err := a.SaveAlertCommands(AlertCommandsFile{Commands: []AlertCommand{
		{Name: "jump", File: "jump.mp3"},
		{Name: "dance", File: "cat.gif", Volume: &vol},
	}}); err != nil {
		t.Fatal(err)
	}
	loaded, err := alerts.LoadFile(yamlPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Commands) != 2 || loaded.Commands[0].Volume != nil || loaded.Commands[1].Volume == nil {
		t.Fatalf("%+v", loaded)
	}
	again, err := a.GetAlertCommands()
	if err != nil || len(again.Commands) != 2 {
		t.Fatalf("%+v %v", again, err)
	}
}

func TestAppVersion_defaultDev(t *testing.T) {
	a := newAppWithStore(config.NewStore(filepath.Join(t.TempDir(), "config.json")))
	if a.AppVersion() != "dev" {
		t.Fatalf("version = %q", a.AppVersion())
	}
}
