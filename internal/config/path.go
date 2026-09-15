package config

import (
	"os"
	"path/filepath"
)

// DefaultPath is os.UserConfigDir()/ytmemchat/config.json.
func DefaultPath() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, AppDirName, FileName), nil
}

// DirPath is the directory that contains config.json for path p.
func DirPath(configFile string) string {
	return filepath.Dir(configFile)
}
