# UC-05: Show live chat in OBS

**Module:** `internal/obs`  
**Status:** Ported  
**Actors:** OBS Browser Source, pipeline via `PublishChat`  
**Goal:** OBS loads `/obs/chat` and new messages appear over WebSocket  
**Preconditions:** Server is serving; chat WebSocket is connected  

## Main scenario (happy path)

1. OBS (or a browser) GETs `/obs/chat`.
2. The page derives its socket URL as a child path (`…/ws`), not a hardcoded `/ws_chat`.
3. The pipeline calls `PublishChat` with a [ChatEvent] (`authorName`, `authorPicture`, `messageText`, `publishedAt`, badges).
4. Connected `/obs/chat/ws` clients receive that JSON.

## Alternative scenarios and errors

* **1a. Trailing slash `/obs/chat/`:** 404 (HTML routes have no trailing slash).
* **1b. `?transparent=1` (or `true`):** the page adds `body.transparent`.
* **GET `/`:** index lists the two OBS URLs; it is not an OBS source.

## Postconditions

* Chat JSON field names match the console chat overlay. YouTube types are not imported here.
