# 2. Nested `/obs` and `/api` HTTP paths

* **Status:** Accepted
* **Date:** 2026-09-15
* **Authors:** @oleg

---

## Context

The console app mixes OBS pages with APIs and splits each page from its socket.

1. Overlay HTML at `/`.
2. Overlay socket at `/ws`.
3. Chat HTML at `/chat`.
4. Chat socket at `/ws_chat`.
5. Media at `/media/`.
6. Operator endpoints at `/webhook/` and `/interrupt/`.

This desktop app is new.

1. Existing OBS sources will be re-added from the settings window.
2. JSON payloads should stay compatible.
3. URLs do not have to.

## Considered options

1. **Keep console routes** for drop-in OBS compatibility.
2. **Nest by surface:** `/obs/chat`, `/obs/chat/ws`, `/obs/overlay`, `/obs/overlay/ws`, `/obs/media/`, `/api/webhook`, `/api/interrupt`.
3. **Query-param single page** (`/?view=chat`) — fewer routes, worse OBS bookmarks.

## Decision

Use option 2.

1. One Browser Source URL per page.
2. That page’s WebSocket is a child `…/ws` derived from `location`.
3. `/` is a human index of the two OBS URLs, not an OBS source.
4. Do not register console aliases unless a migration shim is requested later.
5. Optional `?transparent=1` on `/obs/chat` maps to the existing transparent body class.
6. Optional `?fontSize=` on `/obs/chat` (CSS size with a unit).
7. Optional `?textColor=` on `/obs/chat` (hex without `#`).
8. Overlay query is `debugAudio` only ([024](024-obs-html-audio-sticker.md)).
9. Overlay `?debugAudio=1` shows the audio debug square ([024](024-obs-html-audio-sticker.md)).
10. Chat `?cap=` ([022](022-obs-html-chat-cap.md)).
11. Chat `textColor` may include `#`.

## Consequences

### Positive

* OBS sources are obvious (`/obs/chat` vs `/obs/overlay`).
* Operator APIs are not mixed with Browser Source URLs.
* HTML does not hardcode `/ws_chat`.

### Negative and risks

* Console OBS URLs will not work until the user updates Browser Sources.
* Overlay media must be referenced as `/obs/media/<file>`, not `/media/`.

### Neutral

* Chat and overlay WebSocket JSON field names stay as in the console contracts.
