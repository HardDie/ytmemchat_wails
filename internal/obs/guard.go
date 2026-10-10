package obs

import (
	"net"
	"net/http"
	"net/url"
	"strings"
)

// guard rejects requests that a web page on another site could send.
// OBS and curl pass: they use an IP or localhost and send no foreign Origin.
func guard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !allowedHost(r.Host) {
			http.Error(w, "forbidden host", http.StatusForbidden)
			return
		}
		if r.Method != http.MethodGet && r.Method != http.MethodHead && !sameOrigin(r) {
			http.Error(w, "forbidden origin", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// allowedHost accepts an IP address or "localhost", with or without a port.
// Any other name is refused. That blocks DNS rebinding, where a site points
// its own domain at this machine to read and post as same-origin.
func allowedHost(hostport string) bool {
	host := hostport
	if h, _, err := net.SplitHostPort(hostport); err == nil {
		host = h
	}
	host = strings.TrimSuffix(strings.TrimPrefix(host, "["), "]")
	if strings.EqualFold(host, "localhost") {
		return true
	}
	return net.ParseIP(host) != nil
}

// sameOrigin is true without an Origin header (curl, scripts) or when the
// Origin is this server. Browsers send Origin on every cross-site POST.
func sameOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	return u.Scheme == "http" && strings.EqualFold(u.Host, r.Host)
}
