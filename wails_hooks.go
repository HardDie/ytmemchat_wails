//go:build !nomain

package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) setupWailsHooks() {
	a.emit = func(name string, data any) {
		if a.ctx == nil {
			return
		}
		runtime.EventsEmit(a.ctx, name, data)
	}
}

// PickCommandsFile opens a file dialog for commands.yaml. Empty means cancel.
func (a *App) PickCommandsFile() (string, error) {
	if a.ctx == nil {
		return "", fmt.Errorf("app not started")
	}
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Choose commands.yaml",
		Filters: []runtime.FileFilter{
			{DisplayName: "YAML (*.yaml, *.yml)", Pattern: "*.yaml;*.yml"},
		},
	})
}

// PickMediaDirectory opens a folder dialog for alert media. Empty means cancel.
func (a *App) PickMediaDirectory() (string, error) {
	if a.ctx == nil {
		return "", fmt.Errorf("app not started")
	}
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Choose alert media folder",
	})
}

// PickAlertMediaFile opens a file dialog rooted at the media folder.
// The result is a path relative to that folder (subfolders allowed). Empty means cancel.
func (a *App) PickAlertMediaFile(mediaDir string) (string, error) {
	if a.ctx == nil {
		return "", fmt.Errorf("app not started")
	}
	mediaDir = strings.TrimSpace(mediaDir)
	if mediaDir == "" {
		a.mu.Lock()
		mediaDir = strings.TrimSpace(a.settings.Alerts.MediaPath)
		a.mu.Unlock()
	}
	if mediaDir == "" {
		return "", fmt.Errorf("set a media folder in Configuration")
	}
	abs, err := filepath.Abs(mediaDir)
	if err != nil {
		return "", fmt.Errorf("media folder: %w", err)
	}
	chosen, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title:            "Choose alert media",
		DefaultDirectory: abs,
		Filters: []runtime.FileFilter{
			{DisplayName: "Alert media", Pattern: "*.gif;*.webm;*.mp4;*.mov;*.mp3;*.ogg;*.wav;*.aac;*.flac;*.png;*.jpg;*.jpeg;*.webp"},
		},
	})
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(chosen) == "" {
		return "", nil
	}
	return mediaRelativePath(abs, chosen)
}
