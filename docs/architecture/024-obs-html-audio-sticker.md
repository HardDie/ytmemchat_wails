# 24. Overlay TTS and command audio use a fading HTML sticker

* **Status:** Accepted
* **Date:** 2026-09-24
* **Authors:** @oleg

---

## Context

1. Browsers block Web Audio until a user gesture.
2. Overlay used a click button plus `AudioContext`.
3. TTS started while the context was `suspended`.
4. After resume, speech skipped the beginning.
5. memealerts OBS uses `new Audio().play()`, not `AudioContext`.
6. OBS CEF often allows HTMLMediaElement autoplay.

## Considered options

1. **Keep the click button** — works; needs Interact every refresh.
2. **Web Audio at load + `resume()`** — still suspended without a gesture.
3. **HTML `<audio>` sticker; fade; then `play()`** — same path as video stickers.

## Decision

Use option 3.

1. Overlay connects the WebSocket on load. No audio button.
2. TTS WAV becomes a blob `audio/wav` object URL.
3. Command audio files use `/obs/media/` on the same sticker.
4. Sticker is `.alert-container` plus `.audio-sticker-rect`.
5. Fade in `MEDIA_FADE_MS`, then `audio.play()`.
6. Do not `start()` a BufferSource.
7. Debug rect is off by default (0×0, no fill).
8. `?debugAudio=1` (or `true`) shows the green square.
9. Interrupt still drops only the current TTS sticker.
10. Command audio still uses min-time teardown ([018](018-obs-html-min-sticker-time.md)).
11. Audio stickers do not use the video white frame.
12. OBS Browser Source plays TTS and command audio with no click (verified).
13. Audio min-time timer starts at `play()`, after the fade-in.
14. That timer is TTS and command audio only. Not video.

## Consequences

### Positive

* No click to unlock TTS in OBS CEF (same class as memealerts).
* Speech does not start in a suspended `AudioContext`.

### Negative and risks

* A desktop Chrome tab may still block `play()` until a gesture.
* The debug square is off unless `?debugAudio=1`.

### Neutral

* Queue rules stay [019](019-obs-html-tts-queue.md).
* Overlay URL: `docs/wiki/Chat-URL.md`.
