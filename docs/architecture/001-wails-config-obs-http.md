# 1. Wails as configuration window; OBS via HTTP and WebSocket

* **Status:** Accepted
* **Date:** 2026-09-15
* **Authors:** @oleg

---

## Context

1. ytmemchat must let a streamer configure a YouTube live session.
2. Chat, alerts, and TTS must show on stream.
3. The console app already serves OBS Browser Sources over HTTP.
4. It pushes events on WebSockets.
5. This port adds a desktop UI with Wails + Svelte.

Wails can also emit events into its own webview.

1. Using that as the OBS transport would mean a second protocol.
2. It would also mean a chat renderer inside the desktop window.

## Considered options

1. **Wails Events as the primary chat/overlay path** — Svelte renders chat; OBS is optional later.
2. **Wails window = settings and control; existing HTTP + WebSocket = OBS display** — port the console overlay protocol.
3. **OBS only, no desktop window** — keep the console app.

## Decision

Use option 2.

1. The Svelte webview is a control panel: API key, stream ID, port, start/stop, status.
2. Chat and overlay are served by the Go HTTP server.
3. OBS Browser Sources load those pages.
4. Wails Events may be used only for window status, never as the OBS transport.

## Consequences

### Positive

* OBS setup stays a Browser Source URL, like the console app.
* Overlay JSON stays compatible with the console payloads.
* The desktop UI stays small (forms and status).

### Negative and risks

1. Two processes in one binary: Wails UI and an HTTP server.
2. The HTTP server stays up for the life of the app (see [ADR 009](009-obs-http-process-lifetime.md)).
3. Start/Stop is the YouTube iterator, not the OBS listener.
4. Streamers must still add Browser Sources. The Wails window does not replace OBS.

### Neutral

* Overlay and chat HTML stay `//go:embed` in Go (`internal/obs`), not in `frontend/`.
