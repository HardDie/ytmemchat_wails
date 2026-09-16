package youtube

import "net/http"

// withAPIKey returns a clone of c that adds key as the Data API query param.
// Required when using option.WithHTTPClient: that option skips WithAPIKey.
func withAPIKey(c *http.Client, key string) *http.Client {
	if c == nil {
		c = &http.Client{}
	}
	clone := *c
	base := clone.Transport
	if base == nil {
		base = http.DefaultTransport
	}
	clone.Transport = apiKeyTransport{base: base, key: key}
	return &clone
}

type apiKeyTransport struct {
	base http.RoundTripper
	key  string
}

func (t apiKeyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	cloned := req.Clone(req.Context())
	q := cloned.URL.Query()
	q.Set("key", t.key)
	cloned.URL.RawQuery = q.Encode()
	return t.base.RoundTrip(cloned)
}
