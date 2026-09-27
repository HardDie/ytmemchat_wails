package update

import (
	"strconv"
	"strings"
)

// IsReleaseTag reports a vMAJOR.MINOR.PATCH (optional v prefix) version.
func IsReleaseTag(s string) bool {
	_, _, _, ok := parseSemver(s)
	return ok
}

// Newer reports whether latest is a greater release tag than current.
// Non-tags (dev, commits) are never “newer”.
func Newer(latest, current string) bool {
	lm, ln, lp, okL := parseSemver(latest)
	cm, cn, cp, okC := parseSemver(current)
	if !okL || !okC {
		return false
	}
	if lm != cm {
		return lm > cm
	}
	if ln != cn {
		return ln > cn
	}
	return lp > cp
}

// SameTag reports equal release tags (v prefix optional).
func SameTag(a, b string) bool {
	am, an, ap, okA := parseSemver(a)
	bm, bn, bp, okB := parseSemver(b)
	return okA && okB && am == bm && an == bn && ap == bp
}

func parseSemver(s string) (maj, min, pat int, ok bool) {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "v")
	s = strings.TrimPrefix(s, "V")
	if i := strings.IndexAny(s, "-+"); i >= 0 {
		s = s[:i]
	}
	parts := strings.Split(s, ".")
	if len(parts) != 3 {
		return 0, 0, 0, false
	}
	maj, err1 := strconv.Atoi(parts[0])
	min, err2 := strconv.Atoi(parts[1])
	pat, err3 := strconv.Atoi(parts[2])
	if err1 != nil || err2 != nil || err3 != nil {
		return 0, 0, 0, false
	}
	if maj < 0 || min < 0 || pat < 0 {
		return 0, 0, 0, false
	}
	return maj, min, pat, true
}
