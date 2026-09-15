package alerts

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSaveFile_omitsUnsetVolumeAndScale(t *testing.T) {
	path := filepath.Join(t.TempDir(), "commands.yaml")
	err := SaveFile(path, File{Commands: []Command{
		{Name: "jump", File: "jump.mp3"},
		{Name: "dance", File: "cat.gif", Volume: ptr(0.5), Scale: ptr(1.2)},
	}})
	if err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	if strings.Count(text, "volume:") != 1 || strings.Count(text, "scale:") != 1 {
		t.Fatalf("optional keys: %s", text)
	}
	if !strings.Contains(text, "jump.mp3") || !strings.Contains(text, "name: jump") {
		t.Fatalf("jump missing: %s", text)
	}
	got, err := LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Commands) != 2 {
		t.Fatalf("%+v", got)
	}
	if got.Commands[0].Volume != nil || got.Commands[0].Scale != nil {
		t.Fatalf("jump should omit volume/scale: %+v", got.Commands[0])
	}
	if got.Commands[1].Volume == nil || *got.Commands[1].Volume != 0.5 {
		t.Fatalf("dance volume %+v", got.Commands[1])
	}
}

func TestSaveFile_rejectsDuplicateAndEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "commands.yaml")
	if err := SaveFile(path, File{Commands: []Command{{Name: "jump", File: ""}}}); err == nil {
		t.Fatal("empty file")
	}
	if err := SaveFile(path, File{Commands: []Command{
		{Name: "Jump", File: "a.mp3"},
		{Name: "jump", File: "b.mp3"},
	}}); err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("dup err = %v", err)
	}
}

func TestLoadFile_missing(t *testing.T) {
	_, err := LoadFile(filepath.Join(t.TempDir(), "nope.yaml"))
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("err = %v", err)
	}
}

func TestSaveFile_rejectsExtraDecimals(t *testing.T) {
	path := filepath.Join(t.TempDir(), "commands.yaml")
	if err := SaveFile(path, File{Commands: []Command{{Name: "jump", File: "a.mp3", Volume: ptr(1.234)}}}); err == nil || !strings.Contains(err.Error(), "two digits") {
		t.Fatalf("volume err = %v", err)
	}
	if err := SaveFile(path, File{Commands: []Command{{Name: "jump", File: "a.mp3", Scale: ptr(-1)}}}); err == nil {
		t.Fatal("negative scale")
	}
	if err := SaveFile(path, File{Commands: []Command{{Name: "jump", File: "a.mp3", Volume: ptr(0.55), Scale: ptr(12.5)}}}); err != nil {
		t.Fatal(err)
	}
}

func ptr(v float64) *float64 { return &v }
