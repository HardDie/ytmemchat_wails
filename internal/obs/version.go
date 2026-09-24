package obs

import (
	"bytes"
	"net/http"
	"net/url"
	"strings"
)

const appVersionToken = "__APP_VERSION__"

// versionRedirect tells the OBS page to load the same path with the running version.
type versionRedirect struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

// SocketQuery is the ?v= query a WebSocket must send to stay connected.
// An empty version means "dev".
func SocketQuery(version string) string {
	v := strings.TrimSpace(version)
	if v == "" {
		v = "dev"
	}
	return "?" + QueryVersion + "=" + url.QueryEscape(v)
}

func (s *Server) appVersion() string {
	v := strings.TrimSpace(s.cfg.Version)
	if v == "" {
		return "dev"
	}
	return v
}

func versionedHTML(body []byte, version string) []byte {
	escaped := url.QueryEscape(version)
	return bytes.ReplaceAll(body, []byte(appVersionToken), []byte(escaped))
}

func (s *Server) redirectIfStaleVersion(w http.ResponseWriter, r *http.Request) bool {
	if r.URL.Query().Get(QueryVersion) == s.appVersion() {
		return false
	}
	http.Redirect(w, r, pageURLWithVersion(r.URL.Path, r.URL.Query(), s.appVersion()), http.StatusFound)
	return true
}

func (s *Server) serveVersionedPage(w http.ResponseWriter, r *http.Request, body []byte) {
	if s.redirectIfStaleVersion(w, r) {
		return
	}
	writeHTML(w, versionedHTML(body, s.appVersion()))
}

func (s *Server) redirectStaleSocket(w http.ResponseWriter, r *http.Request) bool {
	if r.URL.Query().Get(QueryVersion) == s.appVersion() {
		return false
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.log.Debug("obs version redirect upgrade", "err", err)
		return true
	}
	defer conn.Close()
	page := strings.TrimSuffix(r.URL.Path, "/ws")
	_ = conn.WriteJSON(versionRedirect{
		Type: "version_redirect",
		URL:  pageURLWithVersion(page, r.URL.Query(), s.appVersion()),
	})
	return true
}

func pageURLWithVersion(path string, q url.Values, version string) string {
	next := cloneQuery(q)
	next.Set(QueryVersion, version)
	return path + "?" + next.Encode()
}

func cloneQuery(q url.Values) url.Values {
	next := make(url.Values, len(q))
	for k, vs := range q {
		next[k] = append([]string(nil), vs...)
	}
	return next
}
