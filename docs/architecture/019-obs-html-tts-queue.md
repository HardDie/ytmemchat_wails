# 19. Overlay TTS plays one at a time from a queue

* **Status:** Accepted
* **Date:** 2026-09-24
* **Authors:** @oleg

---

## Context

1. Chat lines that are not alerts become overlay `type: tts`.
2. WAV bytes are in `payload` (Web Audio, not `/obs/media/`).
3. Several TTS events can arrive while one utterance is still playing.
4. Alert files use a separate `<audio>` / `<video>` path (`displayMediaAlert`).
5. That path has a 5s never-ready drop ([018](018-obs-html-min-sticker-time.md)).

## Considered options

1. **Play every TTS immediately** — overlap, unreadable.
2. **Queue TTS; play one, then the next** — `alertQueue` + `isPlaying`.
3. **Same 5s never-ready timer as alert `<audio>`** — would cut queued speech that is not an `<audio>` element.

## Decision

Use option 2.

1. `type: tts` calls `addToQueue`.
2. `processQueue` starts the next item only when `isPlaying` is false.
3. `displayTTS` builds a WAV blob URL and calls `startAudibleSticker`.
4. `<audio>` `ended` calls `onMediaFinished` → next in queue.
5. Blob/build failure also calls `onMediaFinished` (queue does not stick).
6. `tts_interrupt` `stop()`s only the current source. Queue stays.
7. Alert media is not this queue. It is `displayMediaAlert` (may overlap; review item 9).
8. The 5s never-ready timer applies only to alert `<audio>` files.
9. TTS is never dropped by that timer.
10. After interrupt, `onMediaFinished` / `processQueue` plays the next queued TTS.
11. `clearOverlayPlayback` still empties the queue and stops the source (teardown).
12. Alert overlap and no alert interrupt: [020](020-obs-html-alert-overlap.md).
13. `ttsGeneration` increments on interrupt and teardown.
14. Stale `ttsGeneration` skips `play()` and queue finish.
15. TTS now plays through `startAudibleSticker` ([024](024-obs-html-audio-sticker.md)).
16. WAV is a blob URL on `<audio>`, not `decodeAudioData`.

## Consequences

### Positive

* TTS utterances do not talk over each other.
* A stuck alert MP3 cannot steal TTS.

### Negative and risks

* A long TTS queue can lag behind chat.
* Interrupt skips one line; later queued TTS still play.

### Neutral

* Review notes: `docs/obs-html-review.md` items 9, 10, and 22.
* Use case: `docs/use-cases/tts/uc-08-speak-message.md`.
