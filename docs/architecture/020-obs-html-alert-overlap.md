# 20. Overlay alerts may overlap; interrupt is TTS only

* **Status:** Accepted
* **Date:** 2026-09-24
* **Authors:** @oleg

---

## Context

1. Chat commands that match YAML become overlay `type: alert`.
2. Files are `/obs/media/` via `displayMediaAlert`.
3. TTS is a separate queue ([019](019-obs-html-tts-queue.md)).
4. Operators add those clips in Commands. The list is curated.
5. Clips are short (min on-screen time in [018](018-obs-html-min-sticker-time.md)).

## Considered options

1. **Queue alerts like TTS** — one clip at a time.
2. **Show every alert as it arrives** — several can overlap.
3. **Same interrupt as TTS for command clips** — extra stop for stickers.

## Decision

Use option 2. Do not use option 3.

1. `type: alert` calls `displayMediaAlert` immediately.
2. More than one alert may be on screen at once.
3. Alert audio may overlap.
4. Alert files are not `alertQueue`.
5. `tts_interrupt` does not stop alert media.
6. Home / hotkey / `/api/interrupt` stay TTS-only ([010](010-global-interrupt-hotkey.md)).
7. Commands YAML is the moderation gate, not the overlay.
8. `clearOverlayPlayback` still drops all alerts on teardown ([016](016-obs-html-app-closed-teardown.md)).
9. Command audio uses the HTML sticker ([024](024-obs-html-audio-sticker.md)).

## Consequences

### Positive

* Fast command spam still looks like stickers, not a FIFO.
* Interrupt stays for speech, not for curated clips.

### Negative and risks

* Many alerts at once can stack and mix audio.

### Neutral

* Review notes: `docs/obs-html-review.md` item 9.
* Use case: `docs/use-cases/alerts/uc-07-alert-command.md`.
