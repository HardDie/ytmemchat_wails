package archive

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

var (
	// ErrOutside is returned when a file resolves outside Root.
	ErrOutside = errors.New("path is outside the root")
	// ErrMissing is returned when a listed file does not exist.
	ErrMissing = errors.New("file is missing")
	// ErrReserved is returned when a listed file uses DocumentName.
	ErrReserved = errors.New("path uses the document name")
	// ErrNotFile is returned when a listed path is not a regular file.
	ErrNotFile = errors.New("not a regular file")
	// ErrDocumentMissing is returned when Document does not exist.
	ErrDocumentMissing = errors.New("document is missing")
	// ErrRootDocument is returned when DocumentName is not a file at the archive root.
	ErrRootDocument = errors.New("document is not at the archive root")
)

// PathError is a failure for one path in [Spec.Files].
type PathError struct {
	// Path is the value the caller passed.
	Path string
	// Name is the archive path when one was chosen.
	Name string
	// Err is the reason.
	Err error
}

func (e *PathError) Error() string {
	label := e.Name
	if label == "" {
		label = e.Path
	}
	return label + ": " + e.Err.Error()
}

func (e *PathError) Unwrap() error { return e.Err }

// Spec is one zip.
type Spec struct {
	// Dest is where the zip is written. A path without a .zip suffix gets one.
	Dest string
	// Document is the on-disk file stored under DocumentName.
	Document string
	// DocumentName is Document's path inside the zip, usually at the root.
	DocumentName string
	// Root is the directory Files are relative to.
	Root string
	// Files are the paths to store.
	// A relative path is inside Root.
	// An absolute path must stay inside Root and is stored as that relative path.
	// The same archive path is stored once.
	Files []string
}

type packedFile struct {
	name string
	abs  string
	mod  time.Time
	mode os.FileMode
	data []byte
}

// Validate checks Document and Files without writing a zip.
func Validate(spec Spec) error {
	_, _, err := collect(spec)
	return err
}

// Write builds the zip. A failed write leaves no archive at Dest.
func Write(spec Spec) error {
	if strings.TrimSpace(spec.Dest) == "" {
		return fmt.Errorf("archive: empty destination")
	}
	doc, files, err := collect(spec)
	if err != nil {
		return err
	}
	return writeFile(spec.Dest, doc, files)
}

func collect(spec Spec) (packedFile, []packedFile, error) {
	docName, err := cleanArchiveName(spec.DocumentName)
	if err != nil {
		return packedFile{}, nil, err
	}
	docPath := strings.TrimSpace(spec.Document)
	if docPath == "" {
		return packedFile{}, nil, fmt.Errorf("archive: empty document")
	}
	info, err := os.Stat(docPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return packedFile{}, nil, ErrDocumentMissing
		}
		return packedFile{}, nil, fmt.Errorf("archive: document: %w", err)
	}
	if info.IsDir() {
		return packedFile{}, nil, fmt.Errorf("archive: document is a folder")
	}
	raw, err := os.ReadFile(docPath)
	if err != nil {
		return packedFile{}, nil, fmt.Errorf("archive: document: %w", err)
	}
	doc := packedFile{
		name: docName,
		abs:  docPath,
		mod:  info.ModTime(),
		mode: 0o644,
		data: raw,
	}
	files, err := collectFiles(spec.Root, docName, spec.Files)
	if err != nil {
		return packedFile{}, nil, err
	}
	return doc, files, nil
}

func collectFiles(root, docName string, list []string) ([]packedFile, error) {
	if len(list) == 0 {
		return nil, nil
	}
	root = strings.TrimSpace(root)
	if root == "" {
		return nil, fmt.Errorf("archive: empty root")
	}
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("archive: root: %w", err)
	}
	rootAbs = filepath.Clean(rootAbs)
	rootReal, err := filepath.EvalSymlinks(rootAbs)
	if err != nil {
		return nil, fmt.Errorf("archive: root: %w", err)
	}
	rootReal = filepath.Clean(rootReal)
	seen := make(map[string]struct{}, len(list))
	out := make([]packedFile, 0, len(list))
	for _, file := range list {
		file = strings.TrimSpace(file)
		if file == "" {
			continue
		}
		abs, zipName, err := locate(rootAbs, rootReal, file)
		if err != nil {
			return nil, &PathError{Path: file, Err: err}
		}
		if strings.EqualFold(zipName, docName) {
			return nil, &PathError{Path: file, Name: zipName, Err: ErrReserved}
		}
		if _, ok := seen[zipName]; ok {
			continue
		}
		info, err := os.Stat(abs)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return nil, &PathError{Path: file, Name: zipName, Err: ErrMissing}
			}
			return nil, &PathError{Path: file, Name: zipName, Err: err}
		}
		if !info.Mode().IsRegular() {
			return nil, &PathError{Path: file, Name: zipName, Err: ErrNotFile}
		}
		seen[zipName] = struct{}{}
		out = append(out, packedFile{
			name: zipName,
			abs:  abs,
			mod:  info.ModTime(),
			mode: info.Mode().Perm(),
		})
	}
	return out, nil
}

// locate returns the on-disk path and the slash path stored in the zip.
func locate(rootAbs, rootReal, file string) (string, string, error) {
	var candidate string
	if filepath.IsAbs(filepath.FromSlash(file)) {
		candidate = filepath.Clean(file)
	} else {
		rel := filepath.Clean(filepath.FromSlash(file))
		sep := string(filepath.Separator)
		if rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+sep) {
			return "", "", ErrOutside
		}
		candidate = filepath.Join(rootAbs, rel)
	}
	zipName, err := relativeInside(rootAbs, candidate)
	if err != nil {
		return "", "", err
	}
	real, err := filepath.EvalSymlinks(candidate)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return candidate, zipName, nil
		}
		return "", "", err
	}
	if _, err := relativeInside(rootReal, real); err != nil {
		return "", "", ErrOutside
	}
	return candidate, zipName, nil
}

func relativeInside(root, file string) (string, error) {
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	fileAbs, err := filepath.Abs(file)
	if err != nil {
		return "", err
	}
	rootAbs = filepath.Clean(rootAbs)
	fileAbs = filepath.Clean(fileAbs)
	sep := string(filepath.Separator)
	prefix := rootAbs + sep
	if fileAbs != rootAbs && !strings.HasPrefix(fileAbs, prefix) {
		return "", ErrOutside
	}
	rel, err := filepath.Rel(rootAbs, fileAbs)
	if err != nil {
		return "", err
	}
	rel = filepath.Clean(rel)
	if rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+sep) {
		return "", ErrOutside
	}
	return filepath.ToSlash(rel), nil
}

func cleanArchiveName(name string) (string, error) {
	name = filepath.ToSlash(filepath.Clean(filepath.FromSlash(strings.TrimSpace(name))))
	if name == "" || name == "." || name == ".." || strings.HasPrefix(name, "../") || strings.HasPrefix(name, "/") {
		return "", fmt.Errorf("archive: invalid document name %q", name)
	}
	return name, nil
}

func writeFile(dest string, doc packedFile, files []packedFile) error {
	dest = strings.TrimSpace(dest)
	if !strings.EqualFold(filepath.Ext(dest), ".zip") {
		dest += ".zip"
	}
	dir := filepath.Dir(dest)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("archive: %w", err)
	}
	tmp, err := os.CreateTemp(dir, ".archive-*.zip")
	if err != nil {
		return fmt.Errorf("archive: %w", err)
	}
	tmpName := tmp.Name()
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(tmpName)
		}
	}()
	if err := writeZip(tmp, doc, files); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("archive: %w", err)
	}
	if runtime.GOOS == "windows" {
		_ = os.Remove(dest)
	}
	if err := os.Rename(tmpName, dest); err != nil {
		return fmt.Errorf("archive: %w", err)
	}
	cleanup = false
	return nil
}

func writeZip(w io.Writer, doc packedFile, files []packedFile) error {
	zw := zip.NewWriter(w)
	if err := addStoredBytes(zw, doc.name, doc.data, doc.mode, doc.mod); err != nil {
		_ = zw.Close()
		return err
	}
	for _, f := range files {
		if err := addStoredFile(zw, f); err != nil {
			_ = zw.Close()
			return err
		}
	}
	if err := zw.Close(); err != nil {
		return fmt.Errorf("archive: %w", err)
	}
	return nil
}

func addStoredBytes(zw *zip.Writer, name string, data []byte, mode os.FileMode, mod time.Time) error {
	dst, err := createStored(zw, name, uint64(len(data)), mode, mod)
	if err != nil {
		return err
	}
	if _, err := dst.Write(data); err != nil {
		return fmt.Errorf("archive: %w", err)
	}
	return nil
}

func addStoredFile(zw *zip.Writer, f packedFile) error {
	in, err := os.Open(f.abs)
	if err != nil {
		return fmt.Errorf("archive: %s: %w", f.name, err)
	}
	defer in.Close()
	info, err := in.Stat()
	if err != nil {
		return fmt.Errorf("archive: %s: %w", f.name, err)
	}
	dst, err := createStored(zw, f.name, uint64(info.Size()), f.mode, f.mod)
	if err != nil {
		return err
	}
	if _, err := io.Copy(dst, in); err != nil {
		return fmt.Errorf("archive: %s: %w", f.name, err)
	}
	return nil
}

func createStored(zw *zip.Writer, name string, size uint64, mode os.FileMode, mod time.Time) (io.Writer, error) {
	hdr := &zip.FileHeader{
		Name:               name,
		Method:             zip.Store,
		Modified:           mod.UTC(),
		UncompressedSize64: size,
	}
	if mode == 0 {
		mode = 0o644
	}
	hdr.SetMode(mode)
	w, err := zw.CreateHeader(hdr)
	if err != nil {
		return nil, fmt.Errorf("archive: %w", err)
	}
	return w, nil
}
