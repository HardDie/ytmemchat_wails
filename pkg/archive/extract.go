package archive

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

// ReadRootFile returns the bytes of documentName at the root of the zip.
// Every other entry name is checked as well. Nothing is written.
func ReadRootFile(src, documentName string) ([]byte, error) {
	names, err := listEntryNames(src, documentName)
	if err != nil {
		return nil, err
	}
	doc := names.doc
	r, err := zip.OpenReader(src)
	if err != nil {
		return nil, fmt.Errorf("archive: %w", err)
	}
	defer r.Close()
	for _, f := range r.File {
		name, skip, err := entryName(f.Name)
		if err != nil || skip || f.FileInfo().IsDir() {
			continue
		}
		if name != doc {
			continue
		}
		return readZipFile(f)
	}
	return nil, ErrRootDocument
}

// Extract unpacks src into dest.
// documentName must be a file at the archive root.
// Each other file keeps its path relative to that root.
// A path that leaves dest is rejected. A failed extract does not leave new files in dest.
func Extract(src, dest, documentName string) error {
	listed, err := listEntryNames(src, documentName)
	if err != nil {
		return err
	}
	dest = strings.TrimSpace(dest)
	if dest == "" {
		return fmt.Errorf("archive: empty destination")
	}
	destAbs, err := filepath.Abs(dest)
	if err != nil {
		return fmt.Errorf("archive: %w", err)
	}
	if err := os.MkdirAll(destAbs, 0o755); err != nil {
		return fmt.Errorf("archive: %w", err)
	}
	destReal, err := filepath.EvalSymlinks(destAbs)
	if err != nil {
		return fmt.Errorf("archive: %w", err)
	}
	r, err := zip.OpenReader(src)
	if err != nil {
		return fmt.Errorf("archive: %w", err)
	}
	defer r.Close()
	byName := make(map[string]*zip.File, len(listed.names))
	for _, f := range r.File {
		name, skip, err := entryName(f.Name)
		if err != nil || skip || f.FileInfo().IsDir() {
			continue
		}
		byName[name] = f
	}
	parent := filepath.Dir(destAbs)
	tmp, err := os.MkdirTemp(parent, ".archive-import-*")
	if err != nil {
		return fmt.Errorf("archive: %w", err)
	}
	defer os.RemoveAll(tmp)
	for _, name := range listed.names {
		f := byName[name]
		if f == nil {
			return fmt.Errorf("archive: missing %s", name)
		}
		if err := writeZipTo(tmp, name, f); err != nil {
			return err
		}
	}
	var created []string
	for _, name := range listed.names {
		final, err := placeExtracted(tmp, destReal, name)
		if err != nil {
			removeCreated(created)
			return err
		}
		created = append(created, final)
	}
	return nil
}

type listedEntries struct {
	doc   string
	names []string
}

func listEntryNames(src, documentName string) (listedEntries, error) {
	doc, err := cleanArchiveName(documentName)
	if err != nil {
		return listedEntries{}, err
	}
	r, err := zip.OpenReader(src)
	if err != nil {
		return listedEntries{}, fmt.Errorf("archive: %w", err)
	}
	defer r.Close()
	seen := make(map[string]struct{})
	var names []string
	found := false
	for _, f := range r.File {
		name, skip, err := entryName(f.Name)
		if err != nil {
			return listedEntries{}, &PathError{Path: f.Name, Err: err}
		}
		if skip || f.FileInfo().IsDir() {
			continue
		}
		if f.Mode()&os.ModeSymlink != 0 {
			return listedEntries{}, &PathError{Path: name, Name: name, Err: ErrNotFile}
		}
		if _, ok := seen[name]; ok {
			return listedEntries{}, fmt.Errorf("archive: duplicate %s", name)
		}
		seen[name] = struct{}{}
		if name == doc {
			found = true
		}
		names = append(names, name)
	}
	if !found {
		return listedEntries{}, ErrRootDocument
	}
	return listedEntries{doc: doc, names: names}, nil
}

func entryName(name string) (string, bool, error) {
	name = strings.TrimSpace(strings.ReplaceAll(name, "\\", "/"))
	if name == "" {
		return "", false, fmt.Errorf("archive: empty entry name")
	}
	if strings.HasSuffix(name, "/") {
		return "", true, nil
	}
	name = strings.TrimPrefix(name, "/")
	clean := path.Clean(name)
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") || path.IsAbs(clean) {
		return "", false, ErrOutside
	}
	return clean, false, nil
}

func readZipFile(f *zip.File) ([]byte, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, fmt.Errorf("archive: %w", err)
	}
	defer rc.Close()
	buf, err := io.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("archive: %w", err)
	}
	return buf, nil
}

func writeZipTo(root, name string, f *zip.File) error {
	rel := filepath.FromSlash(name)
	dest := filepath.Join(root, rel)
	if err := insideDir(root, dest); err != nil {
		return &PathError{Path: name, Name: name, Err: err}
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return fmt.Errorf("archive: %w", err)
	}
	rc, err := f.Open()
	if err != nil {
		return fmt.Errorf("archive: %s: %w", name, err)
	}
	defer rc.Close()
	out, err := os.OpenFile(dest, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("archive: %s: %w", name, err)
	}
	if _, err := io.Copy(out, rc); err != nil {
		_ = out.Close()
		return fmt.Errorf("archive: %s: %w", name, err)
	}
	if err := out.Close(); err != nil {
		return fmt.Errorf("archive: %s: %w", name, err)
	}
	return nil
}

func placeExtracted(tmp, destReal, name string) (string, error) {
	rel := filepath.FromSlash(name)
	src := filepath.Join(tmp, rel)
	final := filepath.Join(destReal, rel)
	if err := insideDir(destReal, final); err != nil {
		return "", &PathError{Path: name, Name: name, Err: err}
	}
	parent := filepath.Dir(final)
	if _, err := os.Lstat(parent); err == nil {
		parentReal, err := filepath.EvalSymlinks(parent)
		if err != nil {
			return "", fmt.Errorf("archive: %w", err)
		}
		if err := insideDir(destReal, parentReal); err != nil {
			return "", &PathError{Path: name, Name: name, Err: err}
		}
	}
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return "", fmt.Errorf("archive: %w", err)
	}
	parentReal, err := filepath.EvalSymlinks(parent)
	if err != nil {
		return "", fmt.Errorf("archive: %w", err)
	}
	if err := insideDir(destReal, parentReal); err != nil {
		return "", &PathError{Path: name, Name: name, Err: err}
	}
	existed := false
	if _, err := os.Lstat(final); err == nil {
		existed = true
	}
	if err := moveFile(src, final); err != nil {
		return "", err
	}
	if existed {
		return "", nil
	}
	return final, nil
}

func moveFile(src, dest string) error {
	if err := os.Rename(src, dest); err == nil {
		return nil
	}
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("archive: %w", err)
	}
	defer in.Close()
	tmp, err := os.CreateTemp(filepath.Dir(dest), ".archive-file-*")
	if err != nil {
		return fmt.Errorf("archive: %w", err)
	}
	tmpName := tmp.Name()
	if _, err := io.Copy(tmp, in); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
		return fmt.Errorf("archive: %w", err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("archive: %w", err)
	}
	if runtime.GOOS == "windows" {
		_ = os.Remove(dest)
	}
	if err := os.Rename(tmpName, dest); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("archive: %w", err)
	}
	return nil
}

func removeCreated(paths []string) {
	for i := len(paths) - 1; i >= 0; i-- {
		if paths[i] == "" {
			continue
		}
		_ = os.Remove(paths[i])
	}
}

func insideDir(root, target string) error {
	root = filepath.Clean(root)
	target = filepath.Clean(target)
	sep := string(filepath.Separator)
	if target != root && !strings.HasPrefix(target, root+sep) {
		return ErrOutside
	}
	rel, err := filepath.Rel(root, target)
	if err != nil {
		return ErrOutside
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+sep) {
		return ErrOutside
	}
	return nil
}
