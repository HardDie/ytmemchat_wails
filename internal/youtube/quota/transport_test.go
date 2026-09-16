package quota

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestMethodFromRequest(t *testing.T) {
	tests := []struct {
		path string
		want Method
		ok   bool
	}{
		{"/youtube/v3/videos", VideosList, true},
		{"/youtube/v3/liveChat/messages", LiveChatMessagesList, true},
		{"/youtube/v3/search", SearchList, true},
		{"/youtube/v3/channels", Unknown, true},
		{"/discovery/v1/apis/youtube/v3/rest", "", false},
		{"/healthz", "", false},
		{"/youtube/v3/", "", false},
	}
	for _, tt := range tests {
		got, ok := MethodFromRequest(&http.Request{URL: &url.URL{Path: tt.path}})
		if ok != tt.ok || got != tt.want {
			t.Fatalf("path %q: got (%q,%v), want (%q,%v)", tt.path, got, ok, tt.want, tt.ok)
		}
	}
	if _, ok := MethodFromRequest(nil); ok {
		t.Fatal("nil request")
	}
}

func TestWrapClient_countsHTTPResponsesOnly(t *testing.T) {
	tr := NewTracker()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "videos") {
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, `{"items":[]}`)
			return
		}
		if strings.Contains(r.URL.Path, "search") {
			w.WriteHeader(http.StatusForbidden)
			_, _ = io.WriteString(w, `{"error":{"message":"quotaExceeded"}}`)
			return
		}
		http.NotFound(w, r)
	}))
	t.Cleanup(srv.Close)

	c := WrapClient(srv.Client(), tr)
	respOK, err := c.Get(srv.URL + "/youtube/v3/videos")
	if err != nil {
		t.Fatal(err)
	}
	respOK.Body.Close()
	resp, err := c.Get(srv.URL + "/youtube/v3/search")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if health, err := c.Get(srv.URL + "/healthz"); err == nil {
		health.Body.Close()
	}

	got := tr.Snapshot()
	if got.Units != 1 {
		t.Fatalf("units = %d, want 1 (videos only; /healthz ignored)", got.Units)
	}
	if got.Search != 1 {
		t.Fatalf("search = %d, want 1 (403 still counts)", got.Search)
	}
}

func TestWrapClient_skipsTransportErrors(t *testing.T) {
	tr := NewTracker()
	c := WrapClient(&http.Client{Transport: errTransport{}}, tr)
	_, err := c.Get("http://127.0.0.1/youtube/v3/videos")
	if err == nil {
		t.Fatal("expected transport error")
	}
	if got := tr.Snapshot(); got.Units != 0 {
		t.Fatalf("transport error counted: %+v", got)
	}
}

func TestWrapClient_nilTrackerUnchanged(t *testing.T) {
	orig := &http.Client{}
	got := WrapClient(orig, nil)
	if got != orig {
		t.Fatal("nil tracker should return the same client")
	}
}

type errTransport struct{}

func (errTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, io.ErrUnexpectedEOF
}
