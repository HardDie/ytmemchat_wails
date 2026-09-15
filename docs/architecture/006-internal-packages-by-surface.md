# 6. Internal packages grouped by surface

* **Status:** Accepted
* **Date:** 2026-09-15
* **Authors:** @oleg

---

## Context

The console tree has `clients/youtube`, `clients/youtubev1`, `server`, `chat`, `webhook`, plus `pkg/watermill` to fan in messages. Two WebSocket hubs are almost the same code. `webhook` is two POST handlers. `alerts` and `tts` contain real domain logic.

## Considered options

1. **Copy the console `internal/` tree** — fastest port, more packages than behaviors.
2. **Collapse by surface:** `youtube` + `youtube/nokey`, `obs` (mux, both hubs, HTML, `/api`), keep `alerts` and `tts`, wire in `app.go` without watermill until needed.
3. **One `internal/app` package** — hard to test and easy to create import cycles.

## Decision

Use option 2.

Import direction: `app.go` → `config`, `youtube`, `obs`, `alerts`, `tts`. `obs` does not import `alerts` or `youtube`. Overlay payload types live in `obs`. `youtube` owns `ChatMessage` and `Client`.

## Consequences

### Positive

* One HTTP server to reason about.
* No package for a pair of webhook handlers.
* Alerts/TTS stay testable without net/http.

### Negative and risks

* Port is a reshape, not a file copy; easier to miss a route while moving HTML.

### Neutral

* Watermill can return if Start/Stop fan-in becomes messy.
