package update

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// Apply stages a helper that replaces this install after the process exits.
func (c *Client) Apply() error {
	if c.payload == "" {
		return ErrNoDownload
	}
	dest, _, err := installDest(c.executable())
	if err != nil {
		return err
	}
	if !dirWritable(parentOf(dest)) {
		return fmt.Errorf("update: cannot write %s", parentOf(dest))
	}
	script, err := writeHelper(c.tempDir(), dest, c.payload)
	if err != nil {
		return err
	}
	start := c.Start
	if start == nil {
		start = startDetached
	}
	if runtime.GOOS == "windows" {
		return start("cmd.exe", "/C", script)
	}
	return start("/bin/sh", script)
}

func writeHelper(dir, dest, payload string) (string, error) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	if runtime.GOOS == "windows" {
		path := filepath.Join(dir, "apply.bat")
		body := "@echo off\r\n" +
			"timeout /t 2 /nobreak >nul\r\n" +
			"if exist \"" + dest + ".ytmemchat-old\" rmdir /s /q \"" + dest + ".ytmemchat-old\"\r\n" +
			"if exist \"" + dest + "\" move /y \"" + dest + "\" \"" + dest + ".ytmemchat-old\"\r\n" +
			"move /y \"" + payload + "\" \"" + dest + "\"\r\n" +
			"start \"\" \"" + dest + "\"\r\n" +
			"rmdir /s /q \"" + dest + ".ytmemchat-old\"\r\n"
		if err := os.WriteFile(path, []byte(body), 0o700); err != nil {
			return "", err
		}
		return path, nil
	}
	path := filepath.Join(dir, "apply.sh")
	body := "#!/bin/sh\n" +
		"sleep 2\n" +
		"DEST=\"" + dest + "\"\n" +
		"NEW=\"" + payload + "\"\n" +
		"OLD=\"${DEST}.ytmemchat-old\"\n" +
		"rm -rf \"$OLD\"\n" +
		"if [ -e \"$DEST\" ]; then mv \"$DEST\" \"$OLD\"; fi\n" +
		"mv \"$NEW\" \"$DEST\"\n" +
		"chmod -R u+w \"$DEST\" 2>/dev/null || true\n" +
		"if [ -d \"$DEST\" ]; then open \"$DEST\"; else nohup \"$DEST\" >/dev/null 2>&1 & fi\n" +
		"rm -rf \"$OLD\"\n"
	if err := os.WriteFile(path, []byte(body), 0o700); err != nil {
		return "", err
	}
	return path, nil
}
