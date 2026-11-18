# Code review (Go backend, OBS pages)

Working notes. Not an ADR. Not a wiki how-to.

Scope:

1. Go code in `.`, `bindings/`, `internal/`, `pkg/`.
2. `internal/obs/*.html` and `script.js` (output sinks only).
3. Svelte UI was skimmed, not reviewed.

Baseline:

1. `make test` passes with `-race`.
2. `go vet -tags=nomain ./...` is clean.

How to use this list:

1. Discuss one item at a time in chat.
2. After a decision, set that item’s status.
3. Do not start a code change until the item is `fix`.
4. Show the change for review before taking the next item.
5. Status values: `open`, `discussing`, `skip`, `fix`, `done`.

---

## Bugs

1. Windows TTS: PowerShell injection from chat — `done`
  1. File: `internal/tts/tts_windows.go` (`synthesize`, `speak`).
  2. Chat text is pasted into a `'…'` PowerShell literal.
  3. Only ASCII `'` is doubled.
  4. PowerShell also treats `‘ ’ ‚ ‛` (U+2018–U+201B) as quotes.
  5. `removeEmojis` keeps them (they are punctuation, not symbols).
  6. Chat line `x’); Start-Process calc; $null=(’` runs `calc`.
  7. Any viewer can run commands on the streamer’s PC.
  8. The temp file path is pasted unescaped too.
  9. Idea: pass text, voice, and path in env vars, not in the script.
  10. Fix: `psSynthScript` is a constant in `powershell_script.go`.
  11. Text, voice, and path go in `YTMEMCHAT_TTS_*` env vars.
  12. Windows only (`//go:build windows`), like `powershell.go`.
  13. Dead `speak` (both OSes) and `tts.Speak` removed with it.
2. macOS / Linux TTS: chat text is parsed as a CLI flag — `done`
  1. File: `internal/tts/tts_unix.go` (`synthesize`, `speak`).
  2. Text is the last argv item of `say` / `espeak`.
  3. A message that starts with `-` becomes an option.
  4. Verified: `say … "-opwn.aiff"` created an empty `pwn.aiff`.
  5. `-o/Users/<you>/.zshrc` would truncate that file.
  6. `espeak -w<path>` is the same on Linux.
  7. Idea: send text on stdin (`say -f -`, `espeak --stdin`).
  8. Verified: `say -f -` reads the same text safely.
  9. Fix: `synthCommand` in `tts_unix.go` sends text on stdin.
  10. `say … -f -` and `espeak … --stdin`; argv never holds chat text.
  11. Unit test `tts_unix_test.go` checks argv and stdin.
  12. Integration test speaks `-o<pwn>` / `-w<pwn>` and checks no file.
  13. Verified locally: it fails on the old code, passes on the new.
  14. Windows integration test speaks a `’` payload that ran `New-Item`.
  15. CI: new `integration` job on Linux, macOS, Windows; Linux installs `espeak`.
  16. A temporary `-v` CI run showed the Linux TTS tests PASS, not SKIP.
  17. TTS integration tests now fail instead of skipping (`requireTool`).
3. Publish after `Server.Shutdown` panics — `done`
  1. File: `internal/obs/hub.go` (`publish`, `shutdown`).
  2. `shutdown` closes `send`; a later `publish` sends on it.
  3. Verified with a throwaway test: `send on closed channel`.
  4. Trigger: TTS or alert in flight while settings restart OBS HTTP.
  5. Restart happens on port, media folder, or webhook change.
  6. TTS synthesis is about one second, so the window is real.
  7. Test pane, alert preview, and interrupt hotkey race the same way.
  8. Fix: `hub.done` is closed by `shutdown`; `send` is never closed.
  9. `publish` drops a message after shutdown (decision: no log line).
  10. `loop` stops on `done`; a blocked `publish` returns on shutdown.
  11. `serveWS` refuses to register a socket once `done` is closed.
  12. Tests in `internal/obs/hub_test.go` (all OSes, CI, `-race`).
  13. Verified: all three new tests fail on the old `hub.go`.
4. Webhook is open to any web page and to the LAN — `fix` (in review)
  1. Files: `internal/obs/api.go`, `internal/config/validate.go`.
  2. No auth on `/api/webhook` and `/api/interrupt`.
  3. `Content-Type` is not checked.
  4. A `text/plain` POST from any site skips the CORS preflight.
  5. Any page in the browser can inject TTS lines.
  6. With item 1 this is drive-by code execution on Windows.
  7. `ListenAddr` is `:port`, so every interface listens.
  8. The LAN also sees the media folder and both sockets.
  9. Idea: listen on `127.0.0.1` by default.
  10. Idea: require `application/json` on POST.
  11. Decision: loopback by default, plus a LAN toggle (second-PC setup).
  12. Decision: guard middleware, not a `Content-Type` rule (keeps curl).
  13. Decision: WebSockets use the same-origin check.
  14. Fix: `config.Server.LAN`; `ListenAddr` is `127.0.0.1:<port>` unless LAN.
  15. Fix: `obs.guard` wraps all routes: IP or `localhost` Host only.
  16. Fix: non-GET with a foreign `Origin` gets 403.
  17. Fix: `upgrader` drops `CheckOrigin: true`.
  18. UI: Configuration checkbox “Listen on the local network”.
  19. Tests: `guard_test.go`; they fail with the guard disabled.
  20. Chromium check: pages and chat socket work via 127.0.0.1, localhost, LAN IP.
  21. Docs: ADR 032, CURSOR.md, wiki Configuration, README.
5. One stalled WebSocket client blocks all output — `open`
  1. File: `internal/obs/hub.go` (`broadcast`).
  2. Writes run under `h.mu` with no write deadline.
  3. A client that stops reading blocks the hub loop.
  4. After 64 queued events `publish` blocks the chat goroutine.
6. Media URL is not encoded in the overlay — `open`
  1. File: `internal/obs/overlay.html` (`'/obs/media/' + filename`).
  2. `#` starts a fragment; `?` starts a query; `%` is an escape.
  3. `meme #1.mp4` or `100%.gif` never loads.
  4. Idea: `encodeURIComponent` per path segment.
7. Chat pollers hide failures — `open`
  1. Files: `internal/youtube/v3.go`, `internal/youtube/nokey/client.go`.
  2. v3 retries every error forever at the poll interval.
  3. That includes quota exceeded and chat disabled.
  4. The UI keeps showing Running; no error reaches the window.
  5. nokey never detects chat end; it polls the old token forever.
  6. Both use HTTP clients with no timeout.
8. Alert token matches only the first token — `open`
  1. File: `internal/alerts/alerts.go` (`findToken`).
  2. `@streamer @boom` checks only `streamer`.
  3. `@boom!` looks up `boom!` and misses.
  4. May be intended; needs a decision.
9. Update helper pastes paths into shell double quotes — `open`
  1. File: `pkg/update/apply.go` (`writeHelper`).
  2. `$`, backtick, `"` or `\` in the install path break the script.
  3. `$(…)` in the path would run.
  4. Idea: single-quote escape like `psSingleQuote`.
10. Checksum and asset names match loosely — `open`
  1. Files: `pkg/update/checksum.go`, `pkg/update/asset.go`.
  2. `checksumFor` accepts suffix matches both ways.
  3. `pickAsset` uses `strings.Contains`.
  4. `….tar.gz.sig` would match the archive hint.
  5. No security impact: sums and assets share one release.
11. `archive.Extract` rollback is partial — `open`
  1. File: `pkg/archive/extract.go`.
  2. The doc says a failed extract leaves no new files.
  3. Overwritten files are not restored.
  4. Created folders are not removed.
12. `go test ./...` crashes on macOS without `nomain` — `open`
  1. File: `internal/hotkey/binding_test.go` (`TestBinding_enabled`).
  2. It registers a real hotkey off the main thread.
  3. `make test` and CI pass because they use `-tags=nomain`.

---

## Simplifications

13. Drop the `Out` channels in alerts and TTS — `open`
  1. Files: `internal/alerts/alerts.go`, `internal/tts/tts.go`, `overlay.go`.
  2. Each call sends on a channel the caller reads right back.
  3. Return `(Clip, bool)` and `(Speech, error)` instead.
  4. Removes `Config.Out`, nil-channel errors, and a block risk.
14. One “path inside root” helper — `open`
  1. Four copies: `mediaRelativePath`, `archive.relativeInside`.
  2. Also `archive.insideDir` and `update.safeJoin`.
  3. Go 1.25 has `filepath.IsLocal` and `os.Root`.
  4. `os.Root` would also cover most of `placeExtracted`.
15. One atomic write helper — `open`
  1. Temp file then rename is in `config/store.go`.
  2. Also in `alerts.SaveFile` and `archive.writeFile`.
16. Remove dead code — `open`
  1. `config.DirPath`.
  2. `tts.Speak` and both `speak` funcs.
  3. `tts.SynthesizeToBuffer` (tests only).
  4. `obs.SocketQuery` (tests use it, so recheck before removing).
  5. `update.IsReleaseTag`.
  6. `commands.writeCommandsExport` (tests only).
  7. `MessageIterator.Next` (app uses `GetChan` only).
  8. Dead means unreachable on every target OS.
  9. `deadcode` was run with `GOOS=darwin` and `GOOS=windows`.
  10. Linux needs cgo, so the tool cannot load it on macOS.
  11. Linux builds `tts_unix.go`, same as darwin.
  12. A grep of all `.go` files finds no caller for any item above.
  13. Windows also flags `parseSayVoiceList`; it is used on unix, so not dead.
  14. Test-only helpers stay: `newAppWithStore`, `config.NewStore`, `secret.Memory`.
  15. Sub-item 2 is done as part of item 1.
17. Small cleanups — `open`
  1. `ErrOBSNotListening` is declared in three binding packages.
  2. v3 and nokey repeat the same poll loop shape.
  3. `config.Store.hydrate`: the `writeFile` error branch is a no-op.
