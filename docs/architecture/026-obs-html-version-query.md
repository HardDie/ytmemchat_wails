# 26. OBS pages and sockets carry the app version

* **Status:** Accepted
* **Date:** 2026-09-24
* **Authors:** @oleg

---

## Context

1. OBS Browser Source caches `overlay.html`, `chat.html`, and `script.js`.
2. A rebuilt binary can keep serving the old page.
3. The cache key is the full URL, including the query.

## Considered options

1. **`Cache-Control: no-store`** — always hits the server. No version in the URL.
2. **`?v=` equals the running app version** — a new tag or commit misses the cache.

## Decision

Use option 2.

1. Query name is `v`.
2. Value is `AppVersion` (release tag, or commit when the build is untagged).
3. Unstamped builds use `dev`.
4. `GET /obs/chat` and `GET /obs/overlay` redirect (302) when `v` differs.
5. Other query keys stay (`cap`, `transparent`, `debugAudio`).
6. The redirect target is the same path with the running `v`.
7. `/obs/chat/ws` and `/obs/overlay/ws` read the same `v`.
8. A mismatch upgrades the socket and sends `version_redirect`.
9. `url` is the page path with the running `v`.
10. `script.js` sets `location` to that URL and does not reconnect.
11. `script.js` is requested as `/obs/script.js?v=…`.
12. The token in the HTML files is replaced when the page is served.

## Consequences

### Positive

* A new build loads fresh HTML and `script.js`.
* A cached page still corrects itself on the next socket connect.

### Negative and risks

* OBS source URLs copied in the app stay without `v`. The server adds it.

### Neutral

* `301` is not used. A permanent redirect would stick to an old version.
