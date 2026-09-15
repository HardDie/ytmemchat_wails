# ytmemchat (Wails) Development Guidelines

ytmemchat is a YouTube Live companion: it reads live chat and reacts in realtime (on-stream chat overlay, meme alerts, TTS). This repo is a **Wails v2 + Svelte** port of [HardDie/ytmemchat](https://github.com/HardDie/ytmemchat). Domain logic (YouTube polling, HTTP overlay, WebSockets, command matching, TTS) comes from that project. The new piece is a **desktop configuration window**; OBS still consumes **HTTP pages + WebSocket events**, same as the console app.

**Two surfaces**

1. **Wails window (Svelte)** — settings and control only: API key, stream ID, server port, start/stop, TTS/alerts paths, status/errors. It is not the OBS chat renderer.
2. **Local HTTP server (Go)** — what OBS Browser Sources load. Chat and alerts are pushed over WebSockets, not Wails Events.

**Realtime path**: Go owns YouTube Live Chat ingestion (API v3 when an API key is set, otherwise `internal/youtube/nokey`). Both skip history and push into the same pipeline. New messages go to the chat hub (`/obs/chat` + `/obs/chat/ws`). Alerts and TTS go to the overlay hub (`/obs/overlay` + `/obs/overlay/ws`). The Svelte app may show connection state via bound methods; it must not replace the overlay protocol.

**Reference implementation**: `/Users/oleg/Project/go/ytmemchat` (local clone). Prefer copying packages from there over re-inventing YouTube pagination, the mux, or WebSocket payloads.

---

## Product scope (this app)

**Must have (first vertical slice)**

- Configuration window: persist optional API key, stream/video ID, listen port; **Start / Stop the YouTube iterator** without restarting the process (OBS HTTP stays up).
- HTTP server with the nested `/obs/…` Browser Source URLs (see contract below).
- Live chat HTML page that updates over WebSocket as messages arrive.
- Connection / quota / error state visible in the Wails window (and logs).
- Config UI copies the two OBS URLs (`/obs/chat`, `/obs/overlay`) for the current listen port.

**Port from console app (with the HTTP layer, not after a desktop chat UI)**

- Overlay page for alerts + TTS (`/obs/overlay`, `/obs/overlay/ws`).
- Alert commands (`commands.yaml`, token prefix such as `@jump`) and `/obs/media/`.
- Cross-platform TTS for messages that did not trigger an alert.
- Optional `/api/webhook` and `/api/interrupt` for injecting messages and stopping TTS.

**Out of scope unless explicitly requested**

- Rendering the full OBS chat or overlay inside the Wails webview as the primary display.
- Replacing YouTube Studio chat.

---

## HTTP / WebSocket contract (this port)

This is a new desktop app, so we do **not** keep the console root paths (`/`, `/ws`, `/chat`, `/ws_chat`). OBS sources will be re-added from the config window. JSON payloads stay compatible; **URLs are nested by surface**.

**Why the console layout is awkward**

- `/` is the alert overlay — easy to confuse with “the app” when two Browser Sources exist.
- Chat HTML lives at `/chat` but its socket is `/ws_chat` (sibling, different naming).
- Overlay socket is `/ws`, also a sibling of `/`.
- `/webhook/` and `/interrupt/` sit next to OBS pages even though they are operator APIs.

**Rules**

- One **page URL per OBS Browser Source**. That page’s WebSocket is a child path `…/ws` (HTML derives it from `location`, never a hardcoded `/ws_chat`).
- `/obs/…` is what goes on stream. `/api/…` is for curl/dev tools. Do not mix them.
- No trailing slash on HTML routes. File servers use a trailing slash (`/obs/media/`).
- Do not register the old console aliases unless we later need a migration shim.

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

Optional query on chat: `?transparent=1` to drop the opaque background (map onto the existing `body.transparent` class). Overlay still needs Interact-once for audio autoplay.

**Chat WS payload** (same as `ytmemchat/internal/chat/contract.go`): `authorName`, `authorPicture`, `messageText`, `publishedAt`, `isModerator`, `isOwner`.

**Overlay WS payload** (same as `ytmemchat/internal/server/contract.go`, plus `app_closed`): `type` (`alert` \| `tts` \| `tts_interrupt` \| `app_closed`), `payload`, `filename`, `volume`, `scale`. Overlay media URLs must use `/obs/media/<file>`, not `/media/`. On Wails graceful exit, both sockets get `type: app_closed` before close so OBS pages can show that the app quit; they keep retrying so a later launch reconnects. Unexpected drops still use the generic “disconnected” banner. Chat JSON may include optional `type` (`app_closed` only).

Do not invent a parallel Wails Events protocol for OBS. If the config UI needs a status line, use bound Go methods (and optional Wails events **only** for window status — never as the OBS transport).

**Process vs Start:** Serve `internal/obs` from Wails `OnStartup` until the process exits (or the listen port is saved and the listener is restarted). OBS pages and WebSockets are available whenever the app is open, including before the first Start and after Stop. The Start/Run binding only starts the YouTube iterator and fans each new message into those sockets (chat hub; alerts vs TTS on the overlay hub). Stop cancels the iterator only. See [ADR 009](docs/architecture/009-obs-http-process-lifetime.md).

---

## Persistent configuration

Wails has **no built-in settings store** in v2 (maintainers point at XDG / the OS config dir; v3 may add helpers). Idiomatic pattern:

1. **Store in Go**, not in the webview. `localStorage` is not the source of truth (webview data can be wiped; it is a poor place for an API key).
2. **Path**: `filepath.Join(os.UserConfigDir(), "ytmemchat", "config.json")`  
   - macOS: `~/Library/Application Support/ytmemchat/config.json`  
   - Windows: `%AppData%\ytmemchat\config.json`  
   - Linux: `~/.config/ytmemchat/config.json`  
   `os.UserConfigDir` is enough; `adrg/xdg` is optional if we want stricter XDG on Linux.
3. **Never write next to the binary** (macOS `.app` / Program Files are not writable after install).
4. **Format**: one JSON document with defaults in code. Missing file = first launch, use defaults. Unknown fields ignored (`json` unmarshal). Include a `version` field if we need migrations later.
5. **Lifecycle**: load in `OnStartup`. **Save on every successful settings change** from `SaveSettings` — do not rely on `OnShutdown` alone (force-quit / crash skips it). Atomic write: temp file in the same dir, `fsync`, then `Rename`. File mode `0600` because the file holds `YOUTUBE_API_KEY`.
6. **Bindings**: `GetSettings` / `SaveSettings`, `GetOBSStatus`, `Start` / `Stop` / `GetRunStatus`. Start uses saved settings (window saves the form first). Empty API key → `nokey`; non-empty → v3 only. Never fall back on invalid key. Connect runs in a goroutine. Alerts/TTS are not started yet.
7. **What belongs here**: stream ID (required to Start), optional YouTube API key, listen port, TTS on/off + voice, alerts on/off + command token + media/commands paths, webhook on/off, optional window size. **What does not**: chat history, live iterator state.
8. **API key**: optional. Empty means use `youtube/nokey`. When set, store in this `0600` JSON. OS keychain is a later hardening step. Never log the key.

Do not panic if config is missing (unlike console `config.Get()`). Start is allowed without an API key if a stream ID is set. Show a clear “not configured” state when the stream ID is empty.

---

## YouTube clients (key vs no-key)

The console app has two implementations of `youtube.Client` (`GetMessageIterator` → `ChatMessage`). Port **both**. The console `main.go` currently hardcodes v3 (`if true`); this desktop app **selects by whether the API key is set**.

| Config | Client | Package |
|---|---|---|
| API key empty / whitespace-only | No key: poll `live_chat` HTML | `internal/youtube/nokey` (console `youtubev1`) |
| API key non-empty | YouTube Data API v3 | `internal/youtube` |

**Selection happens once at Start**, after trimming the saved key. Do not inspect key validity by falling through clients.

**Invalid key: no fallback.** If a key is present, only the v3 client runs. If YouTube rejects it (typical API `401` / `403` / `invalidApiKey` / disabled API), **Stop the iterator, keep using v3, and show a distinct error in the Wails window** (e.g. “YouTube API key is invalid”). Do **not** silently start `youtube/nokey`. The user must clear the key (to use the no-key client) or fix the key.

**Do not treat every v3 failure as a bad key.** “Video is not live”, missing `activeLiveChatId`, quota exceeded, and network errors get their own messages. Fallback is still forbidden in all of those cases when a key was provided.

Both clients must keep emitting the same `ChatMessage` shape and skip history on connect. The rest of the pipeline (chat WS, alerts, TTS) must not care which client produced the message.

The settings UI should make the optional key obvious: empty = no-key client; filled = Data API v3 (needs YouTube Data API v3 enabled on that key).

---

## Stack

| Layer | Choice |
|---|---|
| Config UI | [Wails v2](https://wails.io) + Svelte + TypeScript (`wails init -t svelte-ts`) + Vite |
| Overlay / chat | Go `net/http` + gorilla websocket; HTML served like the console app (embed or `internal/*/html`) |
| Backend | Go 1.25+, same domain packages as console ytmemchat |
| YouTube | Data API v3 when an API key is set; otherwise `youtube/nokey`. Same `youtube.Client` interface. |
| Config | JSON file under the OS user config dir (Go `os.UserConfigDir`); settings UI; no committed `.env` |

The tree already has `wails.json`, `main.go`, `app.go`, and `frontend/` from `wails init -t svelte-ts`. Do not run init again (it would nest a second project). After changing exported `App` methods: `make generate`.

---

## Build & Run (macOS)

Prerequisites: Go, Node/npm, Wails CLI (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`), Xcode CLT.

```bash
make help   # all developer targets
make dev    # wails dev
make build  # wails build for this OS
make test
make test-integration
```

Raw Wails/Go equivalents: `wails dev`, `wails build`, `go test -race -tags=nomain .` then `go test -race ./internal/...`. Frontend-only (from `frontend/`): `npm install` then `npm run dev` / `npm run build`. Wails generates bindings under `frontend/wailsjs/` — do not edit those files by hand. After changing exported `App` methods: `make generate`.

YouTube Data API v3 is only required when the user saves an API key. After start, add OBS Browser Sources to `http://127.0.0.1:<port>/obs/chat` and `http://127.0.0.1:<port>/obs/overlay`. Click Interact on the overlay source once so the browser can autoplay audio (same as the console README).

---

## Where things live

This file stays lean. **[README.md](README.md)** is the user-facing entry (what the app is, install, OBS URLs). This file is the agent/implementation spec. Keep them consistent; do not let either rot.

| If you need… | Read |
|---|---|
| What the product is, install, OBS setup, test API (users) | **[README.md](README.md)** — keep current |
| Architecture decisions (ADRs) | **[docs/architecture](docs/architecture/INDEX.md)** |
| Use cases (after a module is ported) | **[docs/use-cases](docs/use-cases/INDEX.md)** |
| Console behavior and historical overlay URLs | [HardDie/ytmemchat README](https://github.com/HardDie/ytmemchat) and local `../ytmemchat` |
| Persisted settings JSON | `internal/config` |
| HTTP mux, overlay + chat WS, OBS HTML | `internal/obs` (console: `internal/server` + `internal/chat`) |
| Normalized chat event + v3 iterator | `internal/youtube` (console: `internal/clients/youtube`) |
| No-key live chat client | `internal/youtube/nokey` (console: `internal/clients/youtubev1`) |
| Alert token matching + `commands.yaml` | `internal/alerts` |
| TTS drivers | `internal/tts` |
| Wails project layout / bindings | [Wails first project](https://wails.io/docs/gettingstarted/firstproject/) |

### Intended tree (after scaffold + port)

Official Wails layout is a **Vite Svelte app in `frontend/`** plus **`package main` at the repo root**. Domain code goes in `internal/` (Go convention). Do **not** switch to SvelteKit unless we have a real need for file-based routing — the config window does not. Overlay HTML stays in Go embeds, not in `frontend/`, because OBS loads the HTTP server, not the Wails webview.

```text
.
├── Makefile                  # make help / dev / build / test / doc
├── README.md                 # users: product, install, OBS, status
├── CURSOR.md                 # agents: contracts and layout
├── docs/
│   ├── architecture/         # ADRs
│   └── use-cases/            # UC files only after the module is ported
├── .github/workflows/
│   ├── test.yml              # go test on every push/PR
│   └── release.yml           # binaries on v* tags
├── wails.json                # Wails CLI: frontend install/build/dev
├── go.mod
├── main.go                   # wails.Run, //go:embed all:frontend/dist
├── app.go                    # App struct: GetSettings, SaveSettings, Start, Stop
├── frontend/                 # config window only (official svelte-ts template)
│   ├── index.html
│   ├── package.json
│   ├── vite.config.ts
│   ├── svelte.config.js
│   ├── src/
│   │   ├── main.ts           # mount App (Svelte 5: mount(), not new App())
│   │   ├── App.svelte        # settings + start/stop + status
│   │   ├── app.css
│   │   ├── lib/              # small Svelte pieces (SettingsForm, StatusBar, CopyUrl)
│   │   └── assets/
│   ├── wailsjs/              # generated bindings — do not edit
│   │   ├── go/main/          # App.ts / App.js from package main
│   │   └── runtime/
│   └── dist/                 # Vite build output (embedded; gitignore contents)
├── internal/
│   ├── config/               # JSON under UserConfigDir
│   ├── youtube/              # ChatMessage, Client, Data API v3
│   │   └── nokey/            # former youtubev1 (no API key)
│   ├── obs/                  # one HTTP server: /obs/*, /api/*, both WS hubs, embed HTML
│   ├── alerts/               # command match + media file server
│   └── tts/
├── pkg/                      # only if something is useful outside this module (prefer not)
└── build/
```

**Layout rules**

- Keep `main.go` / `app.go` at the **repository root**. The Wails CLI expects `wails.json` beside them. Do not move the entrypoint to `cmd/` the way the console app does.
- Bind **one façade** (`App` in `package main`). Put YouTube, HTTP, and config in `internal/` and call them from `app.go`. Extra bind structs only if a second UI surface needs its own API.
- Svelte imports Go via `../wailsjs/go/main/App` (or `$lib` aliases that point at `wailsjs`). Default `wailsjsdir` is `frontend/` — leave it unless we add SvelteKit.
- `frontend/src/lib` is UI only. No HTTP server, no YouTube, no `commands.yaml` parsing.
- OBS pages (`overlay.html`, `chat.html`) live in `internal/obs` next to the handlers (`//go:embed`), not under `frontend/`.
- Tests sit beside the Go package they cover (`internal/alerts/find_token_test.go` style).

### Go package documentation (godoc)

Every Go package in this repo must have documentation that `go doc` can print. Treat that as part of the port, not a follow-up.

**Required**

- Package comment on one file in the package (prefer `doc.go` if the comment would be buried): `// Package youtube …` as a full paragraph. Say what the package is for, not how files are named.
- Every **exported** type, func, method, const, and var has a comment. Start with the name (`// Client is …`, `// GetMessageIterator returns …`).
- Document error behavior that callers must handle (invalid API key vs not live vs network).
- Interfaces document each method’s contract on the interface or on a nearby type comment.

**Do not**

- Leave `TODO` or empty `// Foo …` on exports.
- Duplicate the console app’s undocumented exports — add comments while porting.
- Put user install steps in godoc; that belongs in README.

**How to check** (from the repo root, after the package exists):

```bash
make doc PKG=./internal/tts
make doc-all PKG=./internal/tts
```

`-all` must show the package comment and every export. If `go doc` prints `no Go files` or only a name with no prose, the port is incomplete.

After porting a package, add (or mark Ported) its `go doc` command in [README.md](README.md) under Contributing → Package documentation.

### Tests

Every ported Go package **must** have unit tests. Add integration tests when the package talks to the network, disk, OS TTS, or HTTP — not for pure functions.

| Kind | Files | Build tag | Local command |
|---|---|---|---|
| Unit | `foo_test.go` next to the code | none (`nomain` for package main) | `make test` or `go test ./internal/alerts` |
| Integration (all OS) | `foo_integration_test.go` | `//go:build integration` | `make test-integration` |
| Integration (one OS) | `foo_integration_darwin_test.go` (or `_linux_`, `_windows_`) | `//go:build integration && darwin` (or `linux` / `windows`) | same; other GOOS never compile those files |

**Rules**

- Prefer table-driven tests. Cover happy path and the error cases in the use-case file (invalid API key, not live, bad YAML).
- `internal/` must stay **CGO-free** so CI can test on Ubuntu without WebKit. Do not import Wails from `internal/`.
- Integration tests use `httptest`, temp dirs, and fakes. They **skip** (`t.Skip`) if a real YouTube key or live stream is required; they must not fail CI for missing secrets. Never commit API keys.
- **OS-specific integration tests use GOOS build tags**, not `runtime.GOOS` branches that skip. Example: `//go:build integration && darwin` so Linux/Windows do not compile `say` tests. Skip only when this OS is correct but the tool is missing (`espeak` not installed on Linux CI). Shared helpers may use `//go:build integration` with no GOOS.
- A package with no integration surface (for example pure token matching) does not need an integration file; say so in the package comment if it is unclear.
- Porting is incomplete without tests, godoc, a use-case file, and a README `go doc` row.

CI (every push and pull request): `go test -race -tags=nomain .` then `go test -race ./internal/...` then `go test -tags=integration ./internal/...`. See [ADR 008](docs/architecture/008-github-actions-test-and-release.md). The `nomain` tag skips `main.go` (Wails CGO) so App/pipeline tests run on Ubuntu.

### Releases

Push a tag `vMAJOR.MINOR.PATCH` (for example `v0.1.0`). GitHub Actions builds and attaches:

| Artifact | Runner / platform |
|---|---|
| `ytmemchat-linux-amd64.tar.gz` | `ubuntu-latest` · `linux/amd64` |
| `ytmemchat-linux-arm64.tar.gz` | `ubuntu-24.04-arm` · `linux/arm64` |
| `ytmemchat-windows-amd64.zip` | `windows-latest` · `windows/amd64` |
| `ytmemchat-darwin-universal.zip` | `macos-latest` · `darwin/universal` |

Plus `SHA256SUMS.txt`. Code signing is not part of the first slice.

### Internal packages (optimized vs console)

The console tree splits HTTP into `server` + `chat`, YouTube into `clients/youtube` + `clients/youtubev1`, and a one-handler `webhook` package. That is more packages than behaviors. Collapse by **surface**, not by file count.

| Console | This app |
|---|---|
| `internal/clients/youtube` + `youtubev1` | `internal/youtube` (types + v3) and `internal/youtube/nokey` |
| `internal/server` + `internal/chat` | `internal/obs` — one mux, two hubs, both HTML pages |
| `internal/webhook` | handlers on `obs` (`/api/webhook`, `/api/interrupt`) |
| `pkg/watermill` + `cmd/main.go` fan-in | `App.Start` / `App.Stop` in `app.go` (plain goroutines). Add a bus later only if fan-in gets messy |
| `pkg/logger` | `log/slog` (optional thin wrapper later) |

**Rules**

- **`youtube` owns `ChatMessage` and `Client`.** Both implementations return that type. Nothing else defines a second chat struct.
- **`obs` owns transport** (listen, routes, WebSocket JSON, embeds). Alerts/TTS/youtube do not start listeners. Overlay payload types live next to the overlay hub so `alerts` and `tts` send structs, they do not import `net/http`.
- **Do not give every file a package.** `webhook` is two POST handlers. Duplicate WS upgrade/broadcast in console `server` and `chat` becomes one small hub type used twice (chat payload vs overlay payload).
- **No `internal/clients/` folder** — there is only YouTube.
- **Import direction:** `app.go` → config, youtube, obs, alerts, tts. `alerts` / `tts` may depend on overlay event types in `obs`. `obs` must not import `alerts` or `youtube` (register handlers from `app.go`).
- **Keep `alerts` and `tts` as their own packages** — they have real logic (YAML commands, OS voices). Do not dump them into `obs`.

---

## Key facts to keep in mind

- **OBS is the display; Wails is the control panel.** Do not build a second chat renderer in Svelte as the source of truth. Copy overlay and chat HTML/JS from the console app, then point sockets and media at `/obs/…` as in the contract (relative `…/ws` from the page URL).
- **Go is the source of truth for chat.** Svelte calls bound methods (`SaveSettings`, `Start`, `Stop`, getters for status). The YouTube iterator publishes into the same pipeline as the console app (raw message → chat WS; alerts vs TTS → overlay WS).
- **Keep the WebSocket JSON stable.** Additive fields are OK; renaming or dropping `authorName` / overlay `type` is not. HTTP paths in this repo follow `/obs/…` and `/api/…`, not the console root layout.
- **One live iterator per session.** Start must cancel the previous context. History skip and polling interval stay in the YouTube iterator — not in Svelte.
- **`app.go` is a thin façade.** Bindings are exported, JSON-friendly, and must not block the UI thread on YouTube HTTP. Long work runs in goroutines. The OBS HTTP server starts with the Wails process (`OnStartup`) and stops on app exit, not on Start/Stop. Start/Stop is the YouTube iterator plus fan-out into already-open sockets. Vite HMR must not own the listener.
- **Port packages, do not rewrite blindly.** Bring YouTube iterators, overlay/chat HTML, alerts, and TTS over, then **reshape** into `youtube` / `youtube/nokey` / `obs` as above. Do not copy `clients/`, a second HTTP package, or watermill unless Start/Stop orchestration actually needs a bus.
- **Config is a JSON file under `os.UserConfigDir()/ytmemchat`**, loaded at startup and saved on change. Defaults live in Go. Stream ID is required to Start; API key is optional. Do not panic the process the way console `config.Get()` does.
- **YouTube client is chosen by API key presence**, not by a hardcoded flag. Empty key → `youtube/nokey`. Non-empty key → API v3 only. Invalid key → error in the config window, never fall back to the no-key client.
- **Svelte has no overlay business rules.** No YouTube calls, no WS servers, no alert matching in the frontend.
- **Wails bindings (`frontend/wailsjs`) are generated.** After changing exported methods on `App`, regenerate via `wails dev` / `wails generate module`. Never hand-edit `wailsjs`.
- **Every Go package has godoc** (`go doc` / `go doc -all`). Missing package or export comments are incomplete ports.
- **Secrets and media paths** stay local. Do not log API keys. Alert media files are user-configured paths (`ALERTS_MEDIA_PATH` in the console app).

---

## Agent working agreements

- Match existing Go style in the console repo (`slog`, small `internal/` packages, interfaces at the package boundary).
- Keep this file updated when routes, payloads, stack, or directory layout change.
- **Keep [README.md](README.md) up to date in the same change** whenever user-visible facts move: features, project status, requirements, install/run, config path, OBS URLs, webhook examples, TTS OS notes, license, contributing commands, the package `go doc` table, CI, or release artifacts. README follows [Make a README](https://www.makeareadme.com/): name, description, install, usage, contributing, license, and honest **project status**. Do not dump this file into the README; deep contracts stay here.
- **Use cases only after porting.** When a module is first added under `internal/` (or `app.go` for UC-11), write `docs/use-cases/<module>/uc-NN-….md` from `docs/use-cases/_TEMPLATE.md` and set the row to Ported in `docs/use-cases/INDEX.md`. Do not invent UC files for code that is not in this repo.
- **Godoc on every Go package.** Package comment plus comments on all exports. After porting, add a `go doc ./…` row for that package in README (Package documentation). Verify with `go doc -all` before considering the port done.
- **Tests on every ported package.** Unit tests always; integration tests (`//go:build integration`) when the package hits HTTP, disk, or OS APIs. Platform-specific integration files also tag GOOS (`integration && darwin`). Keep `internal/` CGO-free. CI must stay green.
- **New core decisions get an ADR** in `docs/architecture/` (next number, update `docs/architecture/INDEX.md`). Do not leave accepted decisions only in chat.
- Prefer the smallest change that ports one behavior correctly over a large rewrite.
- Treat HTTP + WebSocket overlay/chat as core, not a follow-up. Do not “simplify” by moving chat into the Wails window.
- Do not revive console routes (`/`, `/ws`, `/chat`, `/ws_chat`, `/media/`, `/webhook/`) as aliases without an explicit migration request.
- If an API key is set, never fall back to `youtube/nokey` on error.
