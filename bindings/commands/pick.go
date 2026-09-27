//go:build !nomain

package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/HardDie/ytmemchat_wails/pkg/archive"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// PickAlertMediaFile opens a file dialog rooted at the media folder.
// The result is a path relative to that folder (subfolders allowed). Empty means cancel.
func (c *Commands) PickAlertMediaFile(mediaDir string) (string, error) {
	ctx := c.app.DialogContext()
	if ctx == nil {
		return "", fmt.Errorf("app not started")
	}
	mediaDir = strings.TrimSpace(mediaDir)
	if mediaDir == "" {
		mediaDir = strings.TrimSpace(c.app.MediaPath())
	}
	if mediaDir == "" {
		return "", fmt.Errorf("set a media folder in Configuration")
	}
	abs, err := filepath.Abs(mediaDir)
	if err != nil {
		return "", fmt.Errorf("media folder: %w", err)
	}
	chosen, err := runtime.OpenFileDialog(ctx, runtime.OpenDialogOptions{
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

// ExportAlertCommands saves a zip of the saved commands.yaml and every media
// file a command uses. The YAML entry is always named commands.yaml.
// An empty string means the operator cancelled the dialog.
func (c *Commands) ExportAlertCommands() (string, error) {
	ctx := c.app.DialogContext()
	if ctx == nil {
		return "", fmt.Errorf("app not started")
	}
	spec, cmds, err := exportSpec(c.app.CommandsPath(), c.app.MediaPath())
	if err != nil {
		return "", err
	}
	opts := runtime.SaveDialogOptions{
		Title:                "Export alert commands",
		DefaultFilename:      "commands.zip",
		CanCreateDirectories: true,
		Filters: []runtime.FileFilter{
			{DisplayName: "Zip archive", Pattern: "*.zip"},
		},
	}
	if abs, err := filepath.Abs(strings.TrimSpace(c.app.MediaPath())); err == nil {
		if st, err := os.Stat(abs); err == nil && st.IsDir() {
			opts.DefaultDirectory = abs
		}
	}
	chosen, err := runtime.SaveFileDialog(ctx, opts)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(chosen) == "" {
		return "", nil
	}
	spec.Dest = chosen
	if err := archive.Write(spec); err != nil {
		return "", mapExportError(cmds, err)
	}
	return chosen, nil
}

// ImportAlertCommands unpacks a zip that has commands.yaml at its root.
// The operator picks the zip, then the folder that receives every file.
// An empty string means the operator cancelled a dialog.
// On success the saved media folder becomes that folder and a custom commands path is cleared.
func (c *Commands) ImportAlertCommands() (string, error) {
	ctx := c.app.DialogContext()
	if ctx == nil {
		return "", fmt.Errorf("app not started")
	}
	zipPath, err := runtime.OpenFileDialog(ctx, runtime.OpenDialogOptions{
		Title: "Choose an alert commands archive",
		Filters: []runtime.FileFilter{
			{DisplayName: "Zip archive", Pattern: "*.zip"},
		},
	})
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(zipPath) == "" {
		return "", nil
	}
	if err := validateImportArchive(zipPath); err != nil {
		return "", err
	}
	dest, err := runtime.OpenDirectoryDialog(ctx, runtime.OpenDialogOptions{
		Title:                "Choose where to keep the imported files",
		CanCreateDirectories: true,
	})
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(dest) == "" {
		return "", nil
	}
	if err := extractImportArchive(zipPath, dest); err != nil {
		return "", err
	}
	if err := c.app.SetAlertsMediaPath(dest); err != nil {
		return "", err
	}
	return dest, nil
}
