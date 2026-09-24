# 15. OBS pages keep one reconnect timer

* **Status:** Accepted
* **Date:** 2026-09-23
* **Authors:** @oleg

---

## Context

1. Chat and overlay retry on `onclose`.
2. They used `setTimeout(connectWebSocket, 500)` with no id.
3. A second `connectWebSocket` could not cancel that retry.
4. Two `onclose` events, or a click overlapping a fail, could queue two timers.
5. That can mean two sockets with handlers.
6. Duplicate chat lines or stacked alerts.

## Considered options

1. **Leave fire-and-forget `setTimeout`** — usual fail loop is one timer; overlap is rare.
2. **One `reconnectTimer`** — schedule if idle; clear on a live connect.

## Decision

Use option 2.

1. `let reconnectTimer = null`.
2. `scheduleReconnect` no-ops if a timer is already set.
3. `onclose` sets `websocket = null` then calls `scheduleReconnect`.
4. `connectWebSocket` `clearTimeout`s first.
5. Delay stays `RETRY_INTERVAL` (500ms).
6. `onerror` still only `close()`s.
7. Retry stays in `onclose`.
8. Same logic in `overlay.html` and `chat.html`.
9. First connect is still a direct `connectWebSocket` ([014](014-obs-html-one-websocket.md)).
10. `CONNECTING_TIMEOUT_MS` is 2500.
11. A watchdog `close()`s a socket still in `CONNECTING`.
12. Retry still goes through `onclose` → `scheduleReconnect`.
13. `onopen` and `onclose` clear the watchdog.

## Consequences

### Positive

* A pending retry cannot open a second socket beside a live one.
* A second `onclose` does not stack timers.

### Negative and risks

* A slow handshake that takes more than 2.5s is aborted and retried.

### Neutral

* Review notes: `docs/obs-html-review.md` items 2 and 21.
