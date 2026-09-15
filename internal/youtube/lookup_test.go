package youtube

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func lookupTestClient(t *testing.T, h http.HandlerFunc) *apiClient {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	c, err := newAPIClient(context.Background(), "k", srv.Client(), srv.URL+"/")
	if err != nil {
		t.Fatal(err)
	}
	return c.(*apiClient)
}

func TestLookupLatest_seedStillLive(t *testing.T) {
	var searched bool
	c := lookupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "search") {
			searched = true
			_, _ = io.WriteString(w, `{"items":[]}`)
			return
		}
		if strings.Contains(r.URL.Path, "videos") {
			_, _ = io.WriteString(w, `{"items":[{"snippet":{"channelId":"UCabc"},"liveStreamingDetails":{"activeLiveChatId":"chat1"}}]}`)
			return
		}
		http.NotFound(w, r)
	})
	got, err := c.lookupLatestBroadcast(context.Background(), "oldvid")
	if err != nil {
		t.Fatal(err)
	}
	if searched {
		t.Fatal("must not search when the seed video is still live")
	}
	if got.VideoID != "oldvid" || got.ChannelID != "UCabc" || got.Kind != BroadcastLive {
		t.Fatalf("%+v", got)
	}
}

func TestLookupLatest_prefersLiveOverUpcoming(t *testing.T) {
	c := lookupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.Contains(r.URL.Path, "search"):
			if r.URL.Query().Get("eventType") == "live" {
				_, _ = io.WriteString(w, `{"items":[{"id":{"videoId":"liveNow"}}]}`)
				return
			}
			_, _ = io.WriteString(w, `{"items":[{"id":{"videoId":"soon"}}]}`)
		case strings.Contains(r.URL.Path, "videos"):
			_, _ = io.WriteString(w, `{"items":[{"snippet":{"channelId":"UCabc"},"liveStreamingDetails":{"actualStartTime":"2020-01-01T00:00:00Z","actualEndTime":"2020-01-01T01:00:00Z"}}]}`)
		default:
			http.NotFound(w, r)
		}
	})
	got, err := c.lookupLatestBroadcast(context.Background(), "vod")
	if err != nil {
		t.Fatal(err)
	}
	if got.VideoID != "liveNow" || got.Kind != BroadcastLive {
		t.Fatalf("%+v", got)
	}
}

func TestLookupLatest_upcomingWhenNotLive(t *testing.T) {
	c := lookupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.Contains(r.URL.Path, "search"):
			if r.URL.Query().Get("eventType") == "upcoming" {
				_, _ = io.WriteString(w, `{"items":[{"id":{"videoId":"soon"}}]}`)
				return
			}
			_, _ = io.WriteString(w, `{"items":[]}`)
		case strings.Contains(r.URL.Path, "videos"):
			_, _ = io.WriteString(w, `{"items":[{"snippet":{"channelId":"UCabc"}}]}`)
		default:
			http.NotFound(w, r)
		}
	})
	got, err := c.lookupLatestBroadcast(context.Background(), "vod")
	if err != nil {
		t.Fatal(err)
	}
	if got.VideoID != "soon" || got.Kind != BroadcastUpcoming || got.ChannelID != "UCabc" {
		t.Fatalf("%+v", got)
	}
}

func TestLookupLatest_noBroadcast(t *testing.T) {
	c := lookupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "search") {
			_, _ = io.WriteString(w, `{"items":[]}`)
			return
		}
		if strings.Contains(r.URL.Path, "videos") {
			_, _ = io.WriteString(w, `{"items":[{"snippet":{"channelId":"UCabc"}}]}`)
			return
		}
		http.NotFound(w, r)
	})
	_, err := c.lookupLatestBroadcast(context.Background(), "vod")
	if !errors.Is(err, ErrNoBroadcast) {
		t.Fatalf("err = %v", err)
	}
}

func TestLookupLatest_unknownVideo(t *testing.T) {
	c := lookupTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"items":[]}`)
	})
	_, err := c.lookupLatestBroadcast(context.Background(), "missing")
	if !errors.Is(err, ErrUnknownVideo) {
		t.Fatalf("err = %v", err)
	}
}

func TestLookupLatest_invalidAPIKey(t *testing.T) {
	body, _ := json.Marshal(map[string]any{
		"error": map[string]any{
			"code":    400,
			"message": "API key not valid. Please pass a valid API key.",
			"errors": []map[string]any{
				{"reason": "keyInvalid", "message": "API key not valid. Please pass a valid API key."},
			},
		},
	})
	c := lookupTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write(body)
	})
	_, err := c.lookupLatestBroadcast(context.Background(), "vid")
	if !IsInvalidAPIKey(err) {
		t.Fatalf("err = %v", err)
	}
}

func TestLookupLatestBroadcast_emptyKey(t *testing.T) {
	_, err := LookupLatestBroadcast(context.Background(), "  ", "vid")
	if !errors.Is(err, ErrEmptyAPIKey) {
		t.Fatalf("err = %v", err)
	}
}
