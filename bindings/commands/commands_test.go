package commands

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/HardDie/ytmemchat_wails/internal/alerts"
	"github.com/HardDie/ytmemchat_wails/internal/obs"
)

type cmdStub struct {
	path     string
	media    string
	skipHTTP bool
	overlay  *obs.Server
	reloaded bool
}

func (s *cmdStub) CommandsPath() string { return s.path }

func (s *cmdStub) MediaPath() string { return s.media }

func (s *cmdStub) SkipHTTP() bool { return s.skipHTTP }

func (s *cmdStub) ReloadOverlay() { s.reloaded = true }

func (s *cmdStub) OverlayServer() *obs.Server { return s.overlay }

func (s *cmdStub) DialogContext() context.Context { return nil }

func TestGetSaveAlertCommands(t *testing.T) {
	c := New(&cmdStub{})
	if _, err := c.GetAlertCommands(); err == nil || !strings.Contains(err.Error(), "commands.yaml") {
		t.Fatalf("path err = %v", err)
	}
	yamlPath := filepath.Join(t.TempDir(), "commands.yaml")
	st := &cmdStub{path: yamlPath, skipHTTP: true}
	c = New(st)
	got, err := c.GetAlertCommands()
	if err != nil {
		t.Fatal(err)
	}
	if got.Path != yamlPath || len(got.Commands) != 0 {
		t.Fatalf("%+v", got)
	}
	vol := 0.5
	if err := c.SaveAlertCommands(AlertCommandsFile{Commands: []AlertCommand{
		{Name: "jump", File: "jump.mp3"},
		{Name: "dance", File: "cat.gif", Volume: &vol},
	}}); err != nil {
		t.Fatal(err)
	}
	if st.reloaded {
		t.Fatal("skipHTTP should not reload overlay")
	}
	loaded, err := alerts.LoadFile(yamlPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Commands) != 2 || loaded.Commands[0].Volume != nil || loaded.Commands[1].Volume == nil {
		t.Fatalf("%+v", loaded)
	}
	again, err := c.GetAlertCommands()
	if err != nil || len(again.Commands) != 2 {
		t.Fatalf("%+v %v", again, err)
	}
}

func TestSaveAlertCommands_reloadsOverlay(t *testing.T) {
	yamlPath := filepath.Join(t.TempDir(), "commands.yaml")
	if err := os.WriteFile(yamlPath, []byte("commands: []\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	st := &cmdStub{path: yamlPath}
	if err := New(st).SaveAlertCommands(AlertCommandsFile{Commands: []AlertCommand{{Name: "a", File: "a.mp3"}}}); err != nil {
		t.Fatal(err)
	}
	if !st.reloaded {
		t.Fatal("expected overlay reload")
	}
}
