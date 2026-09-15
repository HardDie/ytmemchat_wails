package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaults(t *testing.T) {
	d := Defaults()
	if d.Version != CurrentVersion {
		t.Fatalf("version = %d", d.Version)
	}
	if d.Server.Port != "8080" {
		t.Fatalf("port = %q", d.Server.Port)
	}
	if d.Alerts.Token != "@" || !d.Alerts.Enabled {
		t.Fatalf("alerts = %+v", d.Alerts)
	}
	if d.Youtube.APIKey != "" || d.Youtube.StreamID != "" {
		t.Fatal("youtube fields must be empty on first launch")
	}
	if d.HasAPIKey() {
		t.Fatal("HasAPIKey on defaults")
	}
	if err := d.CanStart(); !errors.Is(err, ErrStreamIDRequired) {
		t.Fatalf("CanStart = %v", err)
	}
}

func TestHasAPIKey_trims(t *testing.T) {
	s := Defaults()
	s.Youtube.APIKey = "  "
	if s.HasAPIKey() {
		t.Fatal("whitespace is not a key")
	}
	s.Youtube.APIKey = "abc"
	if !s.HasAPIKey() {
		t.Fatal("expected key")
	}
}

func TestValidateAndParsePort(t *testing.T) {
	s := Defaults()
	s.Youtube.StreamID = "vid"
	if err := s.CanStart(); err != nil {
		t.Fatal(err)
	}
	s.Server.Port = ":9090"
	if s.Server.ListenAddr() != ":9090" {
		t.Fatalf("ListenAddr = %q", s.Server.ListenAddr())
	}
	n, err := ParsePort(":9090")
	if err != nil || n != 9090 {
		t.Fatalf("ParsePort = %d %v", n, err)
	}
	s.Server.Port = "not-a-port"
	if err := s.Validate(); !errors.Is(err, ErrInvalidPort) {
		t.Fatalf("Validate = %v", err)
	}
	s.Server.Port = "8080"
	s.Alerts.Token = "@@"
	if err := s.Validate(); !errors.Is(err, ErrAlertToken) {
		t.Fatalf("token = %v", err)
	}
	s.Alerts.Enabled = false
	s.Alerts.Token = ""
	if err := s.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestLoad_missingFile_defaults(t *testing.T) {
	dir := t.TempDir()
	st := NewStore(filepath.Join(dir, "config.json"))
	got, err := st.Load()
	if err != nil {
		t.Fatal(err)
	}
	if got.Server.Port != "8080" || got.Alerts.Token != "@" {
		t.Fatalf("got %+v", got)
	}
}

func TestSaveLoad_roundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "config.json")
	st := NewStore(path)
	in := Defaults()
	in.Youtube.StreamID = "  liveid  "
	in.Youtube.APIKey = " secret-key "
	in.TTS.VoiceName = "Milena"
	in.Webhook.Enabled = true
	if err := st.Save(in); err != nil {
		t.Fatal(err)
	}
	out, err := st.Load()
	if err != nil {
		t.Fatal(err)
	}
	if out.Youtube.StreamID != "liveid" {
		t.Fatalf("stream = %q", out.Youtube.StreamID)
	}
	if out.Youtube.APIKey != "secret-key" {
		t.Fatal("api key not preserved after trim")
	}
	if out.TTS.VoiceName != "Milena" || !out.Webhook.Enabled {
		t.Fatalf("out %+v", out)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "panic") {
		t.Fatal("unexpected content")
	}
	if err := st.Save(out); err != nil {
		t.Fatal(err)
	}
}

func TestLoad_unknownFieldsIgnored(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, []byte(`{"version":1,"future":true,"server":{"port":"7777"}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := NewStore(path).Load()
	if err != nil {
		t.Fatal(err)
	}
	if got.Server.Port != "7777" {
		t.Fatalf("port = %q", got.Server.Port)
	}
	if got.Alerts.Token != "@" {
		t.Fatalf("defaults not merged: token %q", got.Alerts.Token)
	}
}

func TestLoad_invalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, []byte(`{`), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := NewStore(path).Load()
	if err == nil {
		t.Fatal("expected JSON error")
	}
}

func TestSave_rejectsBadPort(t *testing.T) {
	st := NewStore(filepath.Join(t.TempDir(), "config.json"))
	s := Defaults()
	s.Server.Port = "0"
	if err := st.Save(s); !errors.Is(err, ErrInvalidPort) {
		t.Fatalf("err = %v", err)
	}
}

func TestErrorMessages_omitAPIKey(t *testing.T) {
	s := Defaults()
	s.Youtube.APIKey = "super-secret-key-value"
	s.Server.Port = "bad"
	err := s.Validate()
	if err == nil {
		t.Fatal("expected error")
	}
	if strings.Contains(err.Error(), "super-secret-key-value") {
		t.Fatalf("error leaked API key: %v", err)
	}
}

func TestJSON_doesNotRequireEnv(t *testing.T) {
	var s Settings
	if err := json.Unmarshal([]byte(`{}`), &s); err != nil {
		t.Fatal(err)
	}
}
