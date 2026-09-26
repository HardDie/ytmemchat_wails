# OBS HTML review (`overlay.html`, `chat.html`)

Working notes. Not an ADR. Not a wiki how-to.

Files:

1. `internal/obs/overlay.html`
2. `internal/obs/chat.html`

How to use this list:

1. Discuss one item at a time in chat.
2. After a decision, set that item’s status.
3. Do not start a code change until the item is `fix`.
4. Status values: `open`, `discussing`, `skip`, `fix`, `done`.

---

## Shared

1. Dummy WebSocket on page load — `done`
  1. Both files open a socket with no handlers.
  2. Chat replaces it after 300ms.
  3. Overlay replaces it after the audio button.
  4. Decision: [ADR 014](architecture/014-obs-html-one-websocket.md).
2. Reconnect timer is not cancelled — `done`
  1. `onclose` always `setTimeout(connectWebSocket, 500)`.
  2. Overlap can mean two sockets.
  3. Decision: [ADR 015](architecture/015-obs-html-reconnect-timer.md).
3. `app_closed` only updates the banner — `done`
  1. No media / TTS / chat teardown.
  2. Socket then drops and retry continues.
  3. Overlay now calls `clearOverlayPlayback` on `app_closed`.
  4. Chat rows stay across restarts.
  5. Decision: [ADR 016](architecture/016-obs-html-app-closed-teardown.md).
  6. Overlay `onclose` also calls `clearOverlayPlayback` (crash / drop).
4. Global error handlers reload after 5s — `done`
  1. `window.onerror` and `unhandledrejection`.
  2. Recovers a wedged page.
  3. Also wipes on-screen chat or a playing alert.
  4. No `location.reload`.
  5. Recovery is `recoverFromUnhandledError` → `connectWebSocket`.
  6. Overlay also `clearOverlayPlayback`.
  7. Chat rows stay.
  8. 2s cooldown avoids a reconnect loop.
  9. Decision: [ADR 017](architecture/017-obs-html-recover-no-reload.md).
5. WebSocket `onerror` logs `err.message` — `done`
  1. That event is not an `Error`.
  2. The log is usually `undefined`.
  3. Now logs the event object; still `close()` then `onclose` retries.

---

## Overlay (`overlay.html`)

6. Video hide waits on `ended`, handler attached late — `done`
  1. `src` is set before listeners.
  2. `onended` is set inside `handleMediaLoad`.
  3. App quit can stall `/obs/media/` so `ended` never fires.
  4. A cached clip can finish before `onended` is attached.
  5. OBS does not reload the page on app restart.
  6. Next Preview can stack on the old frame.
  7. Do not remove on `ended` (memealerts min time).
  8. `scheduleVideoTeardown`: `max(duration, 5s)` then fade.
  9. Decision: [ADR 018](architecture/018-obs-html-min-sticker-time.md).
7. No `error` / stall fallback for video — `done`
  1. 404, decode error, or hang has no timeout.
  2. `error` / `abort` now remove the container.
  3. If `handleMediaLoad` never runs, video 5s / image 6s backstop.
  4. After layout, item 6 timer still applies (`laidOut` skips the backstop).
  5. Same ADR: [018](architecture/018-obs-html-min-sticker-time.md).
  6. Audio that never gets `loadedmetadata` is dropped after 5s.
8. `.gif` is treated as `<video>` — `done`
  1. Regex is `mp4|webm|mov|gif`.
  2. GIF in a video tag often never `ended`.
  3. Now `mp4|webm|mov` only. `.gif` uses the `<img>` 6s path.
  4. Decision: [ADR 018](architecture/018-obs-html-min-sticker-time.md).
9. Alerts are not queued — `done`
  1. TTS uses `alertQueue`.
  2. Alerts call `displayMediaAlert` immediately.
  3. TTS queue: [ADR 019](architecture/019-obs-html-tts-queue.md).
  4. Overlap is intended. No interrupt for command clips.
  5. Decision: [ADR 020](architecture/020-obs-html-alert-overlap.md).
  6. No overlay JS change.
10. TTS interrupt vs in-flight `decodeAudioData` — `done`
  1. Interrupt clears the queue and `stop()`s the current source.
  2. A decode that finishes later can still `start()`.
  3. Interrupt no longer flushes the queue (item 22). This item is the decode race.
  4. `ttsGeneration` bumps on interrupt and teardown.
  5. Stale decode does not `start()` and does not call `onMediaFinished`.
  6. Decision: [ADR 019](architecture/019-obs-html-tts-queue.md).
11. `OBS_WIDTH` / `OBS_HEIGHT` captured once — `done`
  1. Browser Source resize uses the old size for placement.
  2. `resize` now copies `innerWidth` / `innerHeight`.
  3. Already visible alerts stay put.
  4. Decision: [ADR 021](architecture/021-obs-html-resize-placement.md).
12. Overlay HTML has no `charset` — `done`
  1. Chat sets UTF-8.
  2. Overlay status text uses emoji.
  3. Stream content is media + TTS only. No chat lines.
  4. HTTP already sends `charset=utf-8`.
  5. Overlay now has the same `<meta charset="UTF-8">` as chat.
13. Audio-only alerts only listen for `ended` — `done`
  1. Same stall class as video.
  2. Nothing visible on the overlay.
  3. `error` / `abort` and a 5s never-ready drop are on item 7.
  4. After metadata, `scheduleAudioTeardown` uses `max(duration, 5s)`.
  5. No fade subtract. `ended` can still drop earlier.
  6. Decision: [ADR 018](architecture/018-obs-html-min-sticker-time.md).

---

## Chat (`chat.html`)

14. `JSON.parse` has no try/catch — `done`
  1. Overlay wraps parse.
  2. Chat does not.
  3. Bad payload can reload the source in 5s.
  4. Chat now catches parse and returns.
  5. Decision: [ADR 017](architecture/017-obs-html-recover-no-reload.md).
  6. Overlay parse is the same try/catch; dispatch is outside.
15. No message cap — `done`
  1. Every line is kept in the DOM.
  2. Default cap is 100 newest rows.
  3. `?cap=` sets the cap. `0` / `none` / `off` / `unlimited` means no cap.
  4. Decision: [ADR 022](architecture/022-obs-html-chat-cap.md).
16. Broken avatars have no fallback — `done`
  1. A 403 YouTube image stays broken.
  2. Placeholder check is only the substring `placeholder`.
  3. `img.onerror` swaps to `generateAvatar`.
  4. `onerror` is cleared first so the canvas URL cannot loop.
  5. Decision: [ADR 023](architecture/023-obs-html-chat-avatar-fallback.md).
17. `generateAvatar` assumes `canvas.getContext('2d')` — `done`
  1. Null context throws.
  2. That hits the 5s reload.
  3. Null context now returns `''`.
  4. Decision: [ADR 023](architecture/023-obs-html-chat-avatar-fallback.md).
18. `onopen` log says media alerts — `done`
  1. Copy-paste from overlay.
  2. Chat now logs `Ready for chat.`
19. `textColor` always prefixes `#` — `done`
  1. Wiki says hex without `#`.
  2. A value that already has `#` becomes `##fff`.
  3. A leading `#` is kept. Otherwise one is prepended.
20. Comment vs layout — `done`
  1. Comment says new items stay at the top.
  2. `column-reverse` + `prepend` puts newest at the bottom.
  3. Comments now say newest at the bottom. No layout change.

---

## Reconnect wait (both pages)

21. `RETRY_INTERVAL` 500ms vs reconnects that take seconds — `done`
  1. 500ms is the pause after `onclose`, not total reconnect time.
  2. `new WebSocket` stays `CONNECTING` until TCP / upgrade fail.
  3. That handshake wait is Chromium, not our timer.
  4. A dead `127.0.0.1` port is often fast.
  5. IPv6 `localhost`, SYN retries, or OBS CEF can take 3–20s.
  6. Then we wait 500ms and try again.
  7. Exact ~5s can be the `onerror` reload (item 4).
  8. Server `httpShutdownTimeout` is also 5s (listener restart).
  9. Hidden OBS sources can stretch `setTimeout(500)`.
  10. Do not raise `RETRY_INTERVAL` to 5s.
  11. A later fix would abort a stuck `CONNECTING` socket.
  12. Item 4 no longer reloads; a 5s wait is not this handler.
  13. Watchdog 2500ms `close()`s `CONNECTING`. Retry still `onclose`.
  14. Decision: [ADR 015](architecture/015-obs-html-reconnect-timer.md).

22. TTS interrupt clears the whole queue — `done`
  1. `interruptCurrentTTS` empties `alertQueue` and `stop()`s the source.
  2. Wanted: skip only the utterance that is playing.
  3. Remaining queued TTS should still play, one by one.
  4. Teardown (`app_closed`, socket close, crash) can still dump the queue.
  5. Decision: [ADR 019](architecture/019-obs-html-tts-queue.md).
  6. Related decode race: item 10.
  7. Interrupt now `stop()`s the source and `processQueue`s. Queue stays.
  8. `onended` ignores a source that is no longer `currentTtsSource`.
  9. `clearOverlayPlayback` still assigns `alertQueue = []`.
