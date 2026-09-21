package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/HardDie/ytmemchat_wails/internal/secret"
)

func TestSaveLoad_vaultOmitsJSONKey(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	v := &secret.Memory{}
	st := NewStoreWithVault(path, v)
	in := Defaults()
	in.Youtube.APIKey = " secret-key "
	if err := st.Save(in); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "secret-key") {
		t.Fatalf("key still in JSON: %s", raw)
	}
	out, err := st.Load()
	if err != nil {
		t.Fatal(err)
	}
	if out.Youtube.APIKey != "secret-key" {
		t.Fatalf("memory key %q", out.Youtube.APIKey)
	}
	if !out.APIKeyInKeychain {
		t.Fatal("want keychain hint")
	}
}

func TestLoad_migratesJSONKeyToVault(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	st := NewStore(path)
	in := Defaults()
	in.Youtube.APIKey = "from-file"
	if err := st.Save(in); err != nil {
		t.Fatal(err)
	}
	v := &secret.Memory{}
	got, err := NewStoreWithVault(path, v).Load()
	if err != nil {
		t.Fatal(err)
	}
	if got.Youtube.APIKey != "from-file" {
		t.Fatalf("key %q", got.Youtube.APIKey)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "from-file") {
		t.Fatalf("file not migrated: %s", raw)
	}
	if !got.APIKeyInKeychain {
		t.Fatal("want keychain hint")
	}
}

func TestLoad_vaultSetFailsLeavesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	st := NewStore(path)
	in := Defaults()
	in.Youtube.APIKey = "keep-me"
	if err := st.Save(in); err != nil {
		t.Fatal(err)
	}
	v := &secret.Memory{}
	v.FailSet(errors.New("set failed"))
	got, err := NewStoreWithVault(path, v).Load()
	if err != nil {
		t.Fatal(err)
	}
	if got.Youtube.APIKey != "keep-me" {
		t.Fatalf("key %q", got.Youtube.APIKey)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), "keep-me") {
		t.Fatal("file should keep the key")
	}
	if got.APIKeyInKeychain {
		t.Fatal("hint should be file")
	}
}

func TestSaveLoad_vaultDownKeepsJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	v := &secret.Memory{}
	v.Down()
	st := NewStoreWithVault(path, v)
	in := Defaults()
	in.Youtube.APIKey = "plain"
	if err := st.Save(in); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), "plain") {
		t.Fatal("want key in JSON")
	}
	out, err := st.Load()
	if err != nil {
		t.Fatal(err)
	}
	if out.Youtube.APIKey != "plain" || out.APIKeyInKeychain {
		t.Fatalf("%+v", out.Youtube)
	}
}

func TestSave_clearsVault(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	v := &secret.Memory{}
	st := NewStoreWithVault(path, v)
	in := Defaults()
	in.Youtube.APIKey = "gone"
	if err := st.Save(in); err != nil {
		t.Fatal(err)
	}
	in.Youtube.APIKey = ""
	if err := st.Save(in); err != nil {
		t.Fatal(err)
	}
	if _, err := v.Get(); !errors.Is(err, secret.ErrNotFound) {
		t.Fatalf("vault %v", err)
	}
}
