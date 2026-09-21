# 5. Official Wails `svelte-ts` layout

* **Status:** Accepted
* **Date:** 2026-09-15
* **Authors:** @oleg

---

## Context

The console app uses `cmd/main.go`. Wails CLI expects `wails.json` beside `main.go` and embeds `frontend/dist`. SvelteKit is a common add-on but needs extra adapter/embed/`wailsjsdir` wiring. This UI is a single settings window.

## Considered options

1. **`wails init -t svelte-ts`** — Vite Svelte at `frontend/`, `app.go` + `main.go` at repo root.
2. **SvelteKit** as the frontend — routing, SSR off, static adapter, embed `frontend/build`.
3. **Keep `cmd/main.go`** like the console app — fights Wails defaults.

## Decision

Use option 1. Bind one struct per window pane under `bindings/` (`home`, `configuration`, `commands`, `test`); they wrap `App` in `package main`. Domain stays in `internal/`. Do not use SvelteKit unless file-based routing is required. After scaffold, use Svelte 5 `mount()` in `frontend/src/main.ts` if the template still uses `new App({ target })`.

## Consequences

### Positive

* Matches Wails docs and generated bindings (`frontend/wailsjs/go/<package>/<Struct>`).
* Frontend remains a normal Vite project.

### Negative and risks

* Current official templates may pin Svelte 5 while bootstrapping with the Svelte 4 component API (blank window until `mount()`).

### Neutral

* Overlay HTML is not part of the Vite app.
