package update

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func unpack(archive, dest string) (string, error) {
	_ = os.RemoveAll(dest)
	if err := os.MkdirAll(dest, 0o700); err != nil {
		return "", err
	}
	switch {
	case strings.HasSuffix(strings.ToLower(archive), ".zip"):
		if err := unzip(archive, dest); err != nil {
			return "", err
		}
	case strings.HasSuffix(strings.ToLower(archive), ".tar.gz") || strings.HasSuffix(strings.ToLower(archive), ".tgz"):
		if err := untarGz(archive, dest); err != nil {
			return "", err
		}
	default:
		return "", fmt.Errorf("update: unknown archive %s", filepath.Base(archive))
	}
	return findPayload(dest)
}

func findPayload(root string) (string, error) {
	var app, bin, win string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		name := info.Name()
		if info.IsDir() && strings.HasSuffix(name, ".app") {
			app = path
			return filepath.SkipDir
		}
		if info.IsDir() {
			return nil
		}
		if name == "ytmemchat.exe" {
			win = path
		}
		if name == "ytmemchat" && info.Mode()&0o111 != 0 {
			bin = path
		}
		if name == "ytmemchat" && bin == "" {
			bin = path
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if app != "" {
		return app, nil
	}
	if win != "" {
		return win, nil
	}
	if bin != "" {
		return bin, nil
	}
	return "", fmt.Errorf("update: archive has no ytmemchat binary")
}

func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()
	for _, f := range r.File {
		if err := unzipFile(dest, f); err != nil {
			return err
		}
	}
	return nil
}

func unzipFile(dest string, f *zip.File) error {
	name, err := safeJoin(dest, f.Name)
	if err != nil {
		return err
	}
	if strings.HasSuffix(f.Name, "/") {
		return os.MkdirAll(name, 0o755)
	}
	if err := os.MkdirAll(filepath.Dir(name), 0o755); err != nil {
		return err
	}
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()
	out, err := os.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode())
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(out, rc)
	closeErr := out.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}

func untarGz(src, dest string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		name, err := safeJoin(dest, hdr.Name)
		if err != nil {
			return err
		}
		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(name, 0o755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(name), 0o755); err != nil {
				return err
			}
			out, err := os.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(hdr.Mode))
			if err != nil {
				return err
			}
			_, copyErr := io.Copy(out, tr)
			closeErr := out.Close()
			if copyErr != nil {
				return copyErr
			}
			if closeErr != nil {
				return closeErr
			}
		}
	}
}

func safeJoin(dest, name string) (string, error) {
	clean := filepath.Clean("/" + strings.ReplaceAll(name, "\\", "/"))
	rel := strings.TrimPrefix(clean, "/")
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("update: bad path %q", name)
	}
	full := filepath.Join(dest, rel)
	if !strings.HasPrefix(full, dest+string(filepath.Separator)) && full != dest {
		return "", fmt.Errorf("update: bad path %q", name)
	}
	return full, nil
}
