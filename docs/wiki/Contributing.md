# Contributing

Pull requests are welcome. For large changes, open an issue first.

When **user-facing** behavior changes, update [README.md](https://github.com/HardDie/ytmemchat_wails/blob/main/README.md) (short facts only) and the matching wiki page. When the **window UI** changes, run `make screenshots`. Porting a module, tests, godoc, and ADRs are described in [CURSOR.md](https://github.com/HardDie/ytmemchat_wails/blob/main/CURSOR.md). Match Go style in [HardDie/ytmemchat](https://github.com/HardDie/ytmemchat).

How to clone and build is in the repository [README](https://github.com/HardDie/ytmemchat_wails/blob/main/README.md#from-source).

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

GitHub Actions (`.github/workflows/test.yml`) runs `make test` then integration tests on every push and pull request. Push a `vMAJOR.MINOR.PATCH` tag to publish versioned release archives such as `ytmemchat-v0.1.0-linux-amd64.tar.gz` (`.github/workflows/release.yml`).

## Package documentation

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
| `internal/secret` | Ported | `go doc -all ./internal/secret` |
| `package main` (bindings façade) | Chat Start/Stop | `go doc -all .` |
