# README window GIF

`window.gif` is the cycling Home → Configuration → Commands → Test animation in [README.md](../../README.md). PNG frames are local intermediates (gitignored).

Regenerate after Svelte UI changes (layout, copy, navigation, new controls):

```bash
make screenshots
```

That starts Vite with `?screenshot=1`, which loads fake Wails bindings (`frontend/src/screenshotBridge.ts`) so capture does not need `wails dev`. Viewport is 760×680, same as `main.go`. Playwright and gifenc live under `scripts/screenshots/` so `frontend/npm install` (release builds) does not download Chromium.
