package obs

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func startTest(t *testing.T, cfg Config) (*Server, *httptest.Server) {
	t.Helper()
	s := New(cfg)
	ts := httptest.NewServer(s.Handler())
	t.Cleanup(func() {
		ts.Close()
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = s.Shutdown(ctx)
	})
	return s, ts
}

func TestIndexAndOBSPages(t *testing.T) {
	_, ts := startTest(t, Config{})
	res, err := http.Get(ts.URL + PathIndex)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status %d", res.StatusCode)
	}
	html := string(body)
	if !strings.Contains(html, PathChat) || !strings.Contains(html, PathOverlay) {
		t.Fatalf("index missing OBS paths: %s", html)
	}

	chat, err := http.Get(ts.URL + PathChat)
	if err != nil {
		t.Fatal(err)
	}
	cb, _ := io.ReadAll(chat.Body)
	chat.Body.Close()
	if !strings.Contains(string(cb), "location.pathname") {
		t.Fatal("chat html must derive websocket from location")
	}
	if strings.Contains(string(cb), "/ws_chat") {
		t.Fatal("console chat socket path leaked")
	}
	if !strings.Contains(string(cb), "app_closed") {
		t.Fatal("chat html must handle graceful app_closed")
	}
	if !strings.Contains(string(cb), "chat_flush") {
		t.Fatal("chat html must handle chat_flush")
	}
	if strings.Contains(string(cb), "location.reload") {
		t.Fatal("chat html must not reload on unhandled errors")
	}
	if !strings.Contains(string(cb), "recoverFromUnhandledError") {
		t.Fatal("chat html must recover via connectWebSocket")
	}
	if !strings.Contains(string(cb), "DEFAULT_CHAT_CAP") {
		t.Fatal("chat html must cap messages by default")
	}
	if !strings.Contains(string(cb), `get('cap')`) {
		t.Fatal("chat html must read cap from the query string")
	}
	if !strings.Contains(string(cb), "trimChatMessages") {
		t.Fatal("chat html must trim old rows")
	}
	if !strings.Contains(string(cb), "avatar.onerror") {
		t.Fatal("chat html must fall back when an avatar URL fails")
	}
	if !strings.Contains(string(cb), "if (!ctx)") {
		t.Fatal("chat html must not throw when canvas 2d is missing")
	}

	ov, err := http.Get(ts.URL + PathOverlay)
	if err != nil {
		t.Fatal(err)
	}
	ob, _ := io.ReadAll(ov.Body)
	ov.Body.Close()
	if !strings.Contains(string(ob), "/obs/media/") {
		t.Fatal("overlay must use /obs/media/")
	}
	if strings.Contains(string(ob), `'/media/`) || strings.Contains(string(ob), `"/media/`) {
		t.Fatal("console media path leaked")
	}
	if !strings.Contains(string(ob), "app_closed") {
		t.Fatal("overlay html must handle graceful app_closed")
	}
	if !strings.Contains(string(ob), "clearOverlayPlayback") {
		t.Fatal("overlay html must tear down media on app_closed")
	}
	if strings.Contains(string(ob), "location.reload") {
		t.Fatal("overlay html must not reload on unhandled errors")
	}
	if !strings.Contains(string(ob), "recoverFromUnhandledError") {
		t.Fatal("overlay html must recover via connectWebSocket")
	}
	if !strings.Contains(string(ob), "MIN_STICKER_PLAYING_SEC") {
		t.Fatal("overlay html must keep video on screen at least min sticker time")
	}
	if strings.Contains(string(ob), "mediaElement.onended") {
		t.Fatal("overlay video must not hide on ended")
	}
	if !strings.Contains(string(ob), "!laidOut") {
		t.Fatal("overlay html must drop media that never loads")
	}
	htmlOv := string(ob)
	if strings.Contains(htmlOv, "mp4|webm|mov|gif") {
		t.Fatal("overlay gif alerts must use img, not video")
	}
	if !strings.Contains(htmlOv, `filename.match(/\.(mp4|webm|mov)$/i)`) {
		t.Fatal("overlay video match must be mp4 webm mov")
	}
	intFn := overlayJSFunc(htmlOv, "interruptCurrentTTS")
	if intFn == "" {
		t.Fatal("overlay html must define interruptCurrentTTS")
	}
	if strings.Contains(intFn, "alertQueue = []") {
		t.Fatal("tts_interrupt must not flush the TTS queue")
	}
	clearFn := overlayJSFunc(htmlOv, "clearOverlayPlayback")
	if !strings.Contains(clearFn, "alertQueue = []") {
		t.Fatal("overlay teardown must dump the TTS queue")
	}
	if !strings.Contains(intFn, "bumpTTSGeneration") {
		t.Fatal("tts_interrupt must invalidate in-flight TTS decode")
	}
	if !strings.Contains(clearFn, "bumpTTSGeneration") {
		t.Fatal("overlay teardown must invalidate in-flight TTS decode")
	}
	if !strings.Contains(htmlOv, "generation !== ttsGeneration") {
		t.Fatal("overlay TTS decode must ignore a stale generation")
	}
	if strings.Contains(htmlOv, "const OBS_WIDTH") {
		t.Fatal("overlay OBS size must update on resize")
	}
	if !strings.Contains(htmlOv, "addEventListener('resize'") {
		t.Fatal("overlay html must listen for window resize")
	}
	if !strings.Contains(htmlOv, `charset="UTF-8"`) {
		t.Fatal("overlay html must set UTF-8 charset")
	}
	if !strings.Contains(htmlOv, "scheduleAudioTeardown") {
		t.Fatal("overlay audio alerts must drop after max(duration, min sticker time)")
	}
}

func overlayJSFunc(html, name string) string {
	sig := "function " + name + "("
	start := strings.Index(html, sig)
	if start < 0 {
		return ""
	}
	rest := html[start+len(sig):]
	next := strings.Index(rest, "\n\tfunction ")
	if next < 0 {
		return html[start:]
	}
	return html[start : start+len(sig)+next]
}

func TestChatTrailingSlashNotFound(t *testing.T) {
	_, ts := startTest(t, Config{})
	res, err := http.Get(ts.URL + PathChat + "/")
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("status %d", res.StatusCode)
	}
}

func TestMediaAndWebhooksOff(t *testing.T) {
	_, ts := startTest(t, Config{})
	res, err := http.Get(ts.URL + PathMedia + "x.mp3")
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("media status %d", res.StatusCode)
	}
	res, err = http.Post(ts.URL+PathWebhook, "application/json", strings.NewReader(`{"message":"x"}`))
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("webhook status %d", res.StatusCode)
	}
}

func TestMediaServesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "clip.txt")
	if err := os.WriteFile(path, []byte("hi"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, ts := startTest(t, Config{MediaPath: dir})
	res, err := http.Get(ts.URL + PathMedia + "clip.txt")
	if err != nil {
		t.Fatal(err)
	}
	b, _ := io.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode != http.StatusOK || string(b) != "hi" {
		t.Fatalf("status %d body %q", res.StatusCode, b)
	}
}

func TestWebhookInjectAndBadJSON(t *testing.T) {
	s, ts := startTest(t, Config{Webhooks: true})
	res, err := http.Post(ts.URL+PathWebhook, "application/json", strings.NewReader(`{"message":"@jump"}`))
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("status %d", res.StatusCode)
	}
	select {
	case msg := <-s.Injected():
		if msg.Author != "webhook" || msg.Text != "@jump" {
			t.Fatalf("%+v", msg)
		}
	case <-time.After(time.Second):
		t.Fatal("no inject")
	}

	res, err = http.Post(ts.URL+PathWebhook, "application/json", strings.NewReader(`not-json`))
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("bad json status %d", res.StatusCode)
	}

	res, err = http.Get(ts.URL + PathWebhook)
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("get webhook %d", res.StatusCode)
	}
}

func TestChatWebSocket(t *testing.T) {
	s, ts := startTest(t, Config{})
	u := "ws" + strings.TrimPrefix(ts.URL, "http") + PathChatWS
	c, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	time.Sleep(30 * time.Millisecond)
	want := NewChatEvent("Ada", "http://pic", "hello", time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC))
	s.PublishChat(want)
	_ = c.SetReadDeadline(time.Now().Add(2 * time.Second))
	var got ChatEvent
	if err := c.ReadJSON(&got); err != nil {
		t.Fatal(err)
	}
	if got.AuthorName != "Ada" || got.MessageText != "hello" || got.AuthorPicture != "http://pic" {
		t.Fatalf("%+v", got)
	}
	if got.PublishedAt != "2026-01-02T03:04:05Z" {
		t.Fatalf("published %q", got.PublishedAt)
	}
}

func TestChatFlushWebSocket(t *testing.T) {
	s, ts := startTest(t, Config{})
	u := "ws" + strings.TrimPrefix(ts.URL, "http") + PathChatWS
	c, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	time.Sleep(30 * time.Millisecond)
	s.PublishChat(FlushChat())
	_ = c.SetReadDeadline(time.Now().Add(2 * time.Second))
	var got ChatEvent
	if err := c.ReadJSON(&got); err != nil {
		t.Fatal(err)
	}
	if got.Type != string(PayloadTypeChatFlush) {
		t.Fatalf("%+v", got)
	}
}

func TestNotifyAppClosed(t *testing.T) {
	s, ts := startTest(t, Config{})
	chatURL := "ws" + strings.TrimPrefix(ts.URL, "http") + PathChatWS
	overURL := "ws" + strings.TrimPrefix(ts.URL, "http") + PathOverlayWS
	chat, _, err := websocket.DefaultDialer.Dial(chatURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer chat.Close()
	over, _, err := websocket.DefaultDialer.Dial(overURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer over.Close()
	time.Sleep(30 * time.Millisecond)
	s.NotifyAppClosed()
	_ = chat.SetReadDeadline(time.Now().Add(2 * time.Second))
	_ = over.SetReadDeadline(time.Now().Add(2 * time.Second))
	var ce ChatEvent
	if err := chat.ReadJSON(&ce); err != nil {
		t.Fatal(err)
	}
	if ce.Type != string(PayloadTypeAppClosed) {
		t.Fatalf("chat %+v", ce)
	}
	var oe OverlayEvent
	if err := over.ReadJSON(&oe); err != nil {
		t.Fatal(err)
	}
	if oe.Type != PayloadTypeAppClosed {
		t.Fatalf("overlay %+v", oe)
	}
}

func TestOverlayWebSocketAndInterrupt(t *testing.T) {
	s, ts := startTest(t, Config{Webhooks: true})
	u := "ws" + strings.TrimPrefix(ts.URL, "http") + PathOverlayWS
	c, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	time.Sleep(30 * time.Millisecond)

	s.PublishOverlay(AlertOverlay("jump.mp3", 0.5, 1.2))
	_ = c.SetReadDeadline(time.Now().Add(2 * time.Second))
	var got OverlayEvent
	if err := c.ReadJSON(&got); err != nil {
		t.Fatal(err)
	}
	if got.Type != PayloadTypeAlert || got.Filename != "jump.mp3" || got.Volume != 0.5 || got.Scale != 1.2 {
		t.Fatalf("%+v", got)
	}

	s.PublishOverlay(TTSOverlay([]byte{1, 2, 3}, 0.8))
	if err := c.ReadJSON(&got); err != nil {
		t.Fatal(err)
	}
	if got.Type != PayloadTypeTTS || got.Volume != 0.8 || string(got.Payload) != "\x01\x02\x03" {
		t.Fatalf("%+v payload %v", got, got.Payload)
	}

	res, err := http.Post(ts.URL+PathInterrupt, "application/json", nil)
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("interrupt %d", res.StatusCode)
	}
	if err := c.ReadJSON(&got); err != nil {
		t.Fatal(err)
	}
	if got.Type != PayloadTypeTTSInterrupt {
		t.Fatalf("%+v", got)
	}
}

func TestListenAndServeEmptyAddr(t *testing.T) {
	s := New(Config{})
	if err := s.ListenAndServe(); err == nil {
		t.Fatal("expected error")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_ = s.Shutdown(ctx)
}

func TestSourceURLs(t *testing.T) {
	if got := ChatSourceURL(":8080"); got != "http://127.0.0.1:8080/obs/chat" {
		t.Fatalf("chat %q", got)
	}
	if got := OverlaySourceURL("9090"); got != "http://127.0.0.1:9090/obs/overlay" {
		t.Fatalf("overlay %q", got)
	}
	if got := ChatSourceURL("0.0.0.0:8080"); got != "http://127.0.0.1:8080/obs/chat" {
		t.Fatalf("wildcard %q", got)
	}
}

func TestOverlayJSONFieldNames(t *testing.T) {
	b, err := json.Marshal(AlertOverlay("a.gif", 1, 1))
	if err != nil {
		t.Fatal(err)
	}
	raw := string(b)
	for _, k := range []string{`"type"`, `"payload"`, `"filename"`, `"volume"`, `"scale"`} {
		if !strings.Contains(raw, k) {
			t.Fatalf("missing %s in %s", k, raw)
		}
	}
}
