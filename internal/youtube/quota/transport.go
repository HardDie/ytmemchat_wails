package quota

import (
	"net/http"
	"strings"
)

// WrapClient returns a shallow clone of c whose Transport records YouTube Data
// API v3 responses on t. c is not mutated. A nil c is treated as a new
// http.Client. A nil t leaves the client unwrapped (caller should pass [Default]).
func WrapClient(c *http.Client, t *Tracker) *http.Client {
	if t == nil {
		if c == nil {
			return &http.Client{}
		}
		return c
	}
	if c == nil {
		c = &http.Client{}
	}
	clone := *c
	base := clone.Transport
	if base == nil {
		base = http.DefaultTransport
	}
	clone.Transport = &roundTripper{next: base, tracker: t}
	return &clone
}

type roundTripper struct {
	next    http.RoundTripper
	tracker *Tracker
}

// RoundTrip implements [http.RoundTripper].
func (rt *roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := rt.next.RoundTrip(req)
	if resp != nil {
		if m, ok := MethodFromRequest(req); ok {
			rt.tracker.Record(m)
		}
	}
	return resp, err
}

// MethodFromRequest classifies a Data API v3 URL. Non-v3 requests return false.
func MethodFromRequest(req *http.Request) (Method, bool) {
	if req == nil || req.URL == nil {
		return "", false
	}
	resource, ok := youtubeV3Resource(req.URL.Path)
	if !ok {
		return "", false
	}
	switch {
	case resource == "liveChat/messages" || strings.HasPrefix(resource, "liveChat/messages/"):
		return LiveChatMessagesList, true
	case resource == "search" || strings.HasPrefix(resource, "search/"):
		return SearchList, true
	case resource == "videos" || strings.HasPrefix(resource, "videos/"):
		return VideosList, true
	default:
		return Unknown, true
	}
}

func youtubeV3Resource(path string) (string, bool) {
	if strings.Contains(path, "/discovery/") {
		return "", false
	}
	const marker = "/youtube/v3/"
	i := strings.Index(path, marker)
	if i < 0 {
		return "", false
	}
	rest := strings.Trim(path[i+len(marker):], "/")
	if rest == "" {
		return "", false
	}
	return rest, true
}
