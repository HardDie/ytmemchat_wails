package main

import (
	"errors"
	"path/filepath"
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
