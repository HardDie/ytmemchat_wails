package quota

import "testing"

func TestCost_knownMethodsAreOne(t *testing.T) {
	for _, m := range []Method{VideosList, LiveChatMessagesList, SearchList, Unknown, Method("channels.list")} {
		if g := Cost(m); g != 1 {
			t.Fatalf("Cost(%q) = %d, want 1", m, g)
		}
	}
}

func TestSearchBucket(t *testing.T) {
	if !searchBucket(SearchList) {
		t.Fatal("search.list must use the search bucket")
	}
	if searchBucket(VideosList) || searchBucket(LiveChatMessagesList) || searchBucket(Unknown) {
		t.Fatal("non-search methods must use the default unit bucket")
	}
}

func TestSnapshotRemaining_floorsAtZero(t *testing.T) {
	s := Snapshot{Units: 12_000, UnitsLimit: DefaultUnitsPerDay, Search: 150, SearchLimit: DefaultSearchPerDay}
	if s.RemainingUnits() != 0 {
		t.Fatalf("units remaining = %d", s.RemainingUnits())
	}
	if s.RemainingSearch() != 0 {
		t.Fatalf("search remaining = %d", s.RemainingSearch())
	}
	s.Units, s.Search = 10, 4
	if s.RemainingUnits() != DefaultUnitsPerDay-10 {
		t.Fatalf("units remaining = %d", s.RemainingUnits())
	}
	if s.RemainingSearch() != DefaultSearchPerDay-4 {
		t.Fatalf("search remaining = %d", s.RemainingSearch())
	}
}
