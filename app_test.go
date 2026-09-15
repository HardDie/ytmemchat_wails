package main

import (
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"github.com/HardDie/ytmemchat_wails/internal/config"
)

func TestGetSaveSettings_roundTrip(t *testing.T) {
	st := config.NewStore(filepath.Join(t.TempDir(), "config.json"))
	a := newAppWithStore(st)
	got := a.GetSettings()
	if got.Port != "8080" || got.StreamID != "" || got.APIKey != "" {
		t.Fatalf("%+v", got)
	}
	if err := a.SaveSettings(SettingsForm{
		StreamID: "  liveid  ",
		APIKey:   " secret ",
		Port:     ":9090",
	}); err != nil {
		t.Fatal(err)
	}
	got = a.GetSettings()
	if got.StreamID != "liveid" || got.APIKey != "secret" || got.Port != ":9090" {
		t.Fatalf("after save %+v", got)
	}
	if a.ConfigPath() != st.Path() {
		t.Fatalf("path %q", a.ConfigPath())
	}
}

func TestSaveSettings_invalidPortLeavesMemory(t *testing.T) {
	st := config.NewStore(filepath.Join(t.TempDir(), "config.json"))
	a := newAppWithStore(st)
	if err := a.SaveSettings(SettingsForm{StreamID: "vid", Port: "8080"}); err != nil {
		t.Fatal(err)
	}
	err := a.SaveSettings(SettingsForm{StreamID: "vid", Port: "nope"})
	if !errors.Is(err, config.ErrInvalidPort) {
		t.Fatalf("err = %v", err)
	}
	got := a.GetSettings()
	if got.Port != "8080" {
		t.Fatalf("port mutated to %q", got.Port)
	}
}

func TestSaveSettings_emptyStreamIDAllowed(t *testing.T) {
	st := config.NewStore(filepath.Join(t.TempDir(), "config.json"))
	a := newAppWithStore(st)
	if err := a.SaveSettings(SettingsForm{Port: "8080"}); err != nil {
		t.Fatal(err)
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
