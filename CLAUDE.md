# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

1. Source of truth
 1. [CURSOR.md](CURSOR.md) is the full agent spec.
 2. It holds the HTTP/WS contract, settings rules, layout, and working agreements.
 3. Read it before changing routes, payloads, bindings, or package layout.
 4. This file is a short entry point. Do not duplicate CURSOR.md here.
 5. Keep CURSOR.md updated when routes, payloads, stack, or layout change.

2. What the app is
 1. Wails v2 + Svelte/TS desktop port of [HardDie/ytmemchat](https://github.com/HardDie/ytmemchat).
 2. The original console app lives in a local clone at `/Users/oleg/Project/go/ytmemchat`.
 3. Port from it. Do not re-invent YouTube pagination, the mux, or WS payloads.
 4. The Wails window is the control panel only.
 5. OBS is the display, served by a local Go HTTP server.

3. Commands
 1. `make help` lists all targets.
 2. `make dev` runs Wails with hot reload.
 3. `make build` builds into `build/bin`.
 4. `make test` runs unit tests with `-race -tags=nomain` (same as CI).
 5. `make test-integration` runs `-tags=integration` over `./internal/...`.
 6. `make vet` and `make fmt` (fmt only checks; it fails on unformatted files).
 7. `make generate` regenerates `frontend/wailsjs` after binding method changes.
 8. `make doc-all PKG=./internal/<pkg>` checks godoc.
 9. `make screenshots` refreshes `docs/screenshots/window.gif` after pane UI changes.

4. Running one test
 1. Always pass `-tags=nomain` for unit tests.
 2. Without it, `main.go` pulls in Wails CGO and native pickers.

```bash
go test -race -count=1 -tags=nomain -run TestName ./internal/obs
go test -race -count=1 -tags=nomain -run TestName .
go test -count=1 -tags=integration -run TestName ./internal/tts
```

5. The `nomain` build split
 1. `main.go`, `wails_hooks.go`, and `*/pick.go` are `//go:build !nomain`.
 2. Each has a `*_nomain.go` no-op twin.
 3. `internal/hotkey` OS binders are excluded under `nomain` too.
 4. When adding Wails-runtime or CGO code, add a `nomain` stub alongside it.
 5. `internal/` must not import Wails and stays CGO-free (except hotkey bind files).

6. Runtime architecture
 1. `main.go` binds pane structs from `bindings/<pane>/`.
 2. Panes: `home`, `configuration`, `commands`, `test`, `update`, plus `sidebar`.
 3. Bindings are thin. Shared runtime lives on `App` in `package main`.
 4. `app.go` owns settings and lifecycle. `app_host.go` is the host API bindings call.
 5. `http.go` owns the OBS listener.
 6. OBS HTTP (`internal/obs`) starts in `OnStartup` and lives for the process.
 7. A saved port change restarts the listener.
 8. Start/Stop only runs or cancels the YouTube iterator (`pipeline.go`).
 9. Client choice is by API key: empty → `internal/youtube/nokey`, set → Data API v3.
 10. Never fall back to `nokey` when a key is set, even on errors.
 11. Each chat line goes to the chat WS hub.
 12. `overlay.go` then sends an alert on a command match, else TTS.
 13. Webhook and Test-pane lines enter the same path without Start.

7. Package boundaries
 1. `internal/obs` owns transport: mux, two WS hubs, embedded HTML and `script.js`.
 2. `obs` must not import `alerts` or `youtube`. `app.go` wires them.
 3. `internal/youtube` owns `ChatMessage` and `Client`. No second chat struct.
 4. `internal/alerts` parses `commands.yaml` and matches tokens.
 5. `internal/tts` holds OS drivers (`say`, PowerShell, `espeak`).
 6. `internal/config` stores JSON under `os.UserConfigDir()/ytmemchat`.
 7. `internal/secret` keeps the API key in the OS keychain.
 8. `pkg/` holds reusable code with no app identity (`archive`, `update`, `version`).
 9. `frontend/src/lib` is UI only. No YouTube, HTTP, or YAML logic in Svelte.
 10. Never hand-edit `frontend/wailsjs/`.

8. HTTP contract essentials
 1. OBS pages are `/obs/chat` and `/obs/overlay`, each with a child `…/ws`.
 2. Operator APIs are under `/api/…`.
 3. Do not revive console routes (`/ws`, `/chat`, `/ws_chat`, `/media/`, `/webhook/`).
 4. WS JSON stays stable. Additive fields only.
 5. OBS page behavior rules are ADRs 014–026 in `docs/architecture/`.

9. Change checklist
 1. New core decision → next-numbered ADR plus a row in `docs/architecture/INDEX.md`.
 2. New Go package → package comment, comments on every export, unit tests.
 3. Also add its `go doc` row on `docs/wiki/Contributing.md`.
 4. Integration tests use `//go:build integration`.
 5. OS-specific integration tests also tag GOOS (`integration && darwin`).
 6. New module → use-case file from `docs/use-cases/_TEMPLATE.md`, plus an INDEX row.
 7. User-facing change → update README (short) and the matching `docs/wiki/` page.

10. Docs style
 1. `.cursor/rules/` sets this for CURSOR.md, ADRs, and the wiki.
 2. Short sentences. One numbered title, then short sub-items.
 3. Never insert a sub-item mid-list. Append it to keep numbering stable.
 4. Wiki pages prefer short sentences. Use lists only for real groups.

11. Releases and versions
 1. Pushing a `vMAJOR.MINOR.PATCH` tag triggers `.github/workflows/release.yml`.
 2. `pkg/version.Build` is stamped via ldflags (tag, or short commit).
 3. `make version` writes the last tag into `wails.json` `productVersion`.
 4. `make dev` and `make build` run it first.
 5. Linux builds need `-tags webkit2_41` on webkit2gtk-4.1 systems. The Makefile adds it.
