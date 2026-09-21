//go:build !nomain

package commands

import (
	"fmt"
	"path/filepath"
	"strings"

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
