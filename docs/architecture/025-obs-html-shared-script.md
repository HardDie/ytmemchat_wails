# 25. Chat and overlay share one socket script

* **Status:** Accepted
* **Date:** 2026-09-24
* **Authors:** @oleg

---

## Context

1. Chat and overlay each owned the socket loop.
2. Recover, retry, and the connecting watchdog were copies.
3. Page behavior still differs.
4. Overlay tears down playback. Chat keeps rows.
5. Ready logs differ.

## Considered options

1. **Keep two copies** — drift on the next socket change.
2. **One `script.js`** — pages pass hooks. Go serves the file.

## Decision

Use option 2.

1. File is `internal/obs/script.js` (`//go:embed`).
2. Route is `GET /obs/script.js`.
3. Content type is `text/javascript`.
4. Shared: recover, reconnect, connecting timeout, status banner.
5. `appClosed` stays `var` in that file.
6. Pages set `socketReadyLog` and `handleSocketMessage` before the file runs.
7. Those hooks are `var` so classic scripts share them.
8. Overlay sets `onConnectionLost` to `clearOverlayPlayback`.
9. Chat leaves that hook unset.
10. Recover and `onclose` call the hook when it is a function.
11. `connectWebSocket` still runs when the shared file loads.
12. Page scripts stay inline by role ([017](017-obs-html-recover-no-reload.md)).

## Consequences

### Positive

* Socket changes land once.
* Chat still does not flush rows on drop.

### Negative and risks

* A page that omits the hooks connects with no message handler.

### Neutral

* Same reconnect rules as [015](015-obs-html-reconnect-timer.md).
