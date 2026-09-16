package quota

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

const (
	// FileName is the daily estimate next to config.json.
	FileName = "quota.json"
	fileMode = 0o600
	dirMode  = 0o700
)

type fileDoc struct {
	Day    string `json:"day"`
	Units  int    `json:"units"`
	Search int    `json:"search"`
}

// ReadFile loads a previously saved snapshot. A missing file yields a zero Snapshot.
func ReadFile(path string) (Snapshot, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Snapshot{}, nil
		}
		return Snapshot{}, fmt.Errorf("quota: read: %w", err)
	}
	var doc fileDoc
	if err := json.Unmarshal(data, &doc); err != nil {
		return Snapshot{}, fmt.Errorf("quota: invalid JSON: %w", err)
	}
	return Snapshot{
		Day:         doc.Day,
		Units:       doc.Units,
		Search:      doc.Search,
		UnitsLimit:  DefaultUnitsPerDay,
		SearchLimit: DefaultSearchPerDay,
	}, nil
}

// WriteFile writes s atomically (temp file, sync, rename). Mode 0600.
func WriteFile(path string, s Snapshot) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, dirMode); err != nil {
		return fmt.Errorf("quota: mkdir: %w", err)
	}
	data, err := json.MarshalIndent(fileDoc{Day: s.Day, Units: s.Units, Search: s.Search}, "", "  ")
	if err != nil {
		return fmt.Errorf("quota: encode: %w", err)
	}
	data = append(data, '\n')
	tmp, err := os.CreateTemp(dir, ".quota-*.tmp")
	if err != nil {
		return fmt.Errorf("quota: temp: %w", err)
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
		return fmt.Errorf("quota: write: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("quota: sync: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("quota: close: %w", err)
	}
	if err := os.Chmod(tmpName, fileMode); err != nil {
		return fmt.Errorf("quota: chmod: %w", err)
	}
	if err := replaceFile(tmpName, path); err != nil {
		return fmt.Errorf("quota: replace: %w", err)
	}
	cleanup = false
	return nil
}

func replaceFile(tmp, dest string) error {
	if runtime.GOOS == "windows" {
		_ = os.Remove(dest)
	}
	return os.Rename(tmp, dest)
}
