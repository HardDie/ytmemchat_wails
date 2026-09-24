# 18. Overlay video stays at least 5s (memealerts)

* **Status:** Accepted
* **Date:** 2026-09-24
* **Authors:** @oleg

---

## Context

1. memealerts OBS `StickerPlayer` does not remove on `ended`.
2. Display time is `max(duration, minStickerPlayingTimeSec)`.
3. Default min is 5 seconds.
4. Fade starts 500ms before that.
5. Short `.webm` plays sound once, then stays until the min time.
6. Extra silent motion is the file, not a second unmuted `play()`.
7. This overlay used to remove on `ended`.
8. That hid short clips as soon as `ended` fired.

## Considered options

1. **Remove on first `ended`** — short stickers vanish too fast vs memealerts.
2. **`max(duration, 5s)` then fade** — match memealerts min time.
3. **Mute-loop until timeout** — not what memealerts JS does (`loop` off).

## Decision

Use option 2.

1. `MIN_STICKER_PLAYING_SEC` is 5.
2. `MEDIA_FADE_MS` is 500 (same as container fade).
3. Do not remove video on `ended`.
4. `loop` stays false.
5. After layout, `play()` (autoplay backup).
6. `scheduleVideoTeardown`: wait `max(duration, 5s)` minus fade, then `removeMediaElement`.
7. If duration is missing or not finite, use 5s.
8. Images still use the 6s timer.
9. Audio-only still uses `ended`.
10. `clearOverlayPlayback` can still cut a clip short ([016](016-obs-html-app-closed-teardown.md)).
11. `error` / `abort` call `removeMediaElement`.
12. If layout never ran, remove after 5s (video) or 6s (image).
13. That backstop no-ops once `laidOut` is true.
14. Audio-only also drops on `error` / `abort`.
15. Audio that never reaches `loadedmetadata` is dropped after 5s.
16. `.gif` is `<img>`, not `<video>`.
17. After audio metadata, drop after `max(duration, 5s)`.
18. That audio timer does not subtract fade (no picture).
19. Audio `ended` / `error` / `abort` still drop immediately. Drop is idempotent.

## Consequences

### Positive

* Short `.webm` alerts stay on screen like memealerts.
* Long clips still play to the end, then fade.

### Negative and risks

* Min 5s is hardcoded (memealerts can change `minStickerPlayingTimeSec` in config).

### Neutral

* Review notes: `docs/obs-html-review.md` items 6, 7, 8, and 13.
