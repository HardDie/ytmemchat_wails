package update

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// Download fetches the archive from the last Check and verifies SHA-256.
func (c *Client) Download() error {
	if c.assetURL == "" || c.sumsURL == "" || c.last.Asset == "" {
		return ErrNotReady
	}
	sums, err := c.getBytes(c.sumsURL, 1<<20)
	if err != nil {
		return fmt.Errorf("update: sums: %w", err)
	}
	want, err := checksumFor(string(sums), c.last.Asset)
	if err != nil {
		return err
	}
	dir := c.tempDir()
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	path := filepath.Join(dir, filepath.Base(c.last.Asset))
	if err := c.getFile(c.assetURL, path); err != nil {
		return err
	}
	got, err := fileSHA256(path)
	if err != nil {
		return err
	}
	if got != want {
		_ = os.Remove(path)
		return fmt.Errorf("%w: got %s want %s", ErrChecksum, got, want)
	}
	payload, err := unpack(path, filepath.Join(dir, "payload"))
	if err != nil {
		return err
	}
	c.archive = path
	c.payload = payload
	return nil
}

func (c *Client) getBytes(url string, limit int64) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	res, err := c.http().Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", res.StatusCode)
	}
	return io.ReadAll(io.LimitReader(res.Body, limit))
}

func (c *Client) getFile(url, dest string) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", userAgent)
	res, err := c.http().Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("update: download HTTP %d", res.StatusCode)
	}
	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	n, err := io.Copy(f, io.LimitReader(res.Body, maxAssetBytes+1))
	closeErr := f.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	if n > maxAssetBytes {
		_ = os.Remove(dest)
		return fmt.Errorf("update: archive too large")
	}
	return nil
}

func fileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func (c *Client) tempDir() string {
	if c.TempDir != "" {
		return c.TempDir
	}
	return filepath.Join(os.TempDir(), "ytmemchat-update")
}

func (c *Client) executable() string {
	if c.Executable != "" {
		return c.Executable
	}
	p, err := os.Executable()
	if err != nil {
		return ""
	}
	return p
}
