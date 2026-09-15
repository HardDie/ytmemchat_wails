package alerts

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"gopkg.in/yaml.v3"
)

// Command is one trigger from commands.yaml.
type Command struct {
	Name   string   `yaml:"name" json:"name"`
	File   string   `yaml:"file" json:"file"`
	Volume *float64 `yaml:"volume,omitempty" json:"volume,omitempty"`
	Scale  *float64 `yaml:"scale,omitempty" json:"scale,omitempty"`
}

// File is the on-disk commands.yaml document.
type File struct {
	Commands []Command `yaml:"commands" json:"commands"`
}

func parseCommands(path string) (*File, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("alerts: read commands file: %w", err)
	}
	var cmd File
	if err := yaml.Unmarshal(raw, &cmd); err != nil {
		return nil, fmt.Errorf("alerts: decode commands file: %w", err)
	}
	if cmd.Commands == nil {
		cmd.Commands = []Command{}
	}
	return &cmd, nil
}

// LoadFile reads commands.yaml. A missing file is an error ([os.ErrNotExist]).
func LoadFile(path string) (File, error) {
	parsed, err := parseCommands(path)
	if err != nil {
		return File{}, err
	}
	return *parsed, nil
}

// SaveFile writes commands.yaml. Nil volume and scale are omitted.
// Duplicate names (case-insensitive) and empty name or file are rejected.
// Volume and scale, when set, must be non-negative with at most two decimal digits.
func SaveFile(path string, f File) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return fmt.Errorf("alerts: empty commands file path")
	}
	if f.Commands == nil {
		f.Commands = []Command{}
	}
	if err := validateCommands(f.Commands); err != nil {
		return err
	}
	data, err := yaml.Marshal(&f)
	if err != nil {
		return fmt.Errorf("alerts: encode commands file: %w", err)
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("alerts: mkdir: %w", err)
	}
	tmp, err := os.CreateTemp(dir, ".commands-*.tmp")
	if err != nil {
		return fmt.Errorf("alerts: temp: %w", err)
	}
	tmpName := tmp.Name()
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(tmpName)
		}
	}()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("alerts: write: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("alerts: close: %w", err)
	}
	if runtime.GOOS == "windows" {
		_ = os.Remove(path)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("alerts: replace: %w", err)
	}
	cleanup = false
	return nil
}

func validateCommands(list []Command) error {
	seen := make(map[string]struct{}, len(list))
	for i, it := range list {
		name := strings.TrimSpace(it.Name)
		file := strings.TrimSpace(it.File)
		if name == "" {
			return fmt.Errorf("alerts: command %d: empty name", i+1)
		}
		if file == "" {
			return fmt.Errorf("alerts: command %q: empty file", name)
		}
		key := strings.ToLower(name)
		if _, ok := seen[key]; ok {
			return fmt.Errorf("duplicate command: %s", name)
		}
		seen[key] = struct{}{}
		if err := validateDecimal(it.Volume, name, "volume"); err != nil {
			return err
		}
		if err := validateDecimal(it.Scale, name, "scale"); err != nil {
			return err
		}
	}
	return nil
}

func validateDecimal(v *float64, command, field string) error {
	if v == nil {
		return nil
	}
	if math.IsNaN(*v) || math.IsInf(*v, 0) || *v < 0 {
		return fmt.Errorf("alerts: command %q: %s must be a non-negative number with at most two decimal places", command, field)
	}
	scaled := *v * 100
	if math.Abs(scaled-math.Round(scaled)) > 1e-6 {
		return fmt.Errorf("alerts: command %q: %s allows at most two digits after the decimal point", command, field)
	}
	return nil
}
