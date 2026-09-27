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
		return start("powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-NoExit", "-File", script)
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
		"$ErrorActionPreference = 'Continue'",
		"Write-Host 'ytmemchat update'",
		"$dest = " + psSingleQuote(dest),
		"$new = " + psSingleQuote(payload),
		"$old = " + psSingleQuote(old),
		"Write-Host \"dest: $dest\"",
		"Write-Host \"new:  $new\"",
		"Write-Host \"old:  $old\"",
		"if (-not (Test-Path -LiteralPath $new)) { Write-Host \"ERROR: new file missing: $new\"; return }",
		"$ok = $false",
		"foreach ($i in 1..30) {",
		"  try {",
		"    if (Test-Path -LiteralPath $old) { Remove-Item -LiteralPath $old -Recurse -Force -ErrorAction Stop }",
		"    if (Test-Path -LiteralPath $dest) { Move-Item -LiteralPath $dest -Destination $old -Force -ErrorAction Stop }",
		"    Move-Item -LiteralPath $new -Destination $dest -Force -ErrorAction Stop",
		"    $ok = $true",
		"    break",
		"  } catch {",
		"    Write-Host ('try ' + $i + ': ' + $_.Exception.Message)",
		"    Start-Sleep -Seconds 1",
		"  }",
		"}",
		"if (-not $ok) { Write-Host 'ERROR: could not replace the exe'; return }",
		"Write-Host 'replaced'",
		"Start-Process -FilePath $dest -WorkingDirectory (Split-Path -LiteralPath $dest)",
		"Write-Host 'started'",
		"if (Test-Path -LiteralPath $old) {",
		"  try { Remove-Item -LiteralPath $old -Recurse -Force -ErrorAction Stop; Write-Host 'removed old' }",
		"  catch { Write-Host (\"left old file: \" + $_.Exception.Message) }",
		"}",
		"Write-Host 'Done. You can close this window.'",
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
