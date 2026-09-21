package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/HardDie/ytmemchat_wails/internal/secret"
)

const fileMode = 0o600
const dirMode = 0o700

// Store reads and writes a Settings JSON file and the optional API-key vault.
type Store struct {
	path  string
	vault secret.Vault
}

// NewStore uses an explicit config.json path (tests, custom locations).
// The YouTube API key stays in JSON (no OS keychain).
func NewStore(path string) *Store {
	return &Store{path: path}
}

// NewStoreWithVault uses path plus a [secret.Vault] for the API key.
func NewStoreWithVault(path string, v secret.Vault) *Store {
	return &Store{path: path, vault: v}
}

// NewDefaultStore uses [DefaultPath] and the OS keychain.
func NewDefaultStore() (*Store, error) {
	p, err := DefaultPath()
	if err != nil {
		return nil, err
	}
	return NewStoreWithVault(p, secret.OS()), nil
}

// Path returns the JSON file path.
func (s *Store) Path() string {
	return s.path
}

func (s *Store) keyVault() secret.Vault {
	if s.vault == nil {
		return secret.Unavailable{}
	}
	return s.vault
}

// Load returns persisted settings. A missing file yields [Defaults] and no error.
// When the vault works, a key still in JSON is moved into the vault and the file
// is rewritten with an empty apiKey. If the vault fails, the file is left as-is.
func (s *Store) Load() (Settings, error) {
	cfg, err := s.readFile()
	if err != nil {
		return Settings{}, err
	}
	return s.hydrate(cfg)
}

func (s *Store) readFile() (Settings, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return Defaults(), nil
		}
		return Settings{}, fmt.Errorf("config: read: %w", err)
	}
	cfg := Defaults()
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Settings{}, fmt.Errorf("config: invalid JSON: %w", err)
	}
	if cfg.Version == 0 {
		cfg.Version = CurrentVersion
	}
	cfg = cfg.TrimSpace()
	if cfg.Server.Port == "" {
		cfg.Server.Port = Defaults().Server.Port
	}
	return cfg, nil
}

func (s *Store) hydrate(cfg Settings) (Settings, error) {
	v := s.keyVault()
	if !v.Available() {
		cfg.APIKeyInKeychain = false
		return cfg, nil
	}
	cfg.APIKeyInKeychain = true
	if cfg.HasAPIKey() {
		if err := v.Set(cfg.Youtube.APIKey); err != nil {
			cfg.APIKeyInKeychain = false
			return cfg, nil
		}
		disk := cfg
		disk.Youtube.APIKey = ""
		if err := s.writeFile(disk); err != nil {
			return cfg, nil
		}
		return cfg, nil
	}
	key, err := v.Get()
	if err != nil {
		if !errors.Is(err, secret.ErrNotFound) {
			cfg.APIKeyInKeychain = false
		}
		return cfg, nil
	}
	cfg.Youtube.APIKey = key
	return cfg, nil
}

// Save writes settings atomically (temp file, sync, rename). Mode 0600.
// When the vault works, the API key is stored there and omitted from JSON.
func (s *Store) Save(cfg Settings) error {
	cfg = cfg.TrimSpace()
	if cfg.Version == 0 {
		cfg.Version = CurrentVersion
	}
	if err := cfg.Validate(); err != nil {
		return err
	}
	disk := cfg
	v := s.keyVault()
	if v.Available() {
		if cfg.HasAPIKey() {
			if err := v.Set(cfg.Youtube.APIKey); err == nil {
				disk.Youtube.APIKey = ""
			}
		} else {
			_ = v.Delete()
			disk.Youtube.APIKey = ""
		}
	}
	return s.writeFile(disk)
}

func (s *Store) writeFile(cfg Settings) error {
	cfg.APIKeyInKeychain = false
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, dirMode); err != nil {
		return fmt.Errorf("config: mkdir: %w", err)
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("config: encode: %w", err)
	}
	data = append(data, '\n')
	tmp, err := os.CreateTemp(dir, ".config-*.tmp")
	if err != nil {
		return fmt.Errorf("config: temp: %w", err)
	}
	tmpName := tmp.Name()
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(tmpName)
		}
	}()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("config: write: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("config: sync: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("config: close: %w", err)
	}
	if err := os.Chmod(tmpName, fileMode); err != nil {
		return fmt.Errorf("config: chmod: %w", err)
	}
	if err := replaceFile(tmpName, s.path); err != nil {
		return fmt.Errorf("config: replace: %w", err)
	}
	cleanup = false
	return nil
}

func replaceFile(tmp, dest string) error {
	if runtime.GOOS == "windows" {
		_ = os.Remove(dest)
	}
	return os.Rename(tmp, dest)
}
