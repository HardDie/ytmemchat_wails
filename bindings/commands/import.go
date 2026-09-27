package commands

import (
	"errors"
	"fmt"

	"gopkg.in/yaml.v3"

	"github.com/HardDie/ytmemchat_wails/internal/alerts"
	"github.com/HardDie/ytmemchat_wails/pkg/archive"
)

func validateImportArchive(zipPath string) error {
	raw, err := archive.ReadRootFile(zipPath, exportCommandsName)
	if err != nil {
		return mapImportError(err)
	}
	var doc alerts.File
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return fmt.Errorf("commands.yaml: %w", err)
	}
	return nil
}

func extractImportArchive(zipPath, dest string) error {
	if err := validateImportArchive(zipPath); err != nil {
		return err
	}
	if err := archive.Extract(zipPath, dest, exportCommandsName); err != nil {
		return mapImportError(err)
	}
	return nil
}

func mapImportError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, archive.ErrRootDocument) {
		return fmt.Errorf("commands.yaml must be at the root of the archive")
	}
	if errors.Is(err, archive.ErrOutside) {
		return fmt.Errorf("archive path must stay inside the chosen folder")
	}
	return err
}
