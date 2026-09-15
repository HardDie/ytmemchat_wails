# 8. GitHub Actions: test on push, binaries on tag

* **Status:** Accepted
* **Date:** 2026-09-15
* **Authors:** @oleg

---

## Context

The desktop app must stay green as packages are ported, and streamers need downloadable binaries. Wails uses CGO and OS webviews, so a single Linux `GOOS=windows` cross-compile is not enough. Packages in `internal/` should stay testable without GTK/WebView.

## Considered options

1. **No CI** — local `go test` only.
2. **`dAppServer/wails-build-action`** — less YAML; extra third-party action; older Go defaults.
3. **Repo workflows:** `test.yml` on every push/PR (`go test -tags=nomain .` then `./internal/...`); `release.yml` on `v*` tags with a matrix of native runners (Linux amd64, Linux arm64, Windows amd64, macOS universal).

## Decision

Use option 3.

* **Unit tests** live next to the package (`foo_test.go`). Run on every push: `go test -race -tags=nomain .` (App/Start/Stop, no Wails CGO) then `go test -race ./internal/...`.
* **Integration tests** use `//go:build integration` and files named `*_integration_test.go`. OS-specific cases add the GOOS (`integration && darwin`, `linux`, `windows`) so they compile only on that platform. Run on every push: `go test -tags=integration ./internal/...`. They must not need YouTube credentials or a display. Skip (don’t fail) if an optional tool is missing **on that OS**.
* **`internal/` stays CGO-free** so Ubuntu CI does not install WebKit for tests.
* **Releases:** tag `v1.2.3` (semver with `v` prefix). Build with `wails build -platform …` on `ubuntu-latest`, `ubuntu-24.04-arm`, `windows-latest`, `macos-latest` (`darwin/universal`). Attach archives plus `SHA256SUMS.txt` to the GitHub Release.
* Until `go.mod` exists, the test workflow skips instead of failing. A tag without `wails.json` fails the release job.

Code signing is out of scope for the first slice.

## Consequences

### Positive

* Domain logic is gated on every push.
* Four OS/arch artifacts from one tag.
* Contributors can run the same commands locally.

### Negative and risks

* macOS universal and Windows builds need GitHub-hosted macOS/Windows minutes.
* `ubuntu-24.04-arm` availability depends on GitHub; if it disappears, switch Linux arm64 to a documented cross-compile fallback.
* `go test ./...` from the repo root pulls Wails CGO (`main.go`). CI and `make test` use `-tags=nomain` on `.` plus `./internal/...`.

### Neutral

* Workflows: `.github/workflows/test.yml`, `.github/workflows/release.yml`.
