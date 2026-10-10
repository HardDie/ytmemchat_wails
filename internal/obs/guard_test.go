package obs

import (
	"net/http"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

func TestAllowedHost(t *testing.T) {
	tests := []struct {
		host string
		want bool
	}{
		{"127.0.0.1:8080", true},
		{"127.0.0.1", true},
		{"localhost:8080", true},
		{"LocalHost", true},
		{"[::1]:8080", true},
		{"192.168.1.10:8080", true},
		{"10.0.0.5", true},
		{"evil.example:8080", false},
		{"127.0.0.1.nip.io:8080", false},
		{"localhost.evil.example", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := allowedHost(tt.host); got != tt.want {
			t.Errorf("allowedHost(%q) = %v, want %v", tt.host, got, tt.want)
		}
	}
}

func doRequest(t *testing.T, method, url, host, origin string) int {
	t.Helper()
	req, err := http.NewRequest(method, url, strings.NewReader(`{"message":"hi"}`))
	if err != nil {
		t.Fatal(err)
	}
	if host != "" {
		req.Host = host
	}
	if origin != "" {
		req.Header.Set("Origin", origin)
	}
	req.Header.Set("Content-Type", "text/plain")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	return res.StatusCode
}

// OBS on this PC and a browser on another PC load pages by IP; both must pass.
func TestGuard_OBSPages(t *testing.T) {
	_, ts := startTest(t, Config{Version: "v1.0.0"})
	page := ts.URL + PathOverlay + "?v=v1.0.0"
	if got := doRequest(t, http.MethodGet, page, "", ""); got != http.StatusOK {
		t.Fatalf("127.0.0.1: status %d", got)
	}
	if got := doRequest(t, http.MethodGet, page, "localhost:8080", ""); got != http.StatusOK {
		t.Fatalf("localhost: status %d", got)
	}
	if got := doRequest(t, http.MethodGet, page, "192.168.1.10:8080", ""); got != http.StatusOK {
		t.Fatalf("LAN IP: status %d", got)
	}
	if got := doRequest(t, http.MethodGet, page, "evil.example:8080", ""); got != http.StatusForbidden {
		t.Fatalf("DNS rebinding host: status %d", got)
	}
}

func TestGuard_webhookOrigin(t *testing.T) {
	_, ts := startTest(t, Config{Webhooks: true})
	u := ts.URL + PathWebhook
	self := "http://" + strings.TrimPrefix(ts.URL, "http://")
	tests := []struct {
		name   string
		origin string
		want   int
	}{
		{"curl, no origin", "", http.StatusNoContent},
		{"same origin", self, http.StatusNoContent},
		{"other site", "http://evil.example", http.StatusForbidden},
		{"other port", "http://127.0.0.1:1", http.StatusForbidden},
		{"opaque origin", "null", http.StatusForbidden},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := doRequest(t, http.MethodPost, u, "", tt.origin); got != tt.want {
				t.Fatalf("status %d, want %d", got, tt.want)
			}
		})
	}
	if got := doRequest(t, http.MethodPost, ts.URL+PathInterrupt, "", "http://evil.example"); got != http.StatusForbidden {
		t.Fatalf("interrupt from other site: status %d", got)
	}
}

func dialWS(t *testing.T, u, host, origin string) (*websocket.Conn, int, error) {
	t.Helper()
	h := http.Header{}
	if host != "" {
		h.Set("Host", host)
	}
	if origin != "" {
		h.Set("Origin", origin)
	}
	c, res, err := websocket.DefaultDialer.Dial(u, h)
	code := 0
	if res != nil {
		code = res.StatusCode
	}
	return c, code, err
}

// OBS opens the socket from the page this server served, so Origin matches Host.
func TestGuard_websocketOrigin(t *testing.T) {
	_, ts := startTest(t, Config{})
	hostport := strings.TrimPrefix(ts.URL, "http://")
	u := "ws://" + hostport + PathOverlayWS + SocketQuery("")

	c, _, err := dialWS(t, u, "", "http://"+hostport)
	if err != nil {
		t.Fatalf("OBS on this PC: %v", err)
	}
	c.Close()

	lan := "192.168.1.10:8080"
	c, _, err = dialWS(t, u, lan, "http://"+lan)
	if err != nil {
		t.Fatalf("browser on another PC: %v", err)
	}
	c.Close()

	c, code, err := dialWS(t, u, "", "http://evil.example")
	if err == nil {
		c.Close()
		t.Fatal("socket from another site was accepted")
	}
	if code != http.StatusForbidden {
		t.Fatalf("other site: status %d", code)
	}
}
