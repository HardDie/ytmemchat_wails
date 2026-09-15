package alerts

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// command is one trigger from commands.yaml.
type command struct {
	Name   string   `yaml:"name"`
	File   string   `yaml:"file"`
	Volume *float64 `yaml:"volume"`
	Scale  *float64 `yaml:"scale"`
}

type commandFile struct {
	Commands []command `yaml:"commands"`
}

func parseCommands(path string) (*commandFile, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("os.Open(): %w", err)
	}
	defer f.Close()

	var cmd commandFile
	if err := yaml.NewDecoder(f).Decode(&cmd); err != nil {
		return nil, fmt.Errorf("yaml.Decode(): %w", err)
	}
	return &cmd, nil
}
