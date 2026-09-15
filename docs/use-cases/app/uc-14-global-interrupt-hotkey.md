# UC-14: Global interrupt shortcut

**Module:** `internal/hotkey`, package main (`hotkey_bind.go`)  
**Status:** Ported  
**Actors:** Operator at the desk (OBS often fullscreen)  
**Goal:** Stop overlay TTS with a keyboard combination even when the ytmemchat window is not focused  
**Preconditions:** App is running so OBS HTTP is up; interrupt shortcut enabled in settings (default on)  

## Main scenario (happy path)

1. App starts and registers the saved chord (default `Ctrl+Shift+I`) with the OS.
2. TTS is speaking on the overlay while OBS is focused.
3. Operator presses the chord.
4. Overlay receives `tts_interrupt` (same event as the Home button).

## Alternative scenarios and errors

* **Shortcut disabled:** nothing is registered.
* **Chord already taken:** register fails; Config and Home show the error; the Home button still works.
* **OBS HTTP down:** the key is ignored (debug log).
* **Linux Wayland:** registration may fail; use the Home button or `/api/interrupt`.

## Postconditions

* Changing the chord in Config and saving unregisters the old combination and registers the new one.
* Shutdown unregisters the hotkey.
