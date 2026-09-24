# 21. Overlay alert placement follows Browser Source resize

* **Status:** Accepted
* **Date:** 2026-09-24
* **Authors:** @oleg

---

## Context

1. Alert position uses `OBS_WIDTH` and `OBS_HEIGHT`.
2. Those used to be set once at script load.
3. OBS can resize the Browser Source without reloading the page.
4. New alerts then placed against the old size.

## Considered options

1. **Keep load-time size** — wrong after a source transform.
2. **Read `innerWidth` / `innerHeight` only in `handleMediaLoad`** — no stored size.
3. **`resize` listener updates `OBS_WIDTH` / `OBS_HEIGHT`** — next alert uses the new size.

## Decision

Use option 3.

1. `OBS_WIDTH` / `OBS_HEIGHT` are `let`.
2. `window` `resize` copies `innerWidth` / `innerHeight`.
3. Already visible alerts stay where they are.
4. Only new alerts use the updated size.

## Consequences

### Positive

* A later Browser Source size change does not need a page reload.

### Negative and risks

* OBS CEF may not fire `resize` for every transform.
* Clips already on screen can sit outside the new bounds until they fade.

### Neutral

* Review notes: `docs/obs-html-review.md` item 11.
