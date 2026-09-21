// Package update checks GitHub Releases for a newer ytmemchat build,
// downloads the OS archive, and verifies SHA-256 before install.
package update

import (
	"fmt"
	"net/http"
	"runtime"
	"time"
)

const (
	// DefaultOwner is the GitHub user or org.
	DefaultOwner = "HardDie"
	// DefaultRepo is the GitHub repository name.
	DefaultRepo = "ytmemchat_wails"
	// ReleasesPage is the browser URL for tagged builds.
	ReleasesPage  = "https://github.com/" + DefaultOwner + "/" + DefaultRepo + "/releases"
	maxAssetBytes = 200 << 20
	userAgent     = "ytmemchat"
)

// ErrNoAsset is returned when this OS/arch has no release archive.
var ErrNoAsset = fmt.Errorf("update: no archive for this OS")

// ErrChecksum is returned when SHA-256 does not match SHA256SUMS.txt.
var ErrChecksum = fmt.Errorf("update: checksum mismatch")

// ErrNotReady is returned when Download or Apply runs before Check.
var ErrNotReady = fmt.Errorf("update: check GitHub first")

// ErrNoDownload is returned when Apply runs before a verified download.
var ErrNoDownload = fmt.Errorf("update: download the archive first")

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
	HTTP       *http.Client
	API        string
	Owner      string
	Repo       string
	Current    string
	GOOS       string
	GOARCH     string
	Executable string
	TempDir    string
	Start      func(name string, args ...string) error

	last     Status
	assetURL string
	sumsURL  string
	archive  string
	payload  string
}

// New uses the public GitHub API and this process’s version and OS.
func New(current string) *Client {
	return &Client{
		HTTP:    &http.Client{Timeout: 60 * time.Second},
		API:     "https://api.github.com",
		Owner:   DefaultOwner,
		Repo:    DefaultRepo,
		Current: current,
		GOOS:    runtime.GOOS,
		GOARCH:  runtime.GOARCH,
		Start:   startDetached,
	}
}
