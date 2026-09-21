package update

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"unicode/utf8"
)

type ghRelease struct {
	TagName string    `json:"tag_name"`
	HTMLURL string    `json:"html_url"`
	Body    string    `json:"body"`
	Assets  []ghAsset `json:"assets"`
}

type ghAsset struct {
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
}

// Check fetches GitHub’s latest release for this OS.
func (c *Client) Check() (Status, error) {
	c.assetURL = ""
	c.sumsURL = ""
	rel, err := c.fetchLatest()
	if err != nil {
		return Status{}, err
	}
	st := Status{
		Current: c.Current,
		Latest:  strings.TrimSpace(rel.TagName),
		Notes:   truncate(rel.Body, 2000),
		URL:     rel.HTMLURL,
		Newer:   Newer(rel.TagName, c.Current),
		Same:    SameTag(rel.TagName, c.Current),
	}
	if st.URL == "" {
		st.URL = ReleasesPage
	}
	asset, err := pickAsset(rel.Assets, c.GOOS, c.GOARCH)
	if err != nil {
		if !errors.Is(err, ErrNoAsset) {
			return Status{}, err
		}
	} else {
		st.Asset = asset.Name
		c.assetURL = asset.URL
		if sums, serr := pickSums(rel.Assets); serr == nil {
			c.sumsURL = sums.URL
			dest, _, derr := installDest(c.executable())
			st.CanInstall = derr == nil && dirWritable(parentOf(dest))
		}
	}
	c.last = st
	c.archive = ""
	c.payload = ""
	return st, nil
}

func (c *Client) fetchLatest() (ghRelease, error) {
	u := strings.TrimRight(c.API, "/") + "/repos/" + c.Owner + "/" + c.Repo + "/releases/latest"
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return ghRelease{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", userAgent)
	res, err := c.http().Do(req)
	if err != nil {
		return ghRelease{}, fmt.Errorf("update: github: %w", err)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return ghRelease{}, err
	}
	if res.StatusCode != http.StatusOK {
		return ghRelease{}, fmt.Errorf("update: github HTTP %d", res.StatusCode)
	}
	var rel ghRelease
	if err := json.Unmarshal(body, &rel); err != nil {
		return ghRelease{}, fmt.Errorf("update: github json: %w", err)
	}
	if strings.TrimSpace(rel.TagName) == "" {
		return ghRelease{}, fmt.Errorf("update: empty release tag")
	}
	return rel, nil
}

func (c *Client) http() *http.Client {
	if c.HTTP != nil {
		return c.HTTP
	}
	return http.DefaultClient
}

func truncate(s string, n int) string {
	s = strings.TrimSpace(s)
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	r := []rune(s)
	return string(r[:n]) + "…"
}
