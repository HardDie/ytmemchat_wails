# UC-11: Start and stop YouTube chat ingest

**Module:** `app.go`  
**Status:** Ported (chat overlay only; alerts/TTS not in this slice)  
**Actors:** Operator in the Wails window  
**Goal:** Start polling live chat and push lines to `/obs/chat/ws`; Stop cancels the iterator  
**Preconditions:** Settings saved with a stream ID; OBS HTTP is listening  

## Main scenario (happy path)

1. The operator clicks Start (the window saves the form first).
2. `CanStart` succeeds. Empty API key selects `youtube/nokey`; a non-empty key selects Data API v3.
3. `GetMessageIterator` runs off the UI thread. History is skipped inside the client.
4. Each `ChatMessage` is published as an OBS [ChatEvent]. Overlay alerts/TTS are not invoked.
5. Stop cancels the iterator context. OBS HTTP stays up.

## Alternative scenarios and errors

* **2a. Empty stream ID:** `ErrStreamIDRequired`; iterator is not started.
* **2b. OBS HTTP not listening:** Start returns an error; no iterator.
* **3a. Invalid API key:** status “YouTube API key is invalid”; **no** nokey fallback.
* **3b. Not live / quota / network:** distinct or wrapped error; still no nokey fallback if a key was set.
* **Start while running:** previous iterator is cancelled first.

## Postconditions

* At most one live iterator. Force-quit skips `app_closed` as before. Alerts and TTS remain unwired.
