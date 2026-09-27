package update

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Apply stages a helper that replaces this install after the process exits.
func (c *Client) Apply() error {
	if err := c.identity(); err != nil {
		return err
	}
	if c.payload == "" {
		return ErrNoDownload
	}
	dest, _, err := installDest(c.executable())
	if err != nil {
		return err
	}
	if !dirWritable(parentOf(dest), c.Name) {
		return fmt.Errorf("update: cannot write %s", parentOf(dest))
	}
	script, err := writeHelper(c.tempDir(), dest, c.payload, c.Name)
	if err != nil {
		return err
	}
	start := c.Start
	if start == nil {
		start = startDetached
	}
	if runtime.GOOS == "windows" {
		return start("powershell", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-File", script)
	}
	return start("/bin/sh", script)
}

func writeHelper(dir, dest, payload, name string) (string, error) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	old := dest + "." + name + "-old"
	if runtime.GOOS == "windows" {
		path := filepath.Join(dir, "apply.ps1")
		if err := os.WriteFile(path, windowsHelper(dest, payload, old), 0o700); err != nil {
			return "", err
		}
		return path, nil
	}
	path := filepath.Join(dir, "apply.sh")
	body := "#!/bin/sh\n" +
		"sleep 2\n" +
		"DEST=\"" + dest + "\"\n" +
		"NEW=\"" + payload + "\"\n" +
		"OLD=\"" + old + "\"\n" +
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

// windowsHelper is UTF-8 with a BOM so PowerShell keeps Unicode install paths.
func windowsHelper(dest, payload, old string) []byte {
	body := strings.Join([]string{
		"$ErrorActionPreference = 'Stop'",
		"Start-Sleep -Seconds 2",
		"$dest = " + psSingleQuote(dest),
		"$new = " + psSingleQuote(payload),
		"$old = " + psSingleQuote(old),
		"if (Test-Path -LiteralPath $old) { Remove-Item -LiteralPath $old -Recurse -Force }",
		"if (Test-Path -LiteralPath $dest) { Move-Item -LiteralPath $dest -Destination $old -Force }",
		"Move-Item -LiteralPath $new -Destination $dest -Force",
		"Start-Process -FilePath $dest",
		"if (Test-Path -LiteralPath $old) { Remove-Item -LiteralPath $old -Recurse -Force }",
		"",
	}, "\r\n")
	out := make([]byte, 0, 3+len(body))
	out = append(out, 0xEF, 0xBB, 0xBF)
	out = append(out, body...)
	return out
}

func psSingleQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}
