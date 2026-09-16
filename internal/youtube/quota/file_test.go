package quota

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadWriteFile_roundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), FileName)
	in := Snapshot{Day: "2026-09-16", Units: 126, Search: 2}
	if err := WriteFile(path, in); err != nil {
		t.Fatal(err)
	}
	got, err := ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.Day != in.Day || got.Units != 126 || got.Search != 2 {
		t.Fatalf("%+v", got)
	}
	if got.UnitsLimit != DefaultUnitsPerDay {
		t.Fatalf("limits %+v", got)
	}
}

func TestReadFile_missing(t *testing.T) {
	got, err := ReadFile(filepath.Join(t.TempDir(), "nope.json"))
	if err != nil {
		t.Fatal(err)
	}
	if got.Units != 0 || got.Day != "" {
		t.Fatalf("%+v", got)
	}
}

func TestTracker_restoreAndPersist(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, FileName)
	tr := NewTracker()
	today := pacificDay(tr.now())
	tr.Restore(Snapshot{Day: today, Units: 100, Search: 3})
	if got := tr.Snapshot(); got.Units != 100 || got.Search != 3 {
		t.Fatalf("restore %+v", got)
	}
	var writes int
	tr.SetPersist(func(s Snapshot) {
		writes++
		if err := WriteFile(path, s); err != nil {
			t.Errorf("write: %v", err)
		}
	})
	tr.Record(VideosList)
	if writes != 1 {
		t.Fatalf("writes = %d", writes)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Fatal("empty file")
	}
	got, err := ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.Units != 101 || got.Search != 3 || got.Day != today {
		t.Fatalf("persisted %+v", got)
	}
	tr.Restore(Snapshot{Day: "2000-01-01", Units: 999, Search: 9})
	if got := tr.Snapshot(); got.Units != 0 || got.Day != today {
		t.Fatalf("stale restore %+v", got)
	}
}
