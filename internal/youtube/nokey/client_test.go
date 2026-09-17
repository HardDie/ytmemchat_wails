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

func TestConvertChatItem_textMessage(t *testing.T) {
	item := mustChatItem(t, `{
		"liveChatTextMessageRenderer": {
			"id": "m1",
			"timestampUsec": "1760000000000000",
			"authorName": {"simpleText": "Sam"},
			"authorPhoto": {"thumbnails": [
				{"url": "https://img/32", "width": 32},
				{"url": "https://img/64", "width": 64}
			]},
			"message": {"runs": [{"text": "hello "}, {"text": "world"}]}
		}
	}`)
	msg := convertChatItem(item)
	if msg.ID != "m1" || msg.Author != "Sam" || msg.Message != "hello world" {
		t.Fatalf("%+v", msg)
	}
	if msg.ImgURL != "https://img/64" {
		t.Fatalf("ImgURL = %q", msg.ImgURL)
	}
	if msg.Type != "textMessageEvent" {
		t.Fatalf("Type = %q", msg.Type)
	}
	want := time.UnixMicro(1760000000000000).UTC()
	if !msg.Timestamp.Equal(want) {
		t.Fatalf("Timestamp = %s, want %s", msg.Timestamp, want)
	}
}

func TestConvertChatItem_avatarProtocolRelative(t *testing.T) {
	item := mustChatItem(t, `{
		"liveChatTextMessageRenderer": {
			"authorName": {"simpleText": "Sam"},
			"authorPhoto": {"thumbnails": [{"url": "//yt3.ggpht.com/a", "width": 64}]},
			"message": {"runs": [{"text": "hi"}]}
		}
	}`)
	msg := convertChatItem(item)
	if msg.ImgURL != "https://yt3.ggpht.com/a" {
		t.Fatalf("ImgURL = %q", msg.ImgURL)
	}
}

func TestConvertChatItem_emojiRuns(t *testing.T) {
	item := mustChatItem(t, `{
		"liveChatTextMessageRenderer": {
			"authorName": {"simpleText": "Sam"},
			"message": {"runs": [
				{"text": "nice "},
				{"emoji": {"shortcuts": [":pog:", ":poggers:"]}}
			]}
		}
	}`)
	msg := convertChatItem(item)
	if msg.Message != "nice :pog:" {
		t.Fatalf("Message = %q", msg.Message)
	}
}

func TestConvertChatItem_superChat(t *testing.T) {
	item := mustChatItem(t, `{
		"liveChatPaidMessageRenderer": {
			"id": "sc1",
			"timestampUsec": "1760000000000001",
			"authorName": {"simpleText": "Pat"},
			"authorPhoto": {"thumbnails": [{"url": "https://img/pat", "width": 88}]},
			"purchaseAmountText": {"simpleText": "$5.00"},
			"message": {"runs": [{"text": "wow"}]}
		}
	}`)
	msg := convertChatItem(item)
	if msg.ID != "sc1" || msg.Author != "Pat" || msg.ImgURL != "https://img/pat" {
		t.Fatalf("%+v", msg)
	}
	if msg.Type != "superChatEvent" || msg.Message != "[$5.00] wow" {
		t.Fatalf("type=%q message=%q", msg.Type, msg.Message)
	}
}

func TestConvertChatItem_superChatNoComment(t *testing.T) {
	item := mustChatItem(t, `{
		"liveChatPaidMessageRenderer": {
			"authorName": {"simpleText": "Pat"},
			"purchaseAmountText": {"simpleText": "¥1000"}
		}
	}`)
	msg := convertChatItem(item)
	if msg.Type != "superChatEvent" || msg.Message != "[¥1000]" {
		t.Fatalf("%+v", msg)
	}
}

func TestConvertChatItem_superSticker(t *testing.T) {
	item := mustChatItem(t, `{
		"liveChatPaidStickerRenderer": {
			"authorName": {"simpleText": "Pat"},
			"purchaseAmountText": {"simpleText": "$2.00"}
		}
	}`)
	msg := convertChatItem(item)
	if msg.Type != "superStickerEvent" || msg.Message != "[$2.00]" {
		t.Fatalf("%+v", msg)
	}
}

func TestConvertChatItem_empty(t *testing.T) {
	if convertChatItem(chatItem{}).Message != "" {
		t.Fatal("empty item should have no message")
	}
}

func mustChatItem(t *testing.T, raw string) chatItem {
	t.Helper()
	var item chatItem
	if err := json.Unmarshal([]byte(raw), &item); err != nil {
		t.Fatal(err)
	}
	return item
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
											"id":            "n1",
											"timestampUsec": "1760000000000000",
											"authorName":    map[string]any{"simpleText": "Pat"},
											"authorPhoto":   map[string]any{"thumbnails": []any{map[string]any{"url": "https://img/pat", "width": 64}}},
											"message":       map[string]any{"runs": []any{map[string]any{"text": "hi"}}},
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
	if !ok || msg.Author != "Pat" || msg.Message != "hi" || msg.ImgURL != "https://img/pat" || msg.ID != "n1" || msg.Type != "textMessageEvent" {
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
