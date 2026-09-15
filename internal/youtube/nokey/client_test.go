package nokey

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestParseContinuationFromHTML(t *testing.T) {
	html := `<html><script>var x = {"liveChatRenderer":true,"continuation":"CONT_ABC"}</script></html>`
	tok, err := parseContinuationFromHTML(html)
	if err != nil {
		t.Fatal(err)
	}
	if tok != "CONT_ABC" {
		t.Fatalf("token = %q", tok)
	}
	_, err = parseContinuationFromHTML(`<html><script>nope</script></html>`)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestConvertToChatMessage(t *testing.T) {
	var r liveChatTextMessageRenderer
	r.AuthorName.SimpleText = "Sam"
	r.Message.Runs = []struct {
		Text string `json:"text"`
	}{{Text: "hello "}, {Text: "world"}}
	msg := convertToChatMessage(r)
	if msg.Author != "Sam" || msg.Message != "hello world" {
		t.Fatalf("%+v", msg)
	}
}

func TestGetMessageIterator_httptest(t *testing.T) {
	var posts int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "live_chat"):
			_, _ = io.WriteString(w, `<html><script>{"liveChatRenderer":1,"continuation":"T0"}</script></html>`)
		case r.Method == http.MethodPost:
			posts++
			w.Header().Set("Content-Type", "application/json")
			payload := map[string]any{
				"continuationContents": map[string]any{
					"liveChatContinuation": map[string]any{
						"continuations": []any{
							map[string]any{"timedContinuationData": map[string]any{"continuation": "T1", "timeoutMs": 1}},
						},
						"actions": []any{
							map[string]any{
								"addChatItemAction": map[string]any{
									"item": map[string]any{
										"liveChatTextMessageRenderer": map[string]any{
											"authorName": map[string]any{"simpleText": "Pat"},
											"message":    map[string]any{"runs": []any{map[string]any{"text": "hi"}}},
										},
									},
								},
							},
						},
					},
				},
			}
			_ = json.NewEncoder(w).Encode(payload)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	c := newClient(srv.Client(), srv.URL+"/live_chat?v=%s", srv.URL+"/get_live_chat", time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	it, err := c.GetMessageIterator(ctx, "abc")
	if err != nil {
		t.Fatal(err)
	}
	msg, ok := it.Next()
	if !ok || msg.Author != "Pat" || msg.Message != "hi" {
		t.Fatalf("msg=%+v ok=%v", msg, ok)
	}
	cancel()
}

func TestGetMessageIterator_missingContinuation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `<html><body>offline</body></html>`)
	}))
	t.Cleanup(srv.Close)
	c := newClient(srv.Client(), srv.URL+"/live_chat?v=%s", srv.URL+"/x", time.Millisecond)
	_, err := c.GetMessageIterator(context.Background(), "abc")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestApplyContinuation(t *testing.T) {
	it := &htmlIterator{pollingDelay: time.Second}
	var res ytInternalResponse
	res.ContinuationContents.LiveChatContinuation.Continuations = []struct {
		TimedContinuationData *struct {
			Continuation string `json:"continuation"`
			TimeoutMs    int64  `json:"timeoutMs"`
		} `json:"timedContinuationData"`
		InvalidationContinuationData *struct {
			Continuation string `json:"continuation"`
			TimeoutMs    int64  `json:"timeoutMs"`
		} `json:"invalidationContinuationData"`
	}{{}}
	res.ContinuationContents.LiveChatContinuation.Continuations[0].TimedContinuationData = &struct {
		Continuation string `json:"continuation"`
		TimeoutMs    int64  `json:"timeoutMs"`
	}{Continuation: "NEXT", TimeoutMs: 50}
	applyContinuation(it, res)
	if it.pageToken != "NEXT" || it.pollingDelay != 50*time.Millisecond {
		t.Fatalf("token=%s delay=%s", it.pageToken, it.pollingDelay)
	}
}
