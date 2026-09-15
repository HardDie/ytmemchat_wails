package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

// mediaRelativePath returns chosen as a slash-separated path inside mediaDir.
// Files outside mediaDir (including parent folders) are rejected.
func mediaRelativePath(mediaDir, chosen string) (string, error) {
	root := strings.TrimSpace(mediaDir)
	file := strings.TrimSpace(chosen)
	if root == "" {
		return "", fmt.Errorf("set a media folder in Configuration")
	}
	if file == "" {
		return "", fmt.Errorf("no file selected")
	}
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("media folder: %w", err)
	}
	fileAbs, err := filepath.Abs(file)
	if err != nil {
		return "", fmt.Errorf("media file: %w", err)
	}
	rootAbs = filepath.Clean(rootAbs)
	fileAbs = filepath.Clean(fileAbs)
	sep := string(filepath.Separator)
	prefix := rootAbs + sep
	if fileAbs != rootAbs && !strings.HasPrefix(fileAbs, prefix) {
		return "", fmt.Errorf("file must be inside the media folder")
	}
	rel, err := filepath.Rel(rootAbs, fileAbs)
	if err != nil {
		return "", fmt.Errorf("media file: %w", err)
	}
	rel = filepath.Clean(rel)
	if rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+sep) {
		return "", fmt.Errorf("file must be inside the media folder")
	}
	return filepath.ToSlash(rel), nil
}
