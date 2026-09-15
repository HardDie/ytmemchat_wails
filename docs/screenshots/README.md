# README window GIF

`window.gif` is the cycling Home → Configuration → Commands → Test animation. On GitHub Releases, `release.yml` generates and publishes `window.gif` as a release asset, which [README.md](../../README.md) embeds directly via `https://github.com/HardDie/ytmemchat_wails/releases/latest/download/window.gif`. PNG frames are local intermediates (gitignored).

Regenerate locally after Svelte UI changes (layout, copy, navigation, new controls):

```bash
make screenshots
```

That starts Vite with `?screenshot=1`, which loads fake Wails bindings (`frontend/src/screenshotBridge.ts`) so capture does not need `wails dev`. Viewport is 760×680, same as `main.go`. Playwright and gifenc live under `scripts/screenshots/` so `frontend/npm install` (release builds) does not download Chromium.
