# 32. OBS HTTP listens on loopback; LAN is opt-in; requests are guarded

* **Status:** Accepted
* **Date:** 2026-10-06
* **Authors:** @oleg

---

## Context

1. The listener used `:<port>`, so every interface was open.
2. Docs and Home only ever showed `127.0.0.1` URLs.
3. LAN hosts could reach the webhook, media folder, and both sockets.
4. Browsers let any site send a simple POST to `127.0.0.1`.
5. That injected TTS lines and alerts when the HTTP API was on.
6. DNS rebinding lets a site read this server as same-origin.
7. Some streamers play on a second PC with headphones.
8. That PC needs the overlay page to hear alerts and TTS.

## Considered options

1. **Keep `:<port>`** — LAN stays open with no opt-in.
2. **Loopback only** — breaks the second-PC setup.
3. **Loopback by default, LAN toggle, request guard** — both setups work.
4. **Require `Content-Type: application/json`** — breaks `curl -d` without `-H`.

## Decision

Use option 3.

1. `config.Server.LAN` (`"lan"` in JSON) defaults to false.
2. Off listens on `127.0.0.1:<port>`; on listens on `:<port>`.
3. Toggling it restarts the listener like a port change.
4. `obs.guard` wraps every route.
5. `Host` must be an IP literal or `localhost`. Other names get 403.
6. That blocks DNS rebinding in both modes.
7. Non-GET requests with a foreign `Origin` get 403.
8. No `Origin` passes, so curl and scripts keep working.
9. WebSockets use gorilla’s default same-origin check.
10. OBS and a LAN browser pass: their page comes from this server.
11. Home still shows only `127.0.0.1` URLs.

## Consequences

### Positive

* Other devices cannot reach the server unless the operator opts in.
* Web pages cannot inject chat lines or read the overlay sockets.

### Negative and risks

* Existing LAN users must turn the toggle on after upgrade.
* A LAN browser must use the PC’s IP, not a host name.
* On a hostile LAN with the toggle on, curl can still reach `/api/…`.

### Neutral

* `Server.Handler` returns the guarded handler, so tests see the guard.
