# ytmemchat (Wails) Development Guidelines

ytmemchat is a YouTube Live companion.

1. It reads live chat.
2. It reacts in realtime: chat overlay, meme alerts, TTS.
3. This repo is a **Wails v2 + Svelte** port of [HardDie/ytmemchat](https://github.com/HardDie/ytmemchat).
4. Domain logic comes from that project: YouTube polling, HTTP overlay, WebSockets, commands, TTS.
5. This app adds a **desktop configuration window**.
6. OBS still uses **HTTP pages + WebSocket events**.

**Two surfaces**

1. **Wails window (Svelte)**
   1. Settings and control only.
   2. API key, stream ID, port, start/stop, TTS/alerts paths, status.
   3. Not the OBS chat renderer.
2. **Local HTTP server (Go)**
   1. What OBS Browser Sources load.
   2. Chat and alerts go over WebSockets.
   3. Not Wails Events.

**Realtime path**

1. Go owns YouTube Live Chat ingestion.
2. API v3 when an API key is set.
3. Otherwise `internal/youtube/nokey`.
4. Both skip history and push into the same pipeline.
5. New messages go to `/obs/chat` and `/obs/chat/ws`.
6. Alerts and TTS go to `/obs/overlay` and `/obs/overlay/ws`.
7. Svelte may show connection state via bound methods.
8. Svelte must not replace the overlay protocol.

**Reference implementation**

1. Local clone: `/Users/oleg/Project/go/ytmemchat`.
2. Prefer copying packages from there.
3. Do not re-invent YouTube pagination, the mux, or WebSocket payloads.

---

## Product scope (this app)

**Must have (first vertical slice)**

1. Configuration window
   1. Persist optional API key, stream/video ID, listen port.
   2. **Start / Stop the YouTube iterator** without restarting the process.
   3. OBS HTTP stays up.
2. HTTP server with nested `/obs/…` Browser Source URLs (see contract below).
3. Live chat HTML page that updates over WebSocket.
4. Connection / quota / error state in the Wails window and logs.
5. Home quota
   1. Show today’s spent Data API units when a key is set.
   2. `liveChatMessages.list` is 5 units.
   3. Persist in `quota.json`.
   4. Reset at midnight Pacific Time.
6. Home copies the two OBS URLs (`/obs/chat`, `/obs/overlay`) for the current port.

**Port from console app (with the HTTP layer, not after a desktop chat UI)**

1. Overlay page for alerts + TTS (`/obs/overlay`, `/obs/overlay/ws`).
2. Alert commands (`commands.yaml`, token prefix such as `@jump`) and `/obs/media/`.
3. Cross-platform TTS for messages that did not trigger an alert.
4. Optional `/api/webhook` and `/api/interrupt`.

**Out of scope unless explicitly requested**

1. Rendering the full OBS chat or overlay inside the Wails webview as the primary display.
2. Replacing YouTube Studio chat.

---

## HTTP / WebSocket contract (this port)

This is a new desktop app.

1. Do **not** keep console root paths (`/`, `/ws`, `/chat`, `/ws_chat`).
2. OBS sources will be re-added from the config window.
3. JSON payloads stay compatible.
4. **URLs are nested by surface**.

**Why the console layout is awkward**

1. `/` is the alert overlay. Easy to confuse with “the app” when two Browser Sources exist.
2. Chat HTML is `/chat`. Its socket is `/ws_chat` (sibling, different naming).
3. Overlay socket is `/ws`, also a sibling of `/`.
4. `/webhook/` and `/interrupt/` sit next to OBS pages even though they are operator APIs.

**Rules**

1. One **page URL per OBS Browser Source**.
2. That page’s WebSocket is a child path `…/ws`.
3. HTML derives the socket from `location`. Never hardcode `/ws_chat`.
4. `/obs/…` is what goes on stream.
5. `/api/…` is for curl/dev tools. Do not mix them.
6. No trailing slash on HTML routes.
7. File servers use a trailing slash (`/obs/media/`).
8. Do not register old console aliases unless we later need a migration shim.

| Method | Route | Role |
|---|---|---|
| GET | `/obs/chat` | Chat overlay HTML (OBS Browser Source) |
| GET | `/obs/chat/ws` | Chat WebSocket |
| GET | `/obs/overlay` | Alert + TTS overlay HTML (OBS Browser Source) |
| GET | `/obs/overlay/ws` | Overlay WebSocket |
| GET | `/obs/media/` | Alert media files (when alerts enabled) |
| GET | `/` | Human index: lists the two OBS URLs (not for OBS) |
| POST | `/api/webhook` | Inject a fake chat message (when webhooks enabled) |
| POST | `/api/interrupt` | Interrupt current TTS |

OBS Browser Sources (example port `8080`):

- Chat: `http://127.0.0.1:8080/obs/chat`
- Overlay: `http://127.0.0.1:8080/obs/overlay`

Chat query `?transparent=1`:

1. Drops the opaque background.
2. Maps onto the existing `body.transparent` class.
3. Overlay still needs Interact-once for audio autoplay.

**Chat WS payload** (same as `ytmemchat/internal/chat/contract.go`, plus control `type`)

1. `authorName`
2. `authorPicture`
3. `messageText`
4. `publishedAt`
5. `isModerator`
6. `isOwner`
7. optional `type`: empty line, `app_closed`, `chat_flush`
8. `chat_flush` clears on-screen rows. Overlay is unchanged.

**Overlay WS payload** (same as `ytmemchat/internal/server/contract.go`, plus `app_closed`)

1. `type`: `alert` \| `tts` \| `tts_interrupt` \| `app_closed`
2. `payload`
3. `filename`
4. `volume`
5. `scale`
6. Overlay media URLs must use `/obs/media/<file>`, not `/media/`.
7. On Wails graceful exit, both sockets get `type: app_closed` before close.
8. OBS pages can show that the app quit. They keep retrying so a later launch reconnects.
9. Unexpected drops still use the generic “disconnected” banner.
10. Overlay has no `chat_flush`.

Do not invent a parallel Wails Events protocol for OBS.

1. Config UI status: bound Go methods.
2. Optional Wails events **only** for window status.
3. Never as the OBS transport.

**Process vs Start**

1. Serve `internal/obs` from Wails `OnStartup`.
2. Keep serving until the process exits, or until a saved port change restarts the listener.
3. OBS pages and WebSockets are available whenever the app is open.
4. That includes before the first Start and after Stop.
5. Start/Run only starts the YouTube iterator.
6. Each new message fans into those sockets (chat hub; alerts vs TTS on the overlay hub).
7. Stop cancels the iterator only.
8. See [ADR 009](docs/architecture/009-obs-http-process-lifetime.md).

---

## Persistent configuration

Wails has **no built-in settings store** in v2. Maintainers point at XDG / the OS config dir. v3 may add helpers.

Idiomatic pattern:

1. **Store in Go**, not in the webview.
   1. `localStorage` is not the source of truth.
   2. Webview data can be wiped.
   3. It is a poor place for an API key.
2. **Path**: `filepath.Join(os.UserConfigDir(), "ytmemchat", "config.json")`
   1. macOS: `~/Library/Application Support/ytmemchat/config.json`
   2. Windows: `%AppData%\ytmemchat\config.json`
   3. Linux: `~/.config/ytmemchat/config.json`
   4. `os.UserConfigDir` is enough.
   5. `adrg/xdg` is optional for stricter XDG on Linux.
3. **Never write next to the binary.**
   1. macOS `.app` / Program Files are not writable after install.
4. **Format**
   1. One JSON document.
   2. Defaults live in code.
   3. Missing file = first launch, use defaults.
   4. Unknown fields ignored (`json` unmarshal).
   5. Include a `version` field if we need migrations later.
5. **Lifecycle**
   1. Load in `OnStartup`.
   2. **Save on every successful settings change** from `SaveSettings`.
   3. Do not rely on `OnShutdown` alone (force-quit / crash skips it).
   4. Atomic write: temp file in the same dir, `fsync`, then `Rename`.
   5. File mode `0600` because the file holds `YOUTUBE_API_KEY`.
6. **Bindings**
   1. One Wails struct per pane under `bindings/` (`home`, `configuration`, `commands`, `test`).
   2. Sidebar chrome uses `bindings/sidebar`.
   3. `App` keeps Start/Stop, OBS HTTP, and persist helpers.
   4. Configuration: `GetSettings`, `SaveSettings`, `ConfigPath`, `GetTTSVoices`, path pickers.
   5. Home: `GetOBSStatus`, `Start`, `Stop`, `GetRunStatus`, `LookupLatestStream`, `InterruptTTS`.
   6. Commands: `GetAlertCommands`, `SaveAlertCommands`, `PickAlertMediaFile`, `PreviewAlert`.
   7. Test: `SendTestMessage`, `FlushChat`.
   8. Sidebar: `AppVersion`.
   9. `PickAlertMediaFile` returns a path relative to the Config media folder.
   10. Files outside that tree are rejected.
   11. `AppVersion` is stamped at link time (`-X github.com/HardDie/ytmemchat_wails/bindings/sidebar.buildVersion`).
   12. Unset `AppVersion` is `dev`.
   13. Start uses last saved settings (window saves the form first).
   14. Empty API key → `nokey`. Non-empty → v3 only.
   15. Never fall back on invalid key.
   16. Connect runs in a goroutine.
   17. After each chat line (YouTube, `POST /api/webhook`, or **Send**): overlay `alert` on a command match, else TTS when enabled.
   18. Inject/test run while OBS HTTP is up. YouTube Start is not required.
   19. Empty `commandsFilePath` skips the matcher.
   20. A bad commands file fails Start (inject logs and skips the matcher).
   21. `SaveAlertCommands` writes `commands.yaml` (omits unset `volume`/`scale`).
   22. Save reloads the matcher without YouTube Start.
   23. Update: `Check`, `Download`, `ApplyAndQuit`.
   24. macOS `CFBundleVersion` and `CFBundleShortVersionString` use the last git tag (no `v`).
   25. That value is Wails `info.productVersion` at `wails build` / `wails dev`.
   26. `make version` writes that tag into `wails.json` (`make dev` / `make build` do this first).
7. **What belongs here**
   1. Stream ID (required to Start).
   2. Optional YouTube API key.
   3. Listen port.
   4. TTS on/off + voice.
   5. Alerts on/off + command token + media/commands paths.
   6. Webhook on/off.
   7. Interrupt hotkey on/off + chord (default `Ctrl+Shift+I`, OS-global).
   8. Optional window size.
   9. The window is setup only (not on stream).
   10. First launch: alerts and TTS off, API key empty.
   11. **Home**: YouTube status, Start/Stop, spent Data API quota when a key is set, interrupt overlay audio, Find latest, OBS URLs.
   12. **Config**: all settings including the interrupt shortcut.
   13. Find latest fills stream ID from the channel of a known video (live, else upcoming; not VOD).
   14. Find latest requires a Data API key. It does not save until Save/Start.
   15. **Commands**: edit `commands.yaml` (add rows; blank volume/scale are not written).
   16. **Test**: send a fake chat line.
   17. **What does not**: chat history, live iterator state.
   18. **Update**: check GitHub, download, quit and replace.
8. **API key**
   1. Optional.
   2. Empty means use `youtube/nokey`.
   3. When set, store in the OS keychain when it is available.
   4. If the keychain is down, keep the key in this `0600` JSON.
   5. Never log the key.
   6. If JSON has a key and the vault works: move it, then save JSON with empty `apiKey`.
   7. If the vault fails: leave the file as-is.
   8. Config shows whether the keychain or the settings file holds the key.

Do not panic if config is missing (unlike console `config.Get()`).

1. Start is allowed without an API key if a stream ID is set.
2. Show a clear “not configured” state when the stream ID is empty.

---

## YouTube clients (key vs no-key)

The console app has two implementations of `youtube.Client` (`GetMessageIterator` → `ChatMessage`).

1. Port **both**.
2. Console `main.go` currently hardcodes v3 (`if true`).
3. This desktop app **selects by whether the API key is set**.

| Config | Client | Package |
|---|---|---|
| API key empty / whitespace-only | No key: poll `live_chat` HTML | `internal/youtube/nokey` (console `youtubev1`) |
| API key non-empty | YouTube Data API v3 | `internal/youtube` |

**Selection**

1. Happens once at Start, after trimming the saved key.
2. Do not inspect key validity by falling through clients.

**Invalid key: no fallback**

1. If a key is present, only the v3 client runs.
2. If YouTube rejects it (typical API `401` / `403` / `invalidApiKey` / disabled API): Stop the iterator, keep using v3.
3. Show a distinct error in the Wails window (e.g. “YouTube API key is invalid”).
4. Do **not** silently start `youtube/nokey`.
5. The user must clear the key (no-key client) or fix the key.

**Local quota estimate**

1. The v3 client wraps its HTTP transport with `internal/youtube/quota` (one call in the constructor).
2. Attach the API key on that client (`WithHTTPClient` skips `WithAPIKey`).
3. Count billed units (`liveChatMessages.list` is 5).
4. Persist `quota.json` next to `config.json` for the Pacific day.
5. Do not add extra Data API or Cloud calls to read remaining quota.
6. Do not wrap `youtube/nokey`.

**Do not treat every v3 failure as a bad key**

1. “Video is not live”, missing `activeLiveChatId`, quota exceeded, and network errors get their own messages.
2. Fallback is still forbidden in all of those cases when a key was provided.

Both clients:

1. Keep emitting the same `ChatMessage` shape.
2. Skip history on connect.
3. The rest of the pipeline (chat WS, alerts, TTS) must not care which client produced the message.

Settings UI:

1. Empty key = no-key client.
2. Filled key = Data API v3 (needs YouTube Data API v3 enabled on that key).

---

## Stack

| Layer | Choice |
|---|---|
| Config UI | [Wails v2](https://wails.io) + Svelte + TypeScript (`wails init -t svelte-ts`) + Vite |
| Overlay / chat | Go `net/http` + gorilla websocket; HTML served like the console app (embed or `internal/*/html`) |
| Backend | Go 1.25+, same domain packages as console ytmemchat |
| YouTube | Data API v3 when an API key is set; otherwise `youtube/nokey`. Same `youtube.Client` interface. |
| Config | JSON file under the OS user config dir (Go `os.UserConfigDir`); settings UI; no committed `.env` |

Scaffold:

1. The tree already has `wails.json`, `main.go`, `app.go`, and `frontend/` from `wails init -t svelte-ts`.
2. Do not run init again (it would nest a second project).
3. After changing exported pane binding methods: `make generate`.

---

## Build & Run (macOS)

Prerequisites:

1. Go
2. Node/npm
3. Wails CLI (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)
4. Xcode CLT

```bash
make help   # all developer targets
make dev    # wails dev
make build  # wails build for this OS (AppVersion + last git tag into CFBundle)
make test
make test-integration
```

Raw equivalents:

1. `wails dev`
2. `wails build`
3. `go test -race -tags=nomain .`
4. `go test -race -tags=nomain ./bindings/...`
5. `go test -race -tags=nomain ./internal/...`
6. Frontend-only (from `frontend/`): `npm install`, then `npm run dev` / `npm run build`.
7. Wails generates bindings under `frontend/wailsjs/`. Do not edit those files by hand.
8. After changing exported pane binding methods: `make generate`.

OBS after start:

1. YouTube Data API v3 is only required when the user saves an API key.
2. Add OBS Browser Sources to `http://127.0.0.1:<port>/obs/chat` and `http://127.0.0.1:<port>/obs/overlay`.
3. Click Interact on the overlay source once so the browser can autoplay audio.

---

## Where things live

This file stays lean.

1. **[README.md](README.md)** is the short user entry (what the app is, install, OBS URLs, status).
2. Operator how-tos are the [wiki](https://github.com/HardDie/ytmemchat_wails/wiki) (`docs/wiki/`).
3. This file is the agent/implementation spec.
4. Keep them consistent. Do not let either rot.

| If you need… | Read |
|---|---|
| What the product is, install, OBS URLs, status | **[README.md](README.md)** — keep short |
| Step-by-step operator guides | **[docs/wiki](docs/wiki/Home.md)** (copy to [GitHub wiki](https://github.com/HardDie/ytmemchat_wails/wiki); page links omit `.md`, screenshots use raw `main` URLs) |
| Architecture decisions (ADRs) | **[docs/architecture](docs/architecture/INDEX.md)** |
| Use cases (after a module is ported) | **[docs/use-cases](docs/use-cases/INDEX.md)** |
| Console behavior and historical overlay URLs | [HardDie/ytmemchat README](https://github.com/HardDie/ytmemchat) and local `../ytmemchat` |
| Persisted settings JSON | `internal/config` |
| HTTP mux, overlay + chat WS, OBS HTML | `internal/obs` (console: `internal/server` + `internal/chat`) |
| Normalized chat event + v3 iterator | `internal/youtube` (console: `internal/clients/youtube`) |
| Local Data API unit estimate | `internal/youtube/quota` (HTTP wrap on the v3 client only; no extra Google calls) |
| No-key live chat client | `internal/youtube/nokey` (console: `internal/clients/youtubev1`) |
| Alert token matching + `commands.yaml` | `internal/alerts` |
| TTS drivers | `internal/tts` |
| Wails project layout / bindings | [Wails first project](https://wails.io/docs/gettingstarted/firstproject/) |

### Intended tree (after scaffold + port)

Official Wails layout:

1. **Vite Svelte app in `frontend/`**.
2. **`package main` at the repo root**.
3. Domain code in `internal/` (Go convention).
4. Do **not** switch to SvelteKit unless we need file-based routing. The config window does not.
5. Overlay HTML stays in Go embeds, not in `frontend/`.
6. OBS loads the HTTP server, not the Wails webview.

```text
.
├── Makefile                  # make help / dev / build / test / screenshots / doc
├── README.md                 # users: product, install, OBS, status
├── CURSOR.md                 # agents: contracts and layout
├── docs/
│   ├── architecture/         # ADRs
│   ├── screenshots/          # README window.gif (`make screenshots`)
│   ├── use-cases/            # UC files only after the module is ported
│   └── wiki/                 # operator how-tos (GitHub wiki source)
├── .github/workflows/
│   ├── test.yml              # go test on every push/PR
│   └── release.yml           # binaries on v* tags
├── wails.json                # Wails CLI; `info.productVersion` = last git tag
├── go.mod
├── main.go                   # wails.Run, Bind pane structs, //go:embed all:frontend/dist
├── app.go                    # App core: settings, pipeline, OBS HTTP (not bound directly)
├── bindings/                 # Wails Bind[] wrappers
│   ├── sidebar/              # AppVersion
│   ├── home/
│   ├── configuration/
│   ├── commands/
│   ├── test/
│   └── update/
├── frontend/                 # config window only (official svelte-ts template)
│   ├── index.html
│   ├── package.json
│   ├── vite.config.ts
│   ├── svelte.config.js
│   ├── src/
│   │   ├── main.ts           # boot App; ?screenshot=1 loads stub bindings
│   │   ├── screenshotBridge.ts
│   │   ├── App.svelte        # pane shell (home / config / commands / test / update)
│   │   ├── style.css
│   │   ├── lib/              # HomePane, ConfigPane, CommandsPane, TestPane, UpdatePane
│   │   └── assets/
│   ├── wailsjs/              # generated bindings — do not edit
│   │   ├── go/               # sidebar, home, configuration, commands, test, update
│   │   └── runtime/
│   └── dist/                 # Vite build output (embedded; gitignore contents)
├── internal/
│   ├── config/               # JSON under UserConfigDir
│   ├── youtube/              # ChatMessage, Client, Data API v3
│   │   ├── quota/            # local Data API unit estimate (no extra Google calls)
│   │   └── nokey/            # former youtubev1 (no API key)
│   ├── obs/                  # one HTTP server: /obs/*, /api/*, both WS hubs, embed HTML
│   ├── hotkey/               # chord parse + OS bind (`!nomain`)
│   ├── alerts/               # command match + media file server
│   ├── tts/
│   ├── secret/               # OS keychain for the YouTube API key
│   └── update/               # GitHub release check + verified install
├── pkg/                      # only if something is useful outside this module (prefer not)
└── build/
```

**Layout rules**

1. Keep `main.go` / `app.go` at the **repository root**.
2. The Wails CLI expects `wails.json` beside them.
3. Do not move the entrypoint to `cmd/` the way the console app does.
4. Keep `App` in `package main` as the desktop core.
5. Bind pane structs under `bindings/` (`home`, `configuration`, `commands`, `test`).
6. Bind `sidebar` for chrome that is not a pane (`AppVersion`).
7. Wrappers may own pane-only Wails methods. Shared runtime stays on `App`.
8. Do not put YouTube/HTTP/config core logic in `bindings/` except pane-specific mapping.
9. Svelte imports Go via `../wailsjs/go/<package>/<Struct>`.
10. Default `wailsjsdir` is `frontend/`. Leave it unless we add SvelteKit.
11. `frontend/src/lib` is UI only. No HTTP server, no YouTube, no `commands.yaml` parsing.
12. OBS pages (`overlay.html`, `chat.html`) live in `internal/obs` next to the handlers (`//go:embed`).
13. Tests sit beside the Go package they cover (`internal/alerts/find_token_test.go` style).
14. `internal/secret` holds the YouTube API key vault. `config.Store` calls it.
15. `internal/update` checks GitHub Releases and stages a verified install.

### Go package documentation (godoc)

Every Go package in this repo must have documentation that `go doc` can print. Treat that as part of the port, not a follow-up.

**Required**

1. Package comment on one file in the package (prefer `doc.go` if the comment would be buried): `// Package youtube …`.
2. Say what the package is for, not how files are named.
3. Every **exported** type, func, method, const, and var has a comment.
4. Start with the name (`// Client is …`, `// GetMessageIterator returns …`).
5. Document error behavior that callers must handle (invalid API key vs not live vs network).
6. Interfaces document each method’s contract on the interface or on a nearby type comment.

**Do not**

1. Leave `TODO` or empty `// Foo …` on exports.
2. Duplicate the console app’s undocumented exports — add comments while porting.
3. Put user install steps in godoc; that belongs in README.

**How to check** (from the repo root, after the package exists):

```bash
make doc PKG=./internal/tts
make doc-all PKG=./internal/tts
```

`-all` must show the package comment and every export.

1. If `go doc` prints `no Go files` or only a name with no prose, the port is incomplete.
2. After porting a package, add (or mark Ported) its `go doc` command on [Contributing](docs/wiki/Contributing.md).
3. Put that row under Package documentation.

### Tests

Every ported Go package **must** have unit tests.

1. Add integration tests when the package talks to the network, disk, OS TTS, or HTTP.
2. Not for pure functions.

| Kind | Files | Build tag | Local command |
|---|---|---|---|
| Unit | `foo_test.go` next to the code | `nomain` for `.`, bindings, and `internal/` | `make test` |
| Integration (all OS) | `foo_integration_test.go` | `//go:build integration` | `make test-integration` |
| Integration (one OS) | `foo_integration_darwin_test.go` (or `_linux_`, `_windows_`) | `//go:build integration && darwin` (or `linux` / `windows`) | same; other GOOS never compile those files |

**Rules**

1. Prefer table-driven tests.
2. Cover happy path and the error cases in the use-case file (invalid API key, not live, bad YAML).
3. `internal/` stays **CGO-free** except OS hotkey bind files.
4. Do not import Wails from `internal/`.
5. OS hotkey registration lives in `internal/hotkey` (`!nomain && !integration`).
6. Tests use `-tags=nomain` so the binder is a no-op (no display).
7. Integration tests use `httptest`, temp dirs, and fakes.
8. They **skip** (`t.Skip`) if a real YouTube key or live stream is required.
9. They must not fail CI for missing secrets. Never commit API keys.
10. **OS-specific integration tests use GOOS build tags**, not `runtime.GOOS` skip branches.
11. Example: `//go:build integration && darwin` so Linux/Windows do not compile `say` tests.
12. Skip only when this OS is correct but the tool is missing (`espeak` not installed on Linux CI).
13. Shared helpers may use `//go:build integration` with no GOOS.
14. A package with no integration surface (pure token matching) does not need an integration file.
15. Say so in the package comment if it is unclear.
16. Porting is incomplete without tests, godoc, a use-case file, and a wiki `go doc` row ([Contributing](docs/wiki/Contributing.md)).

CI (every push and pull request):

1. `go test -race -tags=nomain .`
2. `go test -race -tags=nomain ./bindings/...`
3. `go test -race -tags=nomain ./internal/...`
4. `go test -tags=integration ./internal/...`
5. See [ADR 008](docs/architecture/008-github-actions-test-and-release.md).
6. The `nomain` tag skips `main.go` (Wails CGO) so App/pipeline tests run on Ubuntu.
7. It also skips OS hotkey CGO in `internal/hotkey`.

### Releases

Push a tag `vMAJOR.MINOR.PATCH` (for example `v0.1.0`). GitHub Actions builds and attaches:

| Artifact | Runner / platform |
|---|---|
| `ytmemchat-vMAJOR.MINOR.PATCH-linux-amd64.tar.gz` | `ubuntu-latest` · `linux/amd64` |
| `ytmemchat-vMAJOR.MINOR.PATCH-linux-arm64.tar.gz` | `ubuntu-24.04-arm` · `linux/arm64` |
| `ytmemchat-vMAJOR.MINOR.PATCH-windows-amd64.zip` | `windows-latest` · `windows/amd64` |
| `ytmemchat-vMAJOR.MINOR.PATCH-darwin-universal.zip` | `macos-latest` · `darwin/universal` |

Plus `SHA256SUMS.txt`. Code signing is not part of the first slice.

### Internal packages (optimized vs console)

The console tree splits more packages than behaviors.

1. HTTP: `server` + `chat`.
2. YouTube: `clients/youtube` + `clients/youtubev1`.
3. A one-handler `webhook` package.
4. Collapse by **surface**, not by file count.

| Console | This app |
|---|---|
| `internal/clients/youtube` + `youtubev1` | `internal/youtube` (types + v3) and `internal/youtube/nokey` |
| `internal/server` + `internal/chat` | `internal/obs` — one mux, two hubs, both HTML pages |
| `internal/webhook` | handlers on `obs` (`/api/webhook`, `/api/interrupt`) |
| `pkg/watermill` + `cmd/main.go` fan-in | `App.Start` / `App.Stop` in `app.go` (plain goroutines). Add a bus later only if fan-in gets messy |
| `pkg/logger` | `log/slog` (optional thin wrapper later) |

**Rules**

1. **`youtube` owns `ChatMessage` and `Client`.** Both implementations return that type. Nothing else defines a second chat struct.
2. **`obs` owns transport** (listen, routes, WebSocket JSON, embeds).
3. Alerts/TTS/youtube do not start listeners.
4. Overlay payload types live next to the overlay hub.
5. `alerts` and `tts` send structs. They do not import `net/http`.
6. **Do not give every file a package.** `webhook` is two POST handlers.
7. Duplicate WS upgrade/broadcast in console `server` and `chat` becomes one small hub type used twice.
8. **No `internal/clients/` folder** — there is only YouTube.
9. **Import direction:** `app.go` → config, youtube, obs, alerts, tts, hotkey.
10. `alerts` / `tts` may depend on overlay event types in `obs`.
11. `obs` must not import `alerts` or `youtube` (register handlers from `app.go`).
12. **Keep `alerts` and `tts` as their own packages.** They have real logic (YAML commands, OS voices). Do not dump them into `obs`.
13. `config` may use `internal/secret` for the YouTube API key.
14. `app.go` does not import `internal/update`; the Update pane binding does.

---

## Key facts to keep in mind

1. **OBS is the display; Wails is the control panel.**
   1. Do not build a second chat renderer in Svelte as the source of truth.
   2. Copy overlay and chat HTML/JS from the console app.
   3. Point sockets and media at `/obs/…` (relative `…/ws` from the page URL).
2. **Go is the source of truth for chat.**
   1. Svelte calls bound methods (`SaveSettings`, `Start`, `Stop`, getters for status).
   2. The YouTube iterator publishes into the same pipeline as the console app.
   3. Raw message → chat WS. Alerts vs TTS → overlay WS.
3. **Keep the WebSocket JSON stable.**
   1. Additive fields are OK.
   2. Renaming or dropping `authorName` / overlay `type` is not.
   3. HTTP paths in this repo follow `/obs/…` and `/api/…`, not the console root layout.
4. **One live iterator per session.**
   1. Start must cancel the previous context.
   2. History skip and polling interval stay in the YouTube iterator, not in Svelte.
5. **`app.go` is a thin façade.**
   1. Bindings are exported and JSON-friendly.
   2. They must not block the UI thread on YouTube HTTP.
   3. Long work runs in goroutines.
   4. The OBS HTTP server starts with the Wails process (`OnStartup`) and stops on app exit, not on Start/Stop.
   5. Start/Stop is the YouTube iterator plus fan-out into already-open sockets.
   6. Vite HMR must not own the listener.
6. **Port packages, do not rewrite blindly.**
   1. Bring YouTube iterators, overlay/chat HTML, alerts, and TTS over.
   2. Then **reshape** into `youtube` / `youtube/nokey` / `obs`.
   3. Do not copy `clients/`, a second HTTP package, or watermill unless Start/Stop actually needs a bus.
7. **Config is a JSON file under `os.UserConfigDir()/ytmemchat`.**
   1. Loaded at startup and saved on change.
   2. Defaults live in Go.
   3. Stream ID is required to Start. API key is optional.
   4. Do not panic the process the way console `config.Get()` does.
8. **YouTube client is chosen by API key presence**, not by a hardcoded flag.
   1. Empty key → `youtube/nokey`.
   2. Non-empty key → API v3 only.
   3. Invalid key → error in the config window, never fall back to the no-key client.
9. **Svelte has no overlay business rules.** No YouTube calls, no WS servers, no alert matching in the frontend.
10. **Wails bindings (`frontend/wailsjs`) are generated.**
    1. After changing exported methods on pane binding structs, regenerate via `wails dev` / `wails generate module`.
    2. Never hand-edit `wailsjs`.
11. **Every Go package has godoc** (`go doc` / `go doc -all`). Missing package or export comments are incomplete ports.
12. **Secrets and media paths** stay local.
    1. Do not log API keys.
    2. Alert media files are user-configured paths (`ALERTS_MEDIA_PATH` in the console app).

---

## Agent working agreements

1. Match existing Go style in the console repo (`slog`, small `internal/` packages, interfaces at the package boundary).
2. Keep this file updated when routes, payloads, stack, or directory layout change.
3. **Short prose in this file and ADRs.**
   1. Never pack many facts into one long sentence.
   2. Use a numbered title, then short sub-items.
4. **Keep [README.md](README.md) short.**
   1. Update it in the same change when product status, install, OBS URLs, license, CI, or release artifacts change.
   2. README follows [Make a README](https://www.makeareadme.com/).
   3. Include name, description, install, usage, license, and honest **project status**.
   4. Do not put operator how-tos here (alerts YAML, HTTP curl, API quota, config path, TTS engines).
   5. Those live in [docs/wiki](docs/wiki/Home.md).
   6. Contributor make targets and the package `go doc` table live on [Contributing](docs/wiki/Contributing.md).
   7. Do not dump this file into the README. Deep contracts stay here.
5. **Keep [docs/wiki](docs/wiki/Home.md) in the same change** when operator steps or settings move.
   1. Stream ID, no-key chat, default alerts/TTS, OBS URLs, alert-before-TTS, API key/quota, Commands pane.
6. **Window pane GIF.**
   1. After changing Svelte panes (`App.svelte`, `frontend/src/lib/*`, `style.css`), refresh locally with `make screenshots`.
   2. `README.md` loads it from `docs/screenshots/window.gif`.
   3. Do not hand-edit the GIF.
7. **Use cases only after porting.**
   1. When a module is first added under `internal/` (or `app.go` for UC-11), write a use-case file.
   2. Path: `docs/use-cases/<module>/uc-NN-….md` from `docs/use-cases/_TEMPLATE.md`.
   3. Set the row to Ported in `docs/use-cases/INDEX.md`.
   4. Do not invent UC files for code that is not in this repo.
8. **Godoc on every Go package.**
   1. Package comment plus comments on all exports.
   2. After porting, add a `go doc ./…` row for that package on [Contributing](docs/wiki/Contributing.md) (Package documentation).
   3. Verify with `go doc -all` before considering the port done.
9. **Tests on every ported package.**
   1. Unit tests always.
   2. Integration tests (`//go:build integration`) when the package hits HTTP, disk, or OS APIs.
   3. Platform-specific integration files also tag GOOS (`integration && darwin`).
   4. Keep `internal/` CGO-free. CI must stay green.
10. **New core decisions get an ADR** in `docs/architecture/`.
    1. Use the next number.
    2. Update `docs/architecture/INDEX.md`.
    3. Do not leave accepted decisions only in chat.
11. Prefer the smallest change that ports one behavior correctly over a large rewrite.
12. Treat HTTP + WebSocket overlay/chat as core, not a follow-up. Do not “simplify” by moving chat into the Wails window.
13. Do not revive console routes (`/`, `/ws`, `/chat`, `/ws_chat`, `/media/`, `/webhook/`) as aliases without an explicit migration request.
14. If an API key is set, never fall back to `youtube/nokey` on error.
