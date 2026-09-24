# 23. Chat avatar URL failure uses the letter canvas

* **Status:** Accepted
* **Date:** 2026-09-24
* **Authors:** @oleg

---

## Context

1. Chat rows set `img.src` from `authorPicture`.
2. A URL without the substring `placeholder` was treated as good.
3. A 403 or dead YouTube photo stayed as a broken image.
4. `generateAvatar` already draws a letter when there is no URL.

## Considered options

1. **Leave the broken `img`** — empty hole on stream.
2. **`img.onerror` → `generateAvatar`** — same letter as missing photos.
3. **Hide the avatar on error** — layout jumps.

## Decision

Use option 2.

1. `onerror` is set before `src`.
2. On error, clear `onerror` first (no loop on the canvas URL).
3. Then set `src` to `generateAvatar(authorName)`.
4. If that throws, remove `src`.
5. Missing / `placeholder` URLs use the same fallback path.
7. If `getContext('2d')` is null, `generateAvatar` returns `''`.

## Consequences

### Positive

* A failed YouTube photo still shows a letter.

### Negative and risks

* Empty `src` if canvas 2d is unavailable.

### Neutral

* Review notes: `docs/obs-html-review.md` items 16 and 17.
