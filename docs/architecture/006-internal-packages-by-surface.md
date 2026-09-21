# 6. Internal packages grouped by surface

* **Status:** Accepted
* **Date:** 2026-09-15
* **Authors:** @oleg

---

## Context

The console tree has more packages than behaviors.

1. `clients/youtube`
2. `clients/youtubev1`
3. `server`
4. `chat`
5. `webhook`
6. `pkg/watermill` to fan in messages
7. Two WebSocket hubs are almost the same code.
8. `webhook` is two POST handlers.
9. `alerts` and `tts` contain real domain logic.

## Considered options

1. **Copy the console `internal/` tree** — fastest port, more packages than behaviors.
2. **Collapse by surface**
   1. `youtube` + `youtube/nokey`
   2. `obs` (mux, both hubs, HTML, `/api`)
   3. Keep `alerts` and `tts`
   4. Wire in `app.go` without watermill until needed
3. **One `internal/app` package** — hard to test and easy to create import cycles.

## Decision

Use option 2.

1. Import direction: `package main` → `config`, `youtube`, `obs`, `alerts`, `tts`, `hotkey`, `quota`.
2. `obs` does not import `alerts` or `youtube`.
3. Overlay payload types live in `obs`.
4. `youtube` owns `ChatMessage` and `Client`.
5. `config` uses `secret` for the API key ([012](012-os-keychain-api-key.md)).
6. `internal/update` is used from `bindings/update`, not from `app.go`.

## Consequences

### Positive

* One HTTP server to reason about.
* No package for a pair of webhook handlers.
* Alerts/TTS stay testable without net/http.

### Negative and risks

* Port is a reshape, not a file copy. Easier to miss a route while moving HTML.

### Neutral

* Watermill can return if Start/Stop fan-in becomes messy.
