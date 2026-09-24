# 16. Overlay drops media on `app_closed` and socket close; chat keeps lines

* **Status:** Accepted
* **Date:** 2026-09-23
* **Authors:** @oleg

---

## Context

1. Graceful quit sends `type: app_closed` then closes the sockets.
2. Both pages only flipped the banner.
3. Overlay clips and TTS kept playing.
4. `/obs/media/` is gone after quit, so a video can stall.
5. OBS does not reload the HTML.
6. A later Preview can stack on the old frame.
7. Chat lines across a restart are useful.

## Considered options

1. **Banner only** — stuck overlay media after quit.
2. **Tear down overlay and chat** — chat goes blank on every quit.
3. **Tear down overlay playback only** — chat rows stay.
4. **Also clear overlay on `onclose`** — crash and socket drop.

## Decision

Use option 3 plus option 4.

1. Overlay `app_closed` calls `clearOverlayPlayback`.
2. Pause and drop `.alert-container` media.
3. Remove body `<audio>` nodes.
4. `interruptCurrentTTS` (queue + current source).
5. Then show the closed banner.
6. Chat `app_closed` still only shows the banner.
7. Do not send `chat_flush` on quit.
8. Overlay `onclose` also calls `clearOverlayPlayback`.
9. Force-quit skips `app_closed` (unchanged).
10. Crash / kill still hit `onclose`.
11. Intentional reconnect nulls `onclose` before `close()`, so that path does not clear.
12. `tts_interrupt` does not flush the queue ([019](019-obs-html-tts-queue.md)).
13. `clearOverlayPlayback` must still empty `alertQueue` itself.
14. TTS and command audio stickers are `.alert-container` ([024](024-obs-html-audio-sticker.md)).

## Consequences

### Positive

* A graceful quit does not leave a looping or stalled overlay clip.
* Crash and a dead socket get the same overlay teardown.
* Chat history is still there when the app starts again.

### Negative and risks

* An HTTP listener restart fades a playing alert.
* A late `play()` after teardown is skipped via `ttsGeneration` (item 10).

### Neutral

* Review notes: `docs/obs-html-review.md` item 3.
