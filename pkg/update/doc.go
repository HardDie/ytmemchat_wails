// Package update checks GitHub Releases, downloads the matching archive,
// and verifies SHA-256 before a helper replaces the running install.
//
// The caller passes the GitHub owner, repository, and binary name.
// This package does not know which app it updates.
package update

import (
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"
)

const maxAssetBytes = 200 << 20

// ErrNoAsset is returned when this OS/arch has no release archive.
var ErrNoAsset = fmt.Errorf("update: no archive for this OS")

// ErrChecksum is returned when SHA-256 does not match SHA256SUMS.txt.
var ErrChecksum = fmt.Errorf("update: checksum mismatch")

// ErrNotReady is returned when Download or Apply runs before Check.
var ErrNotReady = fmt.Errorf("update: check GitHub first")

// ErrNoDownload is returned when Apply runs before a verified download.
var ErrNoDownload = fmt.Errorf("update: download the archive first")

// Config is the app identity for one GitHub Releases client.
type Config struct {
	// Owner is the GitHub user or organization.
	Owner string
	// Repo is the GitHub repository name.
	Repo string
	// Name is the file stem inside the archive: Name, Name.exe, or Name.app.
	Name string
	// Current is this binary’s stamped version.
	Current string
	// UserAgent is sent on GitHub HTTP calls. Empty uses Name.
	UserAgent string
}

// Status is one GitHub latest-release check.
type Status struct {
	// Current is this binary’s stamped version.
	Current string `json:"current"`
	// Latest is the GitHub release tag (for example v0.2.0).
	Latest string `json:"latest"`
	// Notes is the release body (may be truncated).
	Notes string `json:"notes"`
	// URL is the HTML release page.
	URL string `json:"url"`
	// Asset is the archive file name for this OS.
	Asset string `json:"asset"`
	// Newer is true when Latest is a semver tag greater than Current.
	Newer bool `json:"newer"`
	// CanInstall is true when an archive exists and the install dir is writable.
	CanInstall bool `json:"canInstall"`
	// Same is true when Current equals Latest (both release tags).
	Same bool `json:"same"`
}

// Client talks to GitHub and stages a verified archive.
type Client struct {
	// HTTP calls GitHub. Nil uses http.DefaultClient.
	HTTP *http.Client
	// API is the GitHub API root. New sets https://api.github.com.
	API string
	// Owner is the GitHub user or organization.
	Owner string
	// Repo is the GitHub repository name.
	Repo string
	// Name is the file stem inside the archive: Name, Name.exe, or Name.app.
	Name string
	// UserAgent is sent on GitHub HTTP calls. Empty uses Name.
	UserAgent string
	// Current is this binary’s stamped version.
	Current string
	// GOOS selects the archive suffix. New sets runtime.GOOS.
	GOOS string
	// GOARCH selects the archive suffix. New sets runtime.GOARCH.
	GOARCH string
	// Executable is this install. Empty uses os.Executable.
	Executable string
	// TempDir stages the download. Empty uses os.TempDir()/Name-update.
	TempDir string
	// Start launches the install helper. Nil starts it detached.
	Start func(name string, args ...string) error

	last     Status
	assetURL string
	sumsURL  string
	archive  string
	payload  string
}

// New uses the public GitHub API and this process’s OS.
// Owner, Repo, and Name come from cfg.
func New(cfg Config) *Client {
	ua := strings.TrimSpace(cfg.UserAgent)
	if ua == "" {
		ua = strings.TrimSpace(cfg.Name)
	}
	if ua == "" {
		ua = "update"
	}
	return &Client{
		HTTP:      &http.Client{Timeout: 60 * time.Second},
		API:       "https://api.github.com",
		Owner:     strings.TrimSpace(cfg.Owner),
		Repo:      strings.TrimSpace(cfg.Repo),
		Name:      strings.TrimSpace(cfg.Name),
		UserAgent: ua,
		Current:   cfg.Current,
		GOOS:      runtime.GOOS,
		GOARCH:    runtime.GOARCH,
		Start:     startDetached,
	}
}

func (c *Client) identity() error {
	if c.Owner == "" || c.Repo == "" || strings.Contains(c.Owner, "/") || strings.Contains(c.Repo, "/") {
		return fmt.Errorf("update: set GitHub owner and repo")
	}
	if !validName(c.Name) {
		return fmt.Errorf("update: set a binary name without path characters")
	}
	return nil
}

func validName(name string) bool {
	if name == "" || name == "." || name == ".." {
		return false
	}
	return !strings.ContainsAny(name, `/\:*?"<>|`)
}

func (c *Client) releasesPage() string {
	return "https://github.com/" + c.Owner + "/" + c.Repo + "/releases"
}

func (c *Client) userAgent() string {
	if c.UserAgent != "" {
		return c.UserAgent
	}
	if c.Name != "" {
		return c.Name
	}
	return "update"
}
