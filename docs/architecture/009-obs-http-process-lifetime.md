# 9. OBS HTTP and WebSockets live with the Wails process

* **Status:** Accepted
* **Date:** 2026-09-15
* **Authors:** @oleg

---

## Context

OBS Browser Sources need a stable page and a child WebSocket.

1. Page: `http://127.0.0.1:<port>/obs/…`
2. If those only exist while a YouTube session is “running”, sources go blank on Stop.
3. Reconnect loops fire.
4. The streamer cannot set up OBS before going live.
5. The Wails window is already a long-lived process.
6. Start/Stop is for ingesting live chat, not for whether OBS can load the overlay.

## Considered options

1. **Listen on Start, close on Stop** — matches a “session = whole pipeline” mental model; OBS disconnects every Stop.
2. **Listen from Wails `OnStartup` until process exit**
   1. Pages and sockets are always there.
   2. Start only runs the YouTube iterator and fans messages into those sockets.
3. **Listen always, but drop WebSocket upgrades until Start** — pages load, sockets fail until Start; worse than 2 for OBS reconnect.

## Decision

Use option 2.

1. **Process (Wails running)**
   1. `internal/obs` serves HTML.
   2. `/obs/chat/ws`, `/obs/overlay/ws`
   3. `/obs/media/` if configured
   4. `/api/…` if enabled
   5. OBS may connect with no live iterator.
2. **Session (Start / Run)**
   1. YouTube `GetMessageIterator`
   2. Fan-out: chat WS, alerts vs TTS → overlay WS
   3. Stop cancels that iterator only.
   4. Connected OBS clients stay up.
   5. They just stop receiving new chat/alert/TTS events until the next Start.
3. Changing listen port (and possibly media path / webhook flag) may restart the HTTP listener.
4. That restart is settings, not Stop. It does not require a YouTube Start.

## Consequences

### Positive

* OBS sources can be added as soon as the desktop app is open.
* Stop does not tear down Browser Sources.
* Start/Stop stays a small control: iterator + fan-out.

### Negative and risks

* The listen port is occupied for the whole time the app is open (usually what we want).
* A port change needs a listener restart. In-flight OBS connections drop once, then reconnect.

### Neutral

1. `obs.Server` already supports listen independent of YouTube.
2. `ListenAndServe` / `Serve` vs `PublishChat` / `PublishOverlay`.
3. Wiring belongs in `app.go` `OnStartup` / `OnShutdown`.
