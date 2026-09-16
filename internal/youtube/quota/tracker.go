package quota

import (
	"sync"
	"time"

	_ "time/tzdata"
)

// Default is the process-wide estimator used by the Data API v3 client.
var Default = NewTracker()

var pacific = mustPacific()

func mustPacific() *time.Location {
	loc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		return time.FixedZone("PST", -8*60*60)
	}
	return loc
}

// Snapshot is a copy of today’s local estimate (Pacific Time calendar day).
type Snapshot struct {
	// Units is spend in the default bucket (videos.list, liveChatMessages.list, unknown).
	Units int
	// Search is search.list calls in the separate daily search bucket.
	Search int
	// UnitsLimit is [DefaultUnitsPerDay] (documented default, not the Cloud project’s approved quota).
	UnitsLimit int
	// SearchLimit is [DefaultSearchPerDay].
	SearchLimit int
	// Day is the Pacific date the counters apply to (YYYY-MM-DD).
	Day string
}

// RemainingUnits is UnitsLimit minus Units, floored at zero.
func (s Snapshot) RemainingUnits() int {
	return remaining(s.UnitsLimit, s.Units)
}

// RemainingSearch is SearchLimit minus Search, floored at zero.
func (s Snapshot) RemainingSearch() int {
	return remaining(s.SearchLimit, s.Search)
}

func remaining(limit, used int) int {
	n := limit - used
	if n < 0 {
		return 0
	}
	return n
}

// Tracker is a thread-safe local unit counter. Zero value is not ready; use [NewTracker].
type Tracker struct {
	mu     sync.Mutex
	now    func() time.Time
	day    string
	units  int
	search int
}

// NewTracker returns an empty estimator for the current Pacific day.
func NewTracker() *Tracker {
	return &Tracker{now: time.Now}
}

// Record adds the published cost of m. Empty m is treated as [Unknown].
func (t *Tracker) Record(m Method) {
	if t == nil {
		return
	}
	if m == "" {
		m = Unknown
	}
	n := Cost(m)
	t.mu.Lock()
	defer t.mu.Unlock()
	t.rollLocked(t.now())
	if searchBucket(m) {
		t.search += n
		return
	}
	t.units += n
}

// Snapshot returns today’s totals. It may roll the Pacific day without recording.
func (t *Tracker) Snapshot() Snapshot {
	if t == nil {
		return emptySnapshot(time.Now())
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	now := t.now()
	t.rollLocked(now)
	return Snapshot{
		Units:       t.units,
		Search:      t.search,
		UnitsLimit:  DefaultUnitsPerDay,
		SearchLimit: DefaultSearchPerDay,
		Day:         t.day,
	}
}

func (t *Tracker) rollLocked(now time.Time) {
	day := pacificDay(now)
	if t.day == day {
		return
	}
	t.day = day
	t.units = 0
	t.search = 0
}

func pacificDay(now time.Time) string {
	return now.In(pacific).Format("2006-01-02")
}

func emptySnapshot(now time.Time) Snapshot {
	return Snapshot{
		UnitsLimit:  DefaultUnitsPerDay,
		SearchLimit: DefaultSearchPerDay,
		Day:         pacificDay(now),
	}
}
