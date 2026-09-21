package update

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewer(t *testing.T) {
	if !Newer("v0.2.0", "v0.1.0") {
		t.Fatal("0.2 > 0.1")
	}
	if Newer("v0.1.0", "v0.1.0") {
		t.Fatal("same")
	}
	if Newer("v1.0.0", "dev") {
		t.Fatal("dev is not comparable")
	}
	if !SameTag("v1.2.3", "1.2.3") {
		t.Fatal("same tag")
	}
}

func TestAssetHint(t *testing.T) {
	h, err := AssetHint("linux", "amd64")
	if err != nil || h != "linux-amd64.tar.gz" {
		t.Fatalf("%q %v", h, err)
	}
	if _, err := AssetHint("plan9", "amd64"); !errors.Is(err, ErrNoAsset) {
		t.Fatalf("err %v", err)
	}
}

func TestChecksumFor(t *testing.T) {
	sums := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa  ytmemchat-v1.0.0-ytmemchat-linux-amd64.tar.gz\n"
	got, err := checksumFor(sums, "ytmemchat-v1.0.0-ytmemchat-linux-amd64.tar.gz")
	if err != nil || got != strings.Repeat("a", 64) {
		t.Fatalf("%q %v", got, err)
	}
}

func TestCheckDownload_ok(t *testing.T) {
	zipBytes := zipWithFile(t, "ytmemchat", []byte("bin"))
	sum := sha256.Sum256(zipBytes)
	asset := "ytmemchat-v1.2.0-foo-darwin-universal.zip"
	sumsBody := hex.EncodeToString(sum[:]) + "  " + asset + "\n"

	var ts *httptest.Server
	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/HardDie/ytmemchat_wails/releases/latest":
			_, _ = io.WriteString(w, `{
				"tag_name": "v1.2.0",
				"html_url": "http://example/rel",
				"body": "hello notes",
				"assets": [
					{"name": "`+asset+`", "browser_download_url": "`+ts.URL+`/file.zip"},
					{"name": "SHA256SUMS.txt", "browser_download_url": "`+ts.URL+`/SHA256SUMS.txt"}
				]
			}`)
		case "/file.zip":
			_, _ = w.Write(zipBytes)
		case "/SHA256SUMS.txt":
			_, _ = w.Write([]byte(sumsBody))
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	c := New("v1.0.0")
	c.API = ts.URL
	c.GOOS = "darwin"
	c.GOARCH = "arm64"
	c.TempDir = t.TempDir()
	c.Executable = filepath.Join(t.TempDir(), "ytmemchat")
	if err := os.WriteFile(c.Executable, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}

	st, err := c.Check()
	if err != nil {
		t.Fatal(err)
	}
	if !st.Newer || st.Latest != "v1.2.0" || st.Asset != asset || !st.CanInstall {
		t.Fatalf("%+v", st)
	}
	if err := c.Download(); err != nil {
		t.Fatal(err)
	}
	var started []string
	c.Start = func(name string, args ...string) error {
		started = append(started, name)
		started = append(started, args...)
		return nil
	}
	if err := c.Apply(); err != nil {
		t.Fatal(err)
	}
	if len(started) == 0 {
		t.Fatal("helper not started")
	}
}

func TestDownload_badChecksum(t *testing.T) {
	zipBytes := zipWithFile(t, "ytmemchat", []byte("bin"))
	asset := "ytmemchat-v1.2.0-foo-linux-amd64.tar.gz"
	sumsBody := strings.Repeat("b", 64) + "  " + asset + "\n"
	var ts *httptest.Server
	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/HardDie/ytmemchat_wails/releases/latest":
			_, _ = io.WriteString(w, `{
				"tag_name": "v1.2.0",
				"html_url": "http://example/rel",
				"body": "",
				"assets": [
					{"name": "`+asset+`", "browser_download_url": "`+ts.URL+`/file.bin"},
					{"name": "SHA256SUMS.txt", "browser_download_url": "`+ts.URL+`/SHA256SUMS.txt"}
				]
			}`)
		case "/file.bin":
			_, _ = w.Write(zipBytes)
		case "/SHA256SUMS.txt":
			_, _ = w.Write([]byte(sumsBody))
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()
	c := New("v1.0.0")
	c.API = ts.URL
	c.GOOS = "linux"
	c.GOARCH = "amd64"
	c.TempDir = t.TempDir()
	c.Executable = filepath.Join(t.TempDir(), "ytmemchat")
	_ = os.WriteFile(c.Executable, []byte("old"), 0o755)
	if _, err := c.Check(); err != nil {
		t.Fatal(err)
	}
	if err := c.Download(); !errors.Is(err, ErrChecksum) {
		t.Fatalf("err = %v", err)
	}
}

func TestUntarFind(t *testing.T) {
	dir := t.TempDir()
	archive := filepath.Join(dir, "a.tar.gz")
	if err := writeTarGz(archive, "ytmemchat", []byte("elf")); err != nil {
		t.Fatal(err)
	}
	got, err := unpack(archive, filepath.Join(dir, "out"))
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(got) != "ytmemchat" {
		t.Fatalf("payload %s", got)
	}
}

func zipWithFile(t *testing.T, name string, body []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, err := zw.Create(name)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write(body); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func writeTarGz(path, name string, body []byte) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	gz := gzip.NewWriter(f)
	tw := tar.NewWriter(gz)
	hdr := &tar.Header{Name: name, Mode: 0755, Size: int64(len(body))}
	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}
	if _, err := tw.Write(body); err != nil {
		return err
	}
	if err := tw.Close(); err != nil {
		return err
	}
	return gz.Close()
}
