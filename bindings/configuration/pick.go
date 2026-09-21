//go:build !nomain

package configuration

import (
	"fmt"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// PickCommandsFile opens a file dialog for commands.yaml. Empty means cancel.
func (c *Configuration) PickCommandsFile() (string, error) {
	ctx := c.app.DialogContext()
	if ctx == nil {
		return "", fmt.Errorf("app not started")
	}
	return runtime.OpenFileDialog(ctx, runtime.OpenDialogOptions{
		Title: "Choose commands.yaml",
		Filters: []runtime.FileFilter{
			{DisplayName: "YAML (*.yaml, *.yml)", Pattern: "*.yaml;*.yml"},
		},
	})
}

// PickMediaDirectory opens a folder dialog for alert media. Empty means cancel.
func (c *Configuration) PickMediaDirectory() (string, error) {
	ctx := c.app.DialogContext()
	if ctx == nil {
		return "", fmt.Errorf("app not started")
	}
	return runtime.OpenDirectoryDialog(ctx, runtime.OpenDialogOptions{
		Title: "Choose alert media folder",
	})
}
