# 1. Wails as configuration window; OBS via HTTP and WebSocket

* **Status:** Accepted
* **Date:** 2026-09-15
* **Authors:** @oleg

---

## Context

ytmemchat must let a streamer configure a YouTube live session and show chat, alerts, and TTS on stream. The console app already serves OBS Browser Sources over HTTP and pushes events on WebSockets. This port adds a desktop UI with Wails + Svelte.

Wails can also emit events into its own webview. Using that as the OBS transport would mean a second protocol and a chat renderer inside the desktop window.

## Considered options

1. **Wails Events as the primary chat/overlay path** — Svelte renders chat; OBS is optional later.
2. **Wails window = settings and control; existing HTTP + WebSocket = OBS display** — port the console overlay protocol.
3. **OBS only, no desktop window** — keep the console app.

## Decision

Use option 2.

The Svelte webview is a control panel: API key, stream ID, port, start/stop, status. Chat and overlay are served by the Go HTTP server. OBS Browser Sources load those pages. Wails Events may be used only for window status, never as the OBS transport.

## Consequences

### Positive

* OBS setup stays a Browser Source URL, like the console app.
* Overlay JSON stays compatible with the console payloads.
* The desktop UI stays small (forms and status).

### Negative and risks

* Two processes in one binary: Wails UI and an HTTP server that stays up for the life of the app (see [ADR 009](009-obs-http-process-lifetime.md)). Start/Stop is the YouTube iterator, not the OBS listener.
* Streamers must still add Browser Sources; the Wails window does not replace OBS.

### Neutral

* Overlay and chat HTML stay `//go:embed` in Go (`internal/obs`), not in `frontend/`.
