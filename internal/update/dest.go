package update

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func installDest(exe string) (dest string, isApp bool, err error) {
	if exe == "" {
		return "", false, fmt.Errorf("update: cannot find this binary")
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return "", false, err
	}
	if runtime.GOOS == "darwin" {
		const marker = ".app" + string(filepath.Separator) + "Contents" + string(filepath.Separator) + "MacOS"
		idx := strings.Index(exe, marker)
		if idx > 0 {
			app := exe[:idx+len(".app")]
			return app, true, nil
		}
	}
	return exe, false, nil
}

func parentOf(path string) string {
	return filepath.Dir(path)
}

func dirWritable(dir string) bool {
	if dir == "" || dir == "." {
		return false
	}
	f, err := os.CreateTemp(dir, ".ytmemchat-write-*")
	if err != nil {
		return false
	}
	name := f.Name()
	_ = f.Close()
	_ = os.Remove(name)
	return true
}
