package configuration

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/HardDie/ytmemchat_wails/internal/config"
	"github.com/HardDie/ytmemchat_wails/internal/secret"
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
	if got.APIKeyInKeychain {
		t.Fatal("file store is not keychain")
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
		f.AlertsCommandsFileCustom = true
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
	if !got.AlertsEnabled || got.AlertsToken != "#" || got.AlertsMediaPath != "/tmp/media" || got.AlertsCommandsFilePath != "/tmp/commands.yaml" || !got.AlertsCommandsFileCustom {
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

func TestGetSettings_legacyCommandsPath(t *testing.T) {
	s := config.Defaults()
	s.Alerts.MediaPath = "/tmp/media"
	s.Alerts.CommandsFilePath = filepath.Join("/tmp/media", "commands.yaml")
	got := New(&configStub{settings: s}).GetSettings()
	if got.AlertsCommandsFileCustom || got.AlertsCommandsFilePath != filepath.Join("/tmp/media", "commands.yaml") {
		t.Fatalf("legacy default %+v", got)
	}
	s.Alerts.CommandsFilePath = "/tmp/other.yaml"
	got = New(&configStub{settings: s}).GetSettings()
	if !got.AlertsCommandsFileCustom || got.AlertsCommandsFilePath != "/tmp/other.yaml" {
		t.Fatalf("other path %+v", got)
	}
}

func TestSaveSettings_commandsPathMode(t *testing.T) {
	c := New(&configStub{settings: config.Defaults(), store: config.NewStore(filepath.Join(t.TempDir(), "config.json"))})
	savePatched(t, c, func(f *SettingsForm) {
		f.Port = "8080"
		f.AlertsMediaPath = "/tmp/media"
		f.AlertsCommandsFilePath = "/tmp/other.yaml"
		f.AlertsCommandsFileCustom = false
	})
	got := c.GetSettings()
	if got.AlertsCommandsFileCustom || got.AlertsCommandsFilePath != filepath.Join("/tmp/media", "commands.yaml") {
		t.Fatalf("default mode %+v", got)
	}
	savePatched(t, c, func(f *SettingsForm) {
		f.AlertsCommandsFilePath = "/tmp/other.yaml"
		f.AlertsCommandsFileCustom = true
	})
	got = c.GetSettings()
	if !got.AlertsCommandsFileCustom || got.AlertsCommandsFilePath != "/tmp/other.yaml" {
		t.Fatalf("custom mode %+v", got)
	}
	savePatched(t, c, func(f *SettingsForm) {
		f.AlertsMediaPath = "/tmp/media"
		f.AlertsCommandsFilePath = filepath.Join("/tmp/media", "commands.yaml")
		f.AlertsCommandsFileCustom = false
	})
	got = c.GetSettings()
	if got.AlertsCommandsFileCustom {
		t.Fatalf("legacy default file %+v", got)
	}
	savePatched(t, c, func(f *SettingsForm) {
		f.AlertsMediaPath = "/tmp/media"
		f.AlertsCommandsFilePath = "/tmp/custom.yaml"
		f.AlertsCommandsFileCustom = true
	})
	got = c.GetSettings()
	if !got.AlertsCommandsFileCustom || got.AlertsCommandsFilePath != "/tmp/custom.yaml" {
		t.Fatalf("other path %+v", got)
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

func TestGetSettings_keychainHint(t *testing.T) {
	st := config.NewStoreWithVault(filepath.Join(t.TempDir(), "config.json"), &secret.Memory{})
	c := New(&configStub{settings: config.Defaults(), store: st})
	savePatched(t, c, func(f *SettingsForm) {
		f.APIKey = "vaulted"
		f.Port = "8080"
	})
	got := c.GetSettings()
	if got.APIKey != "vaulted" || !got.APIKeyInKeychain {
		t.Fatalf("%+v", got)
	}
}
