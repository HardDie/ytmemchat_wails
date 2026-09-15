# 2. Nested `/obs` and `/api` HTTP paths

* **Status:** Accepted
* **Date:** 2026-09-15
* **Authors:** @oleg

---

## Context

The console app exposes overlay at `/`, overlay socket at `/ws`, chat HTML at `/chat`, chat socket at `/ws_chat`, media at `/media/`, and operator endpoints at `/webhook/` and `/interrupt/`. That mixes OBS pages with APIs and splits each page from its socket.

This desktop app is new; existing OBS sources will be re-added from the settings window. JSON payloads should stay compatible; URLs do not have to.

## Considered options

1. **Keep console routes** for drop-in OBS compatibility.
2. **Nest by surface:** `/obs/chat`, `/obs/chat/ws`, `/obs/overlay`, `/obs/overlay/ws`, `/obs/media/`, `/api/webhook`, `/api/interrupt`.
3. **Query-param single page** (`/?view=chat`) — fewer routes, worse OBS bookmarks.

## Decision

Use option 2. One Browser Source URL per page; that page’s WebSocket is a child `…/ws` derived from `location`. `/` is a human index of the two OBS URLs, not an OBS source. Do not register console aliases unless a migration shim is requested later.

Optional `?transparent=1` on `/obs/chat` maps to the existing transparent body class.

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
