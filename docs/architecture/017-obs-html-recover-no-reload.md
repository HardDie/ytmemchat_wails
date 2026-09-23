# 17. OBS pages recover in-place; scripts grouped by role

* **Status:** Accepted
* **Date:** 2026-09-23
* **Authors:** @oleg

---

## Context

1. `window.onerror` and `unhandledrejection` reloaded the Browser Source after 5s.
2. That wiped chat lines and overlay media.
3. Overlay lost the Enable Audio click.
4. Chat `JSON.parse` had no try/catch, so bad JSON hit the global handler.
5. All JS lived in one or two undifferentiated `<script>` blocks.

## Considered options

1. **Keep reload** — recovers a dead page; destroys on-screen state.
2. **Log only** — page stays; socket can stay wedged.
3. **Recover through `connectWebSocket`** — same loop as a drop.

## Decision

Use option 3.

1. Do not call `location.reload`.
2. `recoverFromUnhandledError` logs, then `connectWebSocket`.
3. Overlay also `clearOverlayPlayback`.
4. Chat does not flush rows.
5. 2s cooldown (`recovering`) so a throwing handler cannot spin.
6. `unhandledrejection` uses `preventDefault`.
7. `onerror` returns true (handled).
8. Chat `JSON.parse` is try/catch in `handleChatMessage`.
9. Bad JSON is logged and skipped. It does not recover.
10. Overlay `JSON.parse` uses the same try/catch in `handleOverlayMessage`.
11. HTML is split into `<script>` blocks by role.
12. Banners: `/* <====== … ======> */`.
13. Shared: overlay `appClosed`, `audioContext`; chat `appClosed`.
14. Those stay `var` (classic scripts do not share `let`/`const`).
15. Other bindings are `let` / `const` in the block that uses them.

## Consequences

### Positive

* An unexpected throw does not reset OBS Interact or chat history.
* Recovery uses the same socket loop as [014](014-obs-html-one-websocket.md) / [015](015-obs-html-reconnect-timer.md).
* Script roles are easy to find.

### Negative and risks

* A bug inside `connectWebSocket` is throttled, not reloaded away.
* A dead `AudioContext` still needs Interact (rare).

### Neutral

* Review notes: `docs/obs-html-review.md` items 4 and 14.
