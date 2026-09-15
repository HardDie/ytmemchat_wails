//go:build !nomain

package main

import (
	"fmt"

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
