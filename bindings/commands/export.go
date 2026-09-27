package commands

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/HardDie/ytmemchat_wails/internal/alerts"
	"github.com/HardDie/ytmemchat_wails/pkg/archive"
)

// exportCommandsName is the YAML entry at the root of an export zip.
// A custom commands path on disk still uses this name inside the archive.
const exportCommandsName = "commands.yaml"

// exportSpec checks the saved commands file and the media files it names.
func exportSpec(commandsPath, mediaDir string) (archive.Spec, []alerts.Command, error) {
	commandsPath = strings.TrimSpace(commandsPath)
	mediaDir = strings.TrimSpace(mediaDir)
	if commandsPath == "" {
		return archive.Spec{}, nil, fmt.Errorf("set a commands.yaml path in Configuration")
	}
	if mediaDir == "" {
		return archive.Spec{}, nil, fmt.Errorf("set a media folder in Configuration")
	}
	info, err := os.Stat(commandsPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return archive.Spec{}, nil, fmt.Errorf("save commands before export")
		}
		return archive.Spec{}, nil, fmt.Errorf("export: %w", err)
	}
	if info.IsDir() {
		return archive.Spec{}, nil, fmt.Errorf("commands path is a folder")
	}
	parsed, err := alerts.LoadFile(commandsPath)
	if err != nil {
		return archive.Spec{}, nil, err
	}
	spec := archive.Spec{
		Document:     commandsPath,
		DocumentName: exportCommandsName,
		Root:         mediaDir,
		Files:        commandFiles(parsed.Commands),
	}
	if err := archive.Validate(spec); err != nil {
		return archive.Spec{}, nil, mapExportError(parsed.Commands, err)
	}
	return spec, parsed.Commands, nil
}

func commandFiles(list []alerts.Command) []string {
	out := make([]string, 0, len(list))
	for _, cmd := range list {
		file := strings.TrimSpace(cmd.File)
		if file == "" {
			continue
		}
		out = append(out, file)
	}
	return out
}

func writeCommandsExport(dest, commandsPath, mediaDir string) error {
	spec, cmds, err := exportSpec(commandsPath, mediaDir)
	if err != nil {
		return err
	}
	spec.Dest = dest
	if err := archive.Write(spec); err != nil {
		return mapExportError(cmds, err)
	}
	return nil
}

func mapExportError(list []alerts.Command, err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, archive.ErrDocumentMissing) {
		return fmt.Errorf("save commands before export")
	}
	var pe *archive.PathError
	if !errors.As(err, &pe) {
		return err
	}
	msg := pe.Err
	switch {
	case errors.Is(err, archive.ErrOutside):
		msg = fmt.Errorf("file must be inside the media folder")
	case errors.Is(err, archive.ErrMissing):
		msg = fmt.Errorf("file is missing")
	case errors.Is(err, archive.ErrReserved):
		msg = fmt.Errorf("that name is reserved for commands.yaml")
	case errors.Is(err, archive.ErrNotFile):
		msg = fmt.Errorf("file is not a regular file")
	}
	label := pe.Name
	if label == "" {
		label = pe.Path
	}
	return commandFileError(commandName(list, pe.Path), label, msg)
}

func commandName(list []alerts.Command, path string) string {
	path = strings.TrimSpace(path)
	for _, cmd := range list {
		if strings.TrimSpace(cmd.File) == path {
			return cmd.Name
		}
	}
	return ""
}

func commandFileError(name, file string, err error) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("file %q: %w", file, err)
	}
	return fmt.Errorf("command %q: file %q: %w", name, file, err)
}
