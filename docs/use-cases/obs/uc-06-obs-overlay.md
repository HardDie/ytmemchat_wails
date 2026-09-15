# UC-06: Play alert / TTS on overlay

**Module:** `internal/obs`  
**Status:** Ported  
**Actors:** OBS overlay Browser Source, pipeline via `PublishOverlay`  
**Goal:** Overlay clients receive alert or TTS events and can load media from `/obs/media/`  
**Preconditions:** Server is serving; overlay WebSocket is connected  

## Main scenario (happy path)

1. OBS GETs `/obs/overlay`. The page uses `/obs/media/<file>` and a child `…/ws` socket.
2. The pipeline publishes `AlertOverlay(filename, volume, scale)` or `TTSOverlay(wav, volume)`.
3. `/obs/overlay/ws` clients receive JSON with `type` `alert` or `tts` (plus `payload`, `filename`, `volume`, `scale`).
4. Alert files are served from `MediaPath` at `/obs/media/` when that path is configured.

## Alternative scenarios and errors

* **4a. Empty MediaPath:** `/obs/media/` is not registered (404).
* **TTS audio:** `payload` is WAVE bytes (JSON base64), same as the console overlay.

## Postconditions

* Overlay JSON field names match the console overlay contract. `obs` does not import `alerts` or `tts`. The Wails process keeps this server up until exit.
