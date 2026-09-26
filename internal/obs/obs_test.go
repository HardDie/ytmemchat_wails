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
	if !strings.Contains(string(cb), PathScript) {
		t.Fatal("chat html must load the shared script")
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
	if strings.Contains(string(cb), "function connectWebSocket") {
		t.Fatal("chat html must use the shared connectWebSocket")
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
	if strings.Contains(string(cb), "Ready for media alerts.") {
		t.Fatal("chat html must not log overlay media-alert copy")
	}
	if !strings.Contains(string(cb), "Ready for chat.") {
		t.Fatal("chat html must log chat socket ready")
	}
	if !strings.Contains(string(cb), "debug: received message") {
		t.Fatal("chat html must log received messages when debug is set")
	}
	if !strings.Contains(string(cb), `textColor.charAt(0) !== '#'`) {
		t.Fatal("chat html must accept textColor with or without #")
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
	if !strings.Contains(string(ob), PathScript) {
		t.Fatal("overlay html must load the shared script")
	}
	if strings.Contains(string(ob), "location.reload") {
		t.Fatal("overlay html must not reload on unhandled errors")
	}
	if strings.Contains(string(ob), "function connectWebSocket") {
		t.Fatal("overlay html must use the shared connectWebSocket")
	}
	if !strings.Contains(string(ob), "onConnectionLost") {
		t.Fatal("overlay html must tear down playback when the socket drops")
	}
	if !strings.Contains(string(ob), "MIN_STICKER_PLAYING_SEC") {
		t.Fatal("overlay html must keep video on screen at least min sticker time")
	}
	for _, phrase := range []string{"debug: received message", "video command", "audio command", "image command", "'tts'", "not found type"} {
		if !strings.Contains(string(ob), phrase) {
			t.Fatalf("overlay html missing debug phrase %q", phrase)
		}
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
	if strings.Contains(htmlOv, "start-btn") {
		t.Fatal("overlay html must not require an audio unlock button")
	}
	if strings.Contains(htmlOv, "decodeAudioData") {
		t.Fatal("overlay TTS must not use AudioContext decodeAudioData")
	}
	if !strings.Contains(htmlOv, "startAudibleSticker") {
		t.Fatal("overlay TTS and command audio must use an HTML audio sticker")
	}
	if !strings.Contains(htmlOv, "Ready for media alerts.") {
		t.Fatal("overlay html must log media-alert socket ready")
	}

	jsRes, err := http.Get(ts.URL + PathScript)
	if err != nil {
		t.Fatal(err)
	}
	jsBody, _ := io.ReadAll(jsRes.Body)
	jsRes.Body.Close()
	if jsRes.StatusCode != http.StatusOK {
		t.Fatalf("script status %d", jsRes.StatusCode)
	}
	if ct := jsRes.Header.Get("Content-Type"); !strings.Contains(ct, "javascript") {
		t.Fatalf("script content-type %q", ct)
	}
	js := string(jsBody)
	if !strings.Contains(js, "location.pathname") {
		t.Fatal("shared script must derive websocket from location")
	}
	if !strings.Contains(js, "location.search") {
		t.Fatal("shared script must pass the page query to the socket")
	}
	if !strings.Contains(js, "version_redirect") {
		t.Fatal("shared script must follow a version redirect")
	}
	if strings.Contains(js, "location.reload") {
		t.Fatal("shared script must not reload on unhandled errors")
	}
	if !strings.Contains(js, "recoverFromUnhandledError") {
		t.Fatal("shared script must recover via connectWebSocket")
	}
	if !strings.Contains(js, "CONNECTING_TIMEOUT_MS") {
		t.Fatal("shared script must abort a stuck CONNECTING socket")
	}
	if !strings.Contains(js, "function connectWebSocket") {
		t.Fatal("shared script must own connectWebSocket")
	}
	if strings.Contains(js, "clearOverlayPlayback") {
		t.Fatal("shared script must not call overlay teardown by name")
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

func TestPageVersionRedirect(t *testing.T) {
	_, ts := startTest(t, Config{Version: "v9"})
	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}}
	res, err := client.Get(ts.URL + PathChat + "?cap=3&transparent=1")
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusFound {
		t.Fatalf("status %d", res.StatusCode)
	}
	if loc := res.Header.Get("Location"); loc != "/obs/chat?cap=3&transparent=1&v=v9" {
		t.Fatalf("location %q", loc)
	}

	ok, err := client.Get(ts.URL + PathOverlay + "?" + QueryVersion + "=v9")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(ok.Body)
	ok.Body.Close()
	if ok.StatusCode != http.StatusOK {
		t.Fatalf("status %d", ok.StatusCode)
	}
	if !strings.Contains(string(body), PathScript+"?v=v9") {
		t.Fatalf("script src %s", body)
	}
	if strings.Contains(string(body), "__APP_VERSION__") {
		t.Fatal("version token left in overlay html")
	}

	u := "ws" + strings.TrimPrefix(ts.URL, "http") + PathChatWS + "?cap=3&" + QueryVersion + "=old"
	c, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	_ = c.SetReadDeadline(time.Now().Add(2 * time.Second))
	var msg versionRedirect
	if err := c.ReadJSON(&msg); err != nil {
		t.Fatal(err)
	}
	if msg.Type != "version_redirect" || msg.URL != "/obs/chat?cap=3&v=v9" {
		t.Fatalf("%+v", msg)
	}
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
	u := "ws" + strings.TrimPrefix(ts.URL, "http") + PathChatWS + SocketQuery("")
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
	u := "ws" + strings.TrimPrefix(ts.URL, "http") + PathChatWS + SocketQuery("")
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
	chatURL := "ws" + strings.TrimPrefix(ts.URL, "http") + PathChatWS + SocketQuery("")
	overURL := "ws" + strings.TrimPrefix(ts.URL, "http") + PathOverlayWS + SocketQuery("")
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
	u := "ws" + strings.TrimPrefix(ts.URL, "http") + PathOverlayWS + SocketQuery("")
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

func TestPublishDebugFlag(t *testing.T) {
	s, ts := startTest(t, Config{})
	chatURL := "ws" + strings.TrimPrefix(ts.URL, "http") + PathChatWS + SocketQuery("")
	overURL := "ws" + strings.TrimPrefix(ts.URL, "http") + PathOverlayWS + SocketQuery("")
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

	s.SetDebug(true)
	s.PublishChat(NewChatEvent("Ada", "", "hi", time.Time{}))
	s.PublishOverlay(AlertOverlay("jump.mp3", 1, 1))
	_ = chat.SetReadDeadline(time.Now().Add(2 * time.Second))
	_ = over.SetReadDeadline(time.Now().Add(2 * time.Second))
	var ce ChatEvent
	if err := chat.ReadJSON(&ce); err != nil {
		t.Fatal(err)
	}
	if !ce.Debug || ce.AuthorName != "Ada" || ce.MessageText != "hi" {
		t.Fatalf("chat %+v", ce)
	}
	var oe OverlayEvent
	if err := over.ReadJSON(&oe); err != nil {
		t.Fatal(err)
	}
	if !oe.Debug || oe.Type != PayloadTypeAlert || oe.Filename != "jump.mp3" {
		t.Fatalf("overlay %+v", oe)
	}

	s.SetDebug(false)
	s.PublishChat(NewChatEvent("Ada", "", "next", time.Time{}))
	ce = ChatEvent{}
	if err := chat.ReadJSON(&ce); err != nil {
		t.Fatal(err)
	}
	if ce.Debug || ce.MessageText != "next" {
		t.Fatalf("chat after off %+v", ce)
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
