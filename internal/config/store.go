package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

const fileMode = 0o600
const dirMode = 0o700

// Store reads and writes a Settings JSON file.
type Store struct {
	path string
}

// NewStore uses an explicit config.json path (tests, custom locations).
func NewStore(path string) *Store {
	return &Store{path: path}
}

// NewDefaultStore uses [DefaultPath].
func NewDefaultStore() (*Store, error) {
	p, err := DefaultPath()
	if err != nil {
		return nil, err
	}
	return NewStore(p), nil
}

// Path returns the JSON file path.
func (s *Store) Path() string {
	return s.path
}

// Load returns persisted settings. A missing file yields [Defaults] and no error.
func (s *Store) Load() (Settings, error) {
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

// Save writes settings atomically (temp file, sync, rename). Mode 0600.
func (s *Store) Save(cfg Settings) error {
	cfg = cfg.TrimSpace()
	if cfg.Version == 0 {
		cfg.Version = CurrentVersion
	}
	if err := cfg.Validate(); err != nil {
		return err
	}
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
