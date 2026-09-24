# 14. OBS pages open one WebSocket

* **Status:** Accepted
* **Date:** 2026-09-22
* **Authors:** @oleg

---

## Context

1. Console `overlay.html` / `chat.html` opened a socket at parse time.
2. That first socket had no handlers.
3. After 300ms they closed it and opened a second socket with handlers.
4. The comment said Chrome needs time to drop the previous socket.
5. Overlay mixed that with “settle the browser” after the audio button.
6. The hub still counted the dummy client.
7. Messages in that window were ignored.

## Considered options

1. **Keep dummy + 300ms** — extra handshake; first messages dropped.
2. **No dummy, keep 300ms on first connect** — delay with no previous socket to clear.
3. **Open only inside `connectWebSocket`** — `websocket` starts `null`.

## Decision

Use option 3.

1. `let websocket = null`.
2. The only `new WebSocket` is inside `connectWebSocket`.
3. Chat calls it on load.
4. Overlay also calls it on load ([024](024-obs-html-audio-sticker.md)).
5. No 300ms wait on first connect.
6. That wait existed to finish `CLOSING` on the dummy socket.
7. It was not an AudioContext or OBS load requirement in this repo.
8. Reconnect still waits `RETRY_INTERVAL` (500ms) after `onclose`.
9. `onclose` sets `websocket = null` before that timer.
10. Do not `close()` then `new WebSocket` on the same tick for first connect.
11. Do not put the 300ms delay back unless OBS CEF fails a same-tick first connect.
12. Overlay also calls `connectWebSocket` on load ([024](024-obs-html-audio-sticker.md)).

## Consequences

### Positive

* One client per page on the hub.
* First chat/overlay events are not sent to a socket with no handlers.

### Negative and risks

* If OBS CEF needs a load delay, first connect can fail.
* `onclose` already retries after 500ms in that case.

### Neutral

* Overlay load connect and HTML audio: [024](024-obs-html-audio-sticker.md).
* Review notes: `docs/obs-html-review.md` item 1.
