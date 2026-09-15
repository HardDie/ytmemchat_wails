# UC-08: Speak a non-alert chat message

**Module:** `internal/tts`  
**Status:** Ported  
**Actors:** Pipeline (after a chat line did not match an alert command)  
**Goal:** Turn the message text into WAV audio and hand it to the overlay sink  
**Preconditions:** TTS is enabled; a voice name is configured; `Config.Out` is a non-nil channel with a receiver or buffer  

## Main scenario (happy path)

1. The pipeline calls `TTS.SynthesizeAudio` with the chat text.
2. The package strips emoji and extra whitespace.
3. The OS engine (`say` / `espeak` / PowerShell) writes a WAV file.
4. A [Speech](../../internal/tts/tts.go) value (WAV bytes + volume) is sent on `Out`.
5. Overlay code maps that to WebSocket type `tts`.

## Alternative scenarios and errors

* **2a. Text is empty after stripping emoji:** SynthesizeAudio returns `tts: no speakable text` and sends nothing.
* **1a. `Out` is nil:** SynthesizeAudio returns `tts: nil Out channel`.
* **3a. Voice missing or engine not installed:** SynthesizeAudio returns a wrapped OS error; nothing is sent.
* **3b. Synthesized file is empty:** returns an error; nothing is sent.

## Postconditions

* On success, exactly one `Speech` value was sent and the temp WAV file was deleted.
* Local `Speak` is a separate path (speakers on the host), not used by the OBS overlay.
