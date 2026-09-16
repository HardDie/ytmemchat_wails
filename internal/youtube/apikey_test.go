package youtube

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWithAPIKey_setsQueryParam(t *testing.T) {
	var got string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.URL.Query().Get("key")
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)
	c := withAPIKey(srv.Client(), "secret-key")
	resp, err := c.Get(srv.URL + "/youtube/v3/videos")
	if err != nil {
		t.Fatal(err)
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	if got != "secret-key" {
		t.Fatal("API key missing from request")
	}
}
