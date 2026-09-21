package configuration

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/HardDie/ytmemchat_wails/internal/config"
)

func TestGetSaveSettings_roundTrip(t *testing.T) {
	st := config.NewStore(filepath.Join(t.TempDir(), "config.json"))
	c := New(&configStub{settings: config.Defaults(), store: st})
	got := c.GetSettings()
	if got.Port != "8080" || got.StreamID != "" || got.APIKey != "" {
		t.Fatalf("%+v", got)
	}
	if got.TTSEnabled || got.AlertsEnabled || got.AlertsToken != "@" || got.WebhookEnabled {
		t.Fatalf("defaults %+v", got)
	}
	if !got.InterruptHotkeyEnabled || got.InterruptHotkeyChord != "Ctrl+Shift+I" {
		t.Fatalf("hotkey %+v", got)
	}
	savePatched(t, c, func(f *SettingsForm) {
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
		f.InterruptHotkeyEnabled = true
		f.InterruptHotkeyChord = " ctrl+alt+f8 "
	})
	got = c.GetSettings()
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
	if !got.InterruptHotkeyEnabled || got.InterruptHotkeyChord != "Ctrl+Alt+F8" {
		t.Fatalf("hotkey after save %+v", got)
	}
	if c.ConfigPath() != st.Path() {
		t.Fatalf("path %q", c.ConfigPath())
	}
}

func TestSaveSettings_rejectsBareInterruptKey(t *testing.T) {
	c := New(&configStub{settings: config.Defaults(), store: config.NewStore(filepath.Join(t.TempDir(), "config.json"))})
	err := c.SaveSettings(SettingsForm{Port: "8080", AlertsToken: "@", InterruptHotkeyEnabled: true, InterruptHotkeyChord: "I"})
	if !errors.Is(err, config.ErrInterruptHotkey) {
		t.Fatalf("err = %v", err)
	}
}

func TestSaveSettings_invalidPortLeavesMemory(t *testing.T) {
	st := config.NewStore(filepath.Join(t.TempDir(), "config.json"))
	c := New(&configStub{settings: config.Defaults(), store: st})
	savePatched(t, c, func(f *SettingsForm) {
		f.StreamID = "vid"
		f.Port = "8080"
	})
	f := c.GetSettings()
	f.Port = "nope"
	err := c.SaveSettings(f)
	if !errors.Is(err, config.ErrInvalidPort) {
		t.Fatalf("err = %v", err)
	}
	got := c.GetSettings()
	if got.Port != "8080" {
		t.Fatalf("port mutated to %q", got.Port)
	}
}

func TestSaveSettings_alertToken(t *testing.T) {
	c := New(&configStub{settings: config.Defaults(), store: config.NewStore(filepath.Join(t.TempDir(), "config.json"))})
	f := c.GetSettings()
	f.AlertsEnabled = true
	f.AlertsToken = ""
	f.Port = "8080"
	if err := c.SaveSettings(f); !errors.Is(err, config.ErrAlertToken) {
		t.Fatalf("err = %v", err)
	}
}

func TestSaveSettings_emptyStreamIDAllowed(t *testing.T) {
	c := New(&configStub{settings: config.Defaults(), store: config.NewStore(filepath.Join(t.TempDir(), "config.json"))})
	savePatched(t, c, func(f *SettingsForm) {
		f.StreamID = ""
		f.Port = "8080"
	})
}
