# ytmemchat

<p align="center">
  <img src="build/appicon.png" alt="ytmemchat" width="160">
</p>

ytmemchat is a desktop companion for [YouTube](https://www.youtube.com/) Live. It reads live chat and drives an [OBS](https://obsproject.com/) overlay: on-stream messages, meme alerts, and text-to-speech.

This repository is a [Wails](https://wails.io) + [Svelte](https://svelte.dev) desktop app. You configure and start the pipeline in a native window. OBS Browser Sources load local HTTP pages that update over WebSocket. The Wails window is a control panel; it is not the on-stream chat renderer.

**Project status:** early development. Start/Stop YouTube chat from Home. Alerts and TTS play on the overlay.

## Operator wiki

Step-by-step setup lives in the [wiki](https://github.com/HardDie/ytmemchat_wails/wiki), not in this file:

| Page | What it covers |
|---|---|
| [Getting Started](https://github.com/HardDie/ytmemchat_wails/wiki/Getting-Started) | Stream ID only, see chat (no API key) |
| [Running on macOS](https://github.com/HardDie/ytmemchat_wails/wiki/Running-on-macOS) | Open the unsigned `.app` (Gatekeeper) |
| [Chat URL](https://github.com/HardDie/ytmemchat_wails/wiki/Chat-URL) | OBS chat query parameters |
| [Configuration](https://github.com/HardDie/ytmemchat_wails/wiki/Configuration) | Each settings field, OBS URLs, HTTP API |
| [Commands](https://github.com/HardDie/ytmemchat_wails/wiki/Commands) | Commands pane / `commands.yaml` |
| [YouTube API key](https://github.com/HardDie/ytmemchat_wails/wiki/YouTube-API-key) | Google token, how to create one, default quota |

## Features

- Live YouTube chat (history skipped on connect); optional [Data API v3](https://developers.google.com/youtube/v3) key
- Home / Configuration / Commands / Test window
- OBS chat and alert/TTS overlay over HTTP + WebSocket
- Alert commands (for example `@jump`) and TTS (macOS `say`, Windows PowerShell, Linux `espeak`)
- Optional HTTP API and a Test pane to inject a fake chat line

<p align="center">
  <img src="docs/screenshots/window.gif" alt="Home, Configuration, Commands, and Test panes" width="760">
</p>

## Installation

Download a tagged archive from [GitHub Releases](https://github.com/HardDie/ytmemchat_wails/releases) (Linux amd64/arm64, Windows amd64, macOS universal). Binaries are not code-signed yet. On macOS, [open the unsigned app](https://github.com/HardDie/ytmemchat_wails/wiki/Running-on-macOS). The window sidebar shows the git tag, or the short commit the binary was built from.

From source (Go 1.25+, Node/npm, [Wails CLI](https://wails.io/docs/gettingstarted/installation); macOS Xcode CLT, Windows WebView2, Linux Wails packages):

```bash
git clone https://github.com/HardDie/ytmemchat_wails.git
cd ytmemchat_wails
make build
```

`make dev` runs the app with frontend hot reload.

## Usage

Follow [Getting Started](https://github.com/HardDie/ytmemchat_wails/wiki/Getting-Started). In short: set a live **stream/video ID** on Configuration, copy the Browser Source URLs from Home, then **Start**.

| OBS source | URL (default port `8080`) |
|---|---|
| Chat | `http://127.0.0.1:8080/obs/chat` |
| Alerts + TTS | `http://127.0.0.1:8080/obs/overlay` |

Do not use `http://127.0.0.1:8080/` as an OBS source. Chat query parameters (`transparent`, `fontSize`, `textColor`) are on [Chat URL](https://github.com/HardDie/ytmemchat_wails/wiki/Chat-URL). Overlay needs **Interact** once so audio can autoplay.

Alerts and TTS are off until you enable them. See [Configuration](https://github.com/HardDie/ytmemchat_wails/wiki/Configuration), [Commands](https://github.com/HardDie/ytmemchat_wails/wiki/Commands), and [YouTube API key](https://github.com/HardDie/ytmemchat_wails/wiki/YouTube-API-key).

## Documentation

| Doc | Audience |
|---|---|
| This README | Product, install, OBS URLs, status |
| [Wiki](https://github.com/HardDie/ytmemchat_wails/wiki) | Operator how-tos (source: `docs/wiki/`) |
| [CURSOR.md](CURSOR.md) | Contributors/agents: contracts |
| [docs/architecture](docs/architecture/INDEX.md) | ADRs |
| [docs/use-cases](docs/use-cases/INDEX.md) | Scenarios for ported modules |

## Contributing

When **user-facing** behavior changes, update this README (short facts only) and the matching [wiki](https://github.com/HardDie/ytmemchat_wails/wiki) page. When the **window UI** changes, run `make screenshots`. Porting a module, tests, godoc, and ADRs are described in [CURSOR.md](CURSOR.md). Match Go style in [HardDie/ytmemchat](https://github.com/HardDie/ytmemchat).

```bash
make help              # all targets
make dev
make build
make screenshots       # refresh docs/screenshots/window.gif
make test              # unit tests (race), same as CI
make test-integration
make test-all
make doc PKG=./internal/tts
make doc-all PKG=./internal/tts
```

GitHub Actions (`.github/workflows/test.yml`) runs `make test` then integration tests on every push and pull request. Push a `vMAJOR.MINOR.PATCH` tag to publish release archives (`.github/workflows/release.yml`).

### Package documentation

Each Go package needs a package comment and comments on all exports. Check from the repository root with `make doc-all PKG=./internal/<package>`. Optional HTML: `go run golang.org/x/pkgsite/cmd/pkgsite@latest -http localhost:8081`.

| Package | Status | Check docs |
|---|---|---|
| `internal/config` | Ported | `go doc -all ./internal/config` |
| `internal/youtube` | Ported | `go doc -all ./internal/youtube` |
| `internal/youtube/quota` | Ported | `go doc -all ./internal/youtube/quota` |
| `internal/youtube/nokey` | Ported | `go doc -all ./internal/youtube/nokey` |
| `internal/obs` | Ported | `go doc -all ./internal/obs` |
| `internal/alerts` | Ported | `go doc -all ./internal/alerts` |
| `internal/tts` | Ported | `go doc -all ./internal/tts` |
| `internal/hotkey` | Ported | `go doc -all ./internal/hotkey` |
| `package main` (bindings façade) | Chat Start/Stop | `go doc -all .` |

Pull requests are welcome. For large changes, open an issue first.

## Support

Use the GitHub issue tracker. The original console app is [HardDie/ytmemchat](https://github.com/HardDie/ytmemchat).

## Authors

Desktop port of [HardDie/ytmemchat](https://github.com/HardDie/ytmemchat).

## License

This project is licensed under the GNU General Public License v3.0 — see the [LICENSE](LICENSE) file for details.
