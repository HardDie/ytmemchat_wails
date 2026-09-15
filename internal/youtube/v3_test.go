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
	"time"

	yt "google.golang.org/api/youtube/v3"
)

func TestNew_emptyKey(t *testing.T) {
	_, err := New("  ")
	if !errors.Is(err, ErrEmptyAPIKey) {
		t.Fatalf("err = %v", err)
	}
}

func TestFormatMoneyAndConvert(t *testing.T) {
	if g := formatMoney(1_500_000); g != "1.50" {
		t.Fatalf("money = %q", g)
	}
	ts := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	msg := convertToChatMessage(&yt.LiveChatMessage{
		Id: "id1",
		Snippet: &yt.LiveChatMessageSnippet{
			Type:           "textMessageEvent",
			DisplayMessage: "hello",
			PublishedAt:    ts.Format(time.RFC3339),
		},
		AuthorDetails: &yt.LiveChatMessageAuthorDetails{
			DisplayName:     "Ada",
			ProfileImageUrl: "https://img",
		},
	})
	if msg.Author != "Ada" || msg.Message != "hello" || msg.ID != "id1" {
		t.Fatalf("%+v", msg)
	}
	super := convertToChatMessage(&yt.LiveChatMessage{
		Snippet: &yt.LiveChatMessageSnippet{
			Type:           "superChatEvent",
			DisplayMessage: "wow",
			SuperChatDetails: &yt.LiveChatSuperChatDetails{
				Currency:     "USD",
				AmountMicros: 2000000,
			},
		},
	})
	if !strings.Contains(super.Message, "USD") || !strings.Contains(super.Message, "2.00") {
		t.Fatalf("super = %q", super.Message)
	}
	if convertToChatMessage(nil).Message != "" {
		t.Fatal("nil message")
	}
}

func TestMapAPIError(t *testing.T) {
	if !IsInvalidAPIKey(mapAPIError(errors.New("googleapi: Error 400: API key not valid. Please pass a valid API key., keyInvalid"))) {
		t.Fatal("expected invalid key from message")
	}
	if IsInvalidAPIKey(mapAPIError(fmtErr("quotaExceeded"))) {
		t.Fatal("quota is not invalid key")
	}
}

func fmtErr(s string) error { return errors.New(s) }

func TestGetMessageIterator_notLive(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "videos") {
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"items":[]}`)
			return
		}
		http.NotFound(w, r)
	}))
	t.Cleanup(srv.Close)
	c, err := newAPIClient(context.Background(), "test-key", srv.Client(), srv.URL+"/")
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.GetMessageIterator(context.Background(), "vid")
	if !errors.Is(err, ErrNotLive) {
		t.Fatalf("err = %v", err)
	}
}

func TestGetMessageIterator_invalidAPIKey(t *testing.T) {
	body, _ := json.Marshal(map[string]any{
		"error": map[string]any{
			"code":    400,
			"message": "API key not valid. Please pass a valid API key.",
			"errors": []map[string]any{
				{"reason": "keyInvalid", "message": "API key not valid. Please pass a valid API key."},
			},
		},
	})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)
	c, err := newAPIClient(context.Background(), "bad-key", srv.Client(), srv.URL+"/")
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.GetMessageIterator(context.Background(), "vid")
	if !IsInvalidAPIKey(err) {
		t.Fatalf("err = %v, want ErrInvalidAPIKey", err)
	}
}

func TestGetMessageIterator_happySkipHistory(t *testing.T) {
	var chatLists int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.Contains(r.URL.Path, "videos"):
			_, _ = io.WriteString(w, `{"items":[{"liveStreamingDetails":{"activeLiveChatId":"chat1"}}]}`)
		case strings.Contains(r.URL.Path, "liveChat/messages") || strings.Contains(r.URL.Path, "liveChatMessages"):
			chatLists++
			if chatLists == 1 {
				_, _ = io.WriteString(w, `{"nextPageToken":"p2","pollingIntervalMillis":1,"items":[{"id":"old","snippet":{"type":"textMessageEvent","displayMessage":"history"}}]}`)
				return
			}
			_, _ = io.WriteString(w, `{"nextPageToken":"p3","pollingIntervalMillis":1,"items":[{"id":"n1","snippet":{"type":"textMessageEvent","displayMessage":"live","publishedAt":"2026-01-01T00:00:00Z"},"authorDetails":{"displayName":"Cam","profileImageUrl":"https://a"}}]}`)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	c, err := newAPIClient(context.Background(), "k", srv.Client(), srv.URL+"/")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	it, err := c.GetMessageIterator(ctx, "vid")
	if err != nil {
		t.Fatal(err)
	}
	msg, ok := it.Next()
	if !ok {
		t.Fatal("expected a live message")
	}
	if msg.Message == "history" {
		t.Fatal("history must be skipped")
	}
	if msg.Message != "live" || msg.Author != "Cam" {
		t.Fatalf("%+v", msg)
	}
	cancel()
}

func TestGetChan_select(t *testing.T) {
	ch := make(chan *ChatMessage, 1)
	it := &apiIterator{ctx: context.Background(), messageChan: ch}
	ch <- &ChatMessage{Message: "x"}
	select {
	case m := <-it.GetChan():
		if m.Message != "x" {
			t.Fatalf("%+v", m)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}
}
