# Architecture decision records

ADRs live in this folder. Number them in order (`001`, `002`, …). Status is one of: Proposed, Accepted, Deprecated, Superseded.

| ID | Title | Status |
|---|---|---|
| [001](001-wails-config-obs-http.md) | Wails is the config window; OBS uses HTTP + WebSocket | Accepted |
| [002](002-obs-nested-http-paths.md) | Nested `/obs` and `/api` routes | Accepted |
| [003](003-json-user-config-dir.md) | Settings JSON under `os.UserConfigDir` | Accepted |
| [004](004-youtube-client-by-api-key.md) | Choose YouTube client by API key; never fall back on invalid key | Accepted |
| [005](005-wails-svelte-ts-layout.md) | Official `svelte-ts` layout; `package main` at repo root | Accepted |
| [006](006-internal-packages-by-surface.md) | Collapse console packages into `youtube`, `youtube/nokey`, `obs` | Accepted |
| [007](007-docs-layout.md) | README, CURSOR.md, `docs/` (ADRs, use cases, wiki) | Accepted |
| [008](008-github-actions-test-and-release.md) | GitHub Actions: tests on push, binaries on tag | Accepted |
| [009](009-obs-http-process-lifetime.md) | OBS HTTP lives with the Wails process | Accepted |
| [010](010-global-interrupt-hotkey.md) | OS-level interrupt hotkey | Accepted |
| [011](011-local-youtube-quota-estimate.md) | Local YouTube Data API quota estimate | Accepted |
| [012](012-os-keychain-api-key.md) | OS keychain for the YouTube API key | Accepted |
| [013](013-github-self-update.md) | GitHub Releases for in-app update | Accepted |
| [014](014-obs-html-one-websocket.md) | OBS pages open one WebSocket | Accepted |
| [015](015-obs-html-reconnect-timer.md) | OBS pages keep one reconnect timer | Accepted |
