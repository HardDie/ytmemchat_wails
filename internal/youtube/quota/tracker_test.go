package quota

import (
	"sync"
	"testing"
	"time"
)

func TestTracker_recordAndSnapshot(t *testing.T) {
	tr := NewTracker()
	tr.Record(VideosList)
	tr.Record(LiveChatMessagesList)
	tr.Record(SearchList)
	tr.Record("")
	got := tr.Snapshot()
	if got.Units != 3 {
		t.Fatalf("units = %d, want 3 (videos + chat + unknown empty)", got.Units)
	}
	if got.Search != 1 {
		t.Fatalf("search = %d, want 1", got.Search)
	}
	if got.UnitsLimit != DefaultUnitsPerDay || got.SearchLimit != DefaultSearchPerDay {
		t.Fatalf("limits = %+v", got)
	}
	if got.Day == "" {
		t.Fatal("empty day")
	}
	if got.RemainingUnits() != DefaultUnitsPerDay-3 {
		t.Fatalf("remaining units = %d", got.RemainingUnits())
	}
}

func TestTracker_nilSafe(t *testing.T) {
	var tr *Tracker
	tr.Record(VideosList)
	tr.Reset()
	got := tr.Snapshot()
	if got.Units != 0 || got.UnitsLimit != DefaultUnitsPerDay {
		t.Fatalf("%+v", got)
	}
}

func TestTracker_resetsAtPacificMidnight(t *testing.T) {
	start := time.Date(2026, 9, 16, 23, 59, 0, 0, pacific)
	var mu sync.Mutex
	now := start
	tr := &Tracker{now: func() time.Time {
		mu.Lock()
		defer mu.Unlock()
		return now
	}}
	tr.Record(VideosList)
	tr.Record(SearchList)
	before := tr.Snapshot()
	if before.Units != 1 || before.Search != 1 || before.Day != "2026-09-16" {
		t.Fatalf("before = %+v", before)
	}
	mu.Lock()
	now = start.Add(2 * time.Minute)
	mu.Unlock()
	after := tr.Snapshot()
	if after.Units != 0 || after.Search != 0 || after.Day != "2026-09-17" {
		t.Fatalf("after midnight snapshot = %+v", after)
	}
	tr.Record(LiveChatMessagesList)
	got := tr.Snapshot()
	if got.Units != 1 || got.Search != 0 {
		t.Fatalf("next day record = %+v", got)
	}
}

func TestTracker_reset(t *testing.T) {
	tr := NewTracker()
	tr.Record(VideosList)
	tr.Record(SearchList)
	tr.Reset()
	got := tr.Snapshot()
	if got.Units != 0 || got.Search != 0 {
		t.Fatalf("after reset %+v", got)
	}
	if got.Day == "" {
		t.Fatal("empty day")
	}
	tr.Record(LiveChatMessagesList)
	if tr.Snapshot().Units != 1 {
		t.Fatalf("record after reset %+v", tr.Snapshot())
	}
}

func TestTracker_concurrentRecord(t *testing.T) {
	tr := NewTracker()
	var wg sync.WaitGroup
	const n = 32
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			tr.Record(VideosList)
			tr.Record(SearchList)
		}()
	}
	wg.Wait()
	got := tr.Snapshot()
	if got.Units != n || got.Search != n {
		t.Fatalf("got %+v, want %d/%d", got, n, n)
	}
}
