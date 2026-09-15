# UC-11: Start and stop YouTube chat ingest

**Module:** `app.go` / `pipeline.go`  
**Status:** Ported  
**Actors:** Operator in the Wails window  
**Goal:** Start polling live chat, push lines to `/obs/chat/ws`, and fan each line to overlay alerts or TTS; Stop cancels the iterator  
**Preconditions:** Settings saved with a stream ID; OBS HTTP is listening  

## Main scenario (happy path)

1. The operator clicks Start (the window saves the form first).
2. `CanStart` succeeds. Empty API key selects `youtube/nokey`; a non-empty key selects Data API v3.
3. If alerts are enabled and `commandsFilePath` is set, the matcher is loaded. TTS engine is created when TTS is enabled.
4. `GetMessageIterator` runs off the UI thread. History is skipped inside the client.
5. Each `ChatMessage` is published as an OBS [ChatEvent]. If the text matches an alert command, the overlay gets `alert` and TTS is skipped. Otherwise, if TTS is on, the overlay gets `tts`.
6. Stop cancels the iterator context. OBS HTTP stays up.

## Alternative scenarios and errors

* **2a. Empty stream ID:** `ErrStreamIDRequired`; iterator is not started.
* **2b. OBS HTTP not listening:** Start returns an error; no iterator.
* **3a. Alerts enabled, commands file missing or invalid:** Start returns an error (`alerts commands file: …`); iterator is not started.
* **3b. Alerts enabled, `commandsFilePath` empty:** matcher is skipped (chat still starts; TTS may run).
* **4a. Invalid API key:** status “YouTube API key is invalid”; **no** nokey fallback.
* **4b. Not live / quota / network:** distinct or wrapped error; still no nokey fallback if a key was set.
* **5a. TTS synthesis fails** (no speakable text, missing engine): that line is logged; ingest continues.
* **Start while running:** previous iterator is cancelled first.

## Postconditions

* At most one live iterator. Force-quit skips `app_closed` as before. `POST /api/webhook` uses the same overlay fan-out while HTTP is up, including when the iterator is stopped.
