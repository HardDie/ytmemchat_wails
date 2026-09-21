# UC-10: Interrupt TTS

**Module:** `internal/obs` (`/api/interrupt`)  
**Status:** Ported  
**Actors:** Operator (curl), overlay clients  
**Goal:** Stop the current overlay utterance  
**Preconditions:** Server started with `Webhooks: true`; overlay WebSocket may be connected  

## Main scenario (happy path)

1. Operator POSTs `/api/interrupt`.
2. Handler responds 204 and publishes an overlay event with `type` `tts_interrupt`.
3. Overlay JS stops the current TTS source and clears queued speech. The Wails **Interrupt speech** button publishes the same event without enabling the HTTP API.

## Alternative scenarios and errors

* **Webhooks disabled:** route is not registered (404).
* **No overlay clients:** event is still accepted; nothing is queued for later.

## Postconditions

* Interrupt does not go through the inject channel. It is overlay-only. The Wails **Interrupt speech** binding (`InterruptTTS`) and the OS-global interrupt shortcut publish the same event when HTTP is listening, even if webhooks are off.
