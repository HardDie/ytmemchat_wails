package configuration

import (
	"context"
	"fmt"
	"testing"

	"github.com/HardDie/ytmemchat_wails/internal/config"
)

type configStub struct {
	settings  config.Settings
	hotkeyErr string
	store     *config.Store
	storeErr  error
}

func (s *configStub) SettingsSnapshot() (config.Settings, string) {
	return s.settings, s.hotkeyErr
}

func (s *configStub) ConfigFilePath() string {
	if s.store == nil {
		return ""
	}
	return s.store.Path()
}

func (s *configStub) PersistSettings(next config.Settings) error {
	if s.store == nil {
		if s.storeErr != nil {
			return fmt.Errorf("config: %w", s.storeErr)
		}
		return fmt.Errorf("config: no store")
	}
	if err := s.store.Save(next); err != nil {
		return err
	}
	loaded, err := s.store.Load()
	if err != nil {
		return err
	}
	s.settings = loaded
	return nil
}

func (s *configStub) DialogContext() context.Context { return nil }

func savePatched(t *testing.T, c *Configuration, patch func(*SettingsForm)) {
	t.Helper()
	f := c.GetSettings()
	if patch != nil {
		patch(&f)
	}
	if err := c.SaveSettings(f); err != nil {
		t.Fatal(err)
	}
}
