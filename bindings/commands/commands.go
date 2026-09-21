// Package commands is the Wails binding for the Commands pane.
package commands

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/HardDie/ytmemchat_wails/internal/alerts"
	"github.com/HardDie/ytmemchat_wails/internal/obs"
)

// ErrOBSNotListening is returned when overlay HTTP is down.
var ErrOBSNotListening = fmt.Errorf("OBS HTTP is not listening")

// ErrAlertFileEmpty is returned when PreviewAlert has no filename.
var ErrAlertFileEmpty = fmt.Errorf("command file is empty")

// Commands exposes commands.yaml editing and overlay preview.
type Commands struct {
	app api
}

// api is the desktop App core this pane uses.
type api interface {
	CommandsPath() string
	MediaPath() string
	SkipHTTP() bool
	ReloadOverlay()
	OverlayServer() *obs.Server
	DialogContext() context.Context
}

// New wraps the application core for the Commands pane.
func New(app api) *Commands {
	return &Commands{app: app}
}

// GetAlertCommands loads the saved commands.yaml. A missing file yields an empty list.
func (c *Commands) GetAlertCommands() (AlertCommandsFile, error) {
	path := strings.TrimSpace(c.app.CommandsPath())
	if path == "" {
		return AlertCommandsFile{}, fmt.Errorf("set a commands.yaml path in Configuration")
	}
	parsed, err := alerts.LoadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return AlertCommandsFile{Path: path, Commands: []AlertCommand{}}, nil
		}
		return AlertCommandsFile{}, err
	}
	return AlertCommandsFile{Path: path, Commands: toAlertCommands(parsed.Commands)}, nil
}

// SaveAlertCommands writes commands.yaml. Empty volume and scale are omitted.
func (c *Commands) SaveAlertCommands(in AlertCommandsFile) error {
	path := strings.TrimSpace(c.app.CommandsPath())
	if path == "" {
		return fmt.Errorf("set a commands.yaml path in Configuration")
	}
	if err := alerts.SaveFile(path, alerts.File{Commands: fromAlertCommands(in.Commands)}); err != nil {
		return err
	}
	if c.app.SkipHTTP() {
		return nil
	}
	c.app.ReloadOverlay()
	return nil
}

// PreviewAlert plays one command's media on the OBS overlay using the editor
// file, volume, and scale. Empty volume or scale from the caller should be 1.
func (c *Commands) PreviewAlert(filename string, volume, scale float64) error {
	filename = strings.TrimSpace(filename)
	if filename == "" {
		return ErrAlertFileEmpty
	}
	srv := c.app.OverlayServer()
	if srv == nil {
		return ErrOBSNotListening
	}
	srv.PublishOverlay(obs.AlertOverlay(filename, volume, scale))
	return nil
}

func toAlertCommands(in []alerts.Command) []AlertCommand {
	out := make([]AlertCommand, len(in))
	for i, cmd := range in {
		out[i] = AlertCommand{Name: cmd.Name, File: cmd.File, Volume: cmd.Volume, Scale: cmd.Scale}
	}
	return out
}

func fromAlertCommands(in []AlertCommand) []alerts.Command {
	out := make([]alerts.Command, len(in))
	for i, cmd := range in {
		out[i] = alerts.Command{Name: strings.TrimSpace(cmd.Name), File: strings.TrimSpace(cmd.File), Volume: cmd.Volume, Scale: cmd.Scale}
	}
	return out
}
