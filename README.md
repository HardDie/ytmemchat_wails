# ytmemchat

ytmemchat is a desktop companion for [YouTube](https://www.youtube.com/) Live. It reads live chat and drives an [OBS](https://obsproject.com/) overlay: on-stream messages, meme alerts, and text-to-speech.

This repository is a [Wails](https://wails.io) + [Svelte](https://svelte.dev) desktop app. You configure and start the pipeline in a native window. OBS Browser Sources load local HTTP pages that update over WebSocket.

**Project status:** early development. The Wails window is scaffolded (`make dev`); settings, OBS HTTP on launch, and YouTube Start are not wired yet. Until those land, use the [console app](https://github.com/HardDie/ytmemchat) on stream.

## Features

- Live YouTube chat with history skipped on connect
- Optional [YouTube Data API v3](https://developers.google.com/youtube/v3) key; empty key uses the no-key live chat client
- Desktop settings window (API key, stream ID, port, start/stop, status)
- OBS chat page and alert/TTS overlay over HTTP + WebSocket
- Alert commands from `commands.yaml` (for example `@jump`)
- TTS on macOS (`say`), Windows (PowerShell), and Linux (`espeak`)
- Optional HTTP API to inject a test message or interrupt TTS

The Wails window is a control panel. It is not the on-stream chat renderer.

## Requirements

- [Go](https://go.dev/dl/) 1.25 or later
- [Node.js](https://nodejs.org/) (npm)
- [Wails CLI](https://wails.io/docs/gettingstarted/installation): `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- macOS: Xcode Command Line Tools. Windows: WebView2. Linux: Wails system packages plus `espeak` for TTS
- A YouTube live video ID (the `v=` value in the watch URL)
- Optional: a Google Cloud API key with **YouTube Data API v3** enabled. Leave the key empty to use the no-key client. A **wrong** key does not fall back; the window reports that the key is invalid

## Installation

From source:

```bash
git clone <this-repo>
cd ytmemchat_wails
make build
```

Tagged versions also publish archives on GitHub Releases (Linux amd64/arm64, Windows amd64, macOS universal).

During development:

```bash
make dev
```

Settings are stored in a local JSON file (mode `0600`), not in the repo:

| OS | Path |
|---|---|
| macOS | `~/Library/Application Support/ytmemchat/config.json` |
| Windows | `%AppData%\ytmemchat\config.json` |
| Linux | `~/.config/ytmemchat/config.json` |

## Usage

1. Open the app. OBS URLs are already served (default port `8080`). Set the **stream/video ID**. Optionally set a YouTube API key.
2. In OBS, add two **Browser Sources**:

| Source | URL (port `8080`) |
|---|---|
| Chat | `http://127.0.0.1:8080/obs/chat` |
| Alerts + TTS | `http://127.0.0.1:8080/obs/overlay` |

3. Size each source to your canvas (for example `1920x1080`).
4. On the overlay source, click **Interact** once and allow audio so TTS and alert sounds can autoplay.
5. Click **Start** (or Run) in the app to pull live chat. **Stop** ends YouTube polling; OBS sources stay connected.

Chat with a transparent background: `http://127.0.0.1:8080/obs/chat?transparent=1`.

The config window can copy these URLs for the current port. Opening `http://127.0.0.1:8080/` in a browser lists them; do not use `/` as an OBS source.

### Alert commands

Point the app at a media folder and a YAML file:

```yaml
commands:
  - name: "jump"
    file: "mario_jump.mp3"
    volume: 0.5
  - name: "dance"
    file: "dancing_cat.gif"
    scale: 1.2
```

Chat messages that contain the command token (default `@`) plus a command name play that file on the overlay. Messages that do not match a command can be spoken by TTS when TTS is enabled.

### Test API

With the server running and webhooks enabled:

```bash
curl -X POST http://127.0.0.1:8080/api/webhook \
     -H 'Content-Type: application/json' \
     -d '{"message": "@jump"}'

curl -X POST http://127.0.0.1:8080/api/interrupt
```

## How it works

1. Go polls YouTube live chat (Data API v3 if a key is set, otherwise the no-key client).
2. Each message is sent to the chat WebSocket (`/obs/chat/ws`).
3. If the text matches an alert command, the overlay WebSocket (`/obs/overlay/ws`) gets an alert. Otherwise, if TTS is on, the overlay gets speech audio.
4. OBS Browser Sources render those pages.

## Roadmap

- Scaffold Wails v2 + Svelte window (done); persist settings and start OBS HTTP on launch
- Wire YouTube Start/Stop into the already-open OBS sockets
- Copy OBS URLs from the settings window
- GitHub Actions: tests on push; tagged releases with Linux (amd64, arm64), Windows, and macOS binaries
- Later: OS keychain for the API key; code-signed installers

See [CURSOR.md](CURSOR.md) for layout, routes, and implementation rules. Architecture decisions are in [docs/architecture](docs/architecture/INDEX.md). Use cases are added under [docs/use-cases](docs/use-cases/INDEX.md) after each module is ported.

## Documentation

| Doc | Audience |
|---|---|
| [README.md](README.md) | Users: install, OBS, status |
| [CURSOR.md](CURSOR.md) | Contributors/agents: contracts |
| [docs/architecture](docs/architecture/INDEX.md) | ADRs |
| [docs/use-cases](docs/use-cases/INDEX.md) | Scenarios for ported modules |

## Contributing

The Wails window is scaffolded. Settings save, OBS listen-on-launch, and Start/Stop are not wired yet. When you change **user-facing** behavior (features, install steps, OBS URLs, settings location, status, CI, releases), update this README in the same change. When you **port a module**, add a use-case file under `docs/use-cases/<module>/`, set its row to Ported in [docs/use-cases/INDEX.md](docs/use-cases/INDEX.md), give the package complete [Go documentation](https://go.dev/doc/comment), add a `go doc` command for it in the table below, and add unit tests (plus integration tests when the package hits HTTP, disk, or the OS). New architecture choices get an ADR in [docs/architecture](docs/architecture/INDEX.md).

Developer-oriented contracts live in [CURSOR.md](CURSOR.md). Match Go style in [HardDie/ytmemchat](https://github.com/HardDie/ytmemchat). Day-to-day commands:

```bash
make help              # all targets
make dev               # wails dev (needs scaffold)
make build             # local production binary
make test              # unit tests (race), same as CI
make test-integration  # integration tests
make test-all          # unit then integration
make doc PKG=./internal/tts
make doc-all PKG=./internal/tts
```

Unit tests are `*_test.go` in the same package. Integration tests use `//go:build integration`. If a test is tied to one OS, add that GOOS to the tag (`integration && darwin`, `integration && linux`, `integration && windows`) so it is not compiled elsewhere. Skip only when the platform matches but an optional binary is missing. Keep `internal/` free of Wails/CGO so CI does not need WebKit. Integration tests must not require YouTube credentials.

GitHub Actions ([`.github/workflows/test.yml`](.github/workflows/test.yml)) runs those two test commands on every **push** and **pull request** (skipped until `go.mod` exists).

### Releases

Push a semver tag with a `v` prefix:

```bash
git tag v0.1.0
git push origin v0.1.0
```

[`.github/workflows/release.yml`](.github/workflows/release.yml) publishes a GitHub Release with:

| File | Platform |
|---|---|
| `ytmemchat-linux-amd64.tar.gz` | Linux x86_64 |
| `ytmemchat-linux-arm64.tar.gz` | Linux ARM64 |
| `ytmemchat-windows-amd64.zip` | Windows x86_64 |
| `ytmemchat-darwin-universal.zip` | macOS Intel + Apple Silicon |
| `SHA256SUMS.txt` | checksums |

Requires `wails.json`. Binaries are not code-signed yet.

### Package documentation

Each Go package must have a package comment and comments on all exported names. After a package is ported, check it from the **repository root**:

```bash
make doc PKG=./internal/<package>
make doc-all PKG=./internal/<package>
```

Optional HTML browse of the whole module:

```bash
go run golang.org/x/pkgsite/cmd/pkgsite@latest -http localhost:8081
```

Then open `http://localhost:8081` and select this module.

| Package | Status | Check docs |
|---|---|---|
| `internal/config` | Ported | `go doc -all ./internal/config` |
| `internal/youtube` | Ported | `go doc -all ./internal/youtube` |
| `internal/youtube/nokey` | Ported | `go doc -all ./internal/youtube/nokey` |
| `internal/obs` | Ported | `go doc -all ./internal/obs` |
| `internal/alerts` | Ported | `go doc -all ./internal/alerts` |
| `internal/tts` | Ported | `go doc -all ./internal/tts` |
| `package main` (bindings façade) | Scaffolded | `go doc -all .` |

Set **Status** to Ported in this table when `go doc -all` prints a real package comment and every export is described.

Pull requests are welcome once the project is public. For large changes, open an issue first.

## Support

Use the GitHub issue tracker on this repository (when a remote exists). The original console app is [HardDie/ytmemchat](https://github.com/HardDie/ytmemchat).

## Authors

Desktop port of [HardDie/ytmemchat](https://github.com/HardDie/ytmemchat).

## License

License not chosen yet. Do not assume MIT or any other terms until a `LICENSE` file is added.
