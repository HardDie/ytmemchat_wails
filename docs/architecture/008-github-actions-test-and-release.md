# 8. GitHub Actions: test on push, binaries on tag

* **Status:** Accepted
* **Date:** 2026-09-15
* **Authors:** @oleg

---

## Context

1. The desktop app must stay green as packages are ported.
2. Streamers need downloadable binaries.
3. Wails uses CGO and OS webviews.
4. A single Linux `GOOS=windows` cross-compile is not enough.
5. Packages in `internal/` should stay testable without GTK/WebView.

## Considered options

1. **No CI** — local `go test` only.
2. **`dAppServer/wails-build-action`** — less YAML; extra third-party action; older Go defaults.
3. **Repo workflows**
   1. `test.yml` on every push/PR (`go test -tags=nomain .` then `./internal/...`).
   2. `release.yml` on `v*` tags with a matrix of native runners (Linux amd64, Linux arm64, Windows amd64, macOS universal).

## Decision

Use option 3.

1. **Unit tests**
   1. Live next to the package (`foo_test.go`).
   2. Run on every push: `go test -race -tags=nomain .`
   3. Then `go test -race -tags=nomain ./bindings/...`
   4. Then `go test -race ./internal/...`
2. **Integration tests**
   1. Use `//go:build integration` and files named `*_integration_test.go`.
   2. OS-specific cases add the GOOS (`integration && darwin`, `linux`, `windows`).
   3. They compile only on that platform.
   4. Run on every push: `go test -tags=integration ./internal/...`
   5. They must not need YouTube credentials or a display.
   6. Skip (don’t fail) if an optional tool is missing **on that OS**.
3. **`internal/` stays CGO-free** so Ubuntu CI does not install WebKit for tests.
4. **Releases**
   1. Tag `v1.2.3` (semver with `v` prefix).
   2. Checkout with `fetch-depth: 0` so tags exist.
   3. Install the Wails CLI at the same version as `go.mod`.
   4. Prefetch modules (`go mod download`, retries) with `GODEBUG=http2client=0`.
   5. That avoids proxy.golang.org HTTP/2 stream errors on large zips (notably `google.golang.org/genproto`) on arm64.
   6. Build with `wails build -skipbindings -platform …`.
   7. Stamp version with `-ldflags "-X github.com/HardDie/ytmemchat_wails/bindings/sidebar.buildVersion=…"`.
   8. Bindings are committed under `frontend/wailsjs`. Generate locally after changing pane binding methods.
   9. Linux (Ubuntu 24.04) also passes `-tags webkit2_41` because the runner has `libwebkit2gtk-4.1-dev`, not 4.0.
   10. Native runners: `ubuntu-latest`, `ubuntu-24.04-arm`, `windows-latest`, `macos-latest` (`darwin/universal`).
   11. Linux jobs install `libx11-dev` and run under `xvfb-run`.
   12. `golang.design/x/hotkey` panics in `init()` when `DISPLAY` is unset.
   13. The `publish` job uploads binary archives and `SHA256SUMS.txt` to GitHub Release assets.
   14. `README.md` loads `window.gif` from `docs/screenshots/window.gif`.
5. Until `go.mod` exists, the test workflow skips instead of failing.
6. A tag without `wails.json` fails the release job.
7. Code signing is out of scope for the first slice.

## Consequences

### Positive

* Domain logic is gated on every push.
* Four OS/arch artifacts from one tag.
* Contributors can run the same commands locally.

### Negative and risks

* macOS universal and Windows builds need GitHub-hosted macOS/Windows minutes.
* `ubuntu-24.04-arm` availability depends on GitHub. If it disappears, switch Linux arm64 to a documented cross-compile fallback.
* `go test ./...` from the repo root pulls Wails CGO (`main.go`). CI and `make test` use `-tags=nomain` on `.` plus `./internal/...`.

### Neutral

* Workflows: `.github/workflows/test.yml`, `.github/workflows/release.yml`.
