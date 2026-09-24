# 22. Chat overlay caps DOM rows; query can disable

* **Status:** Accepted
* **Date:** 2026-09-24
* **Authors:** @oleg

---

## Context

1. Each chat line is kept in `#chat-container`.
2. Only `chat_flush` cleared rows.
3. A long Browser Source session grew the DOM without bound.

## Considered options

1. **Keep every row** — unbounded memory.
2. **Hard cap in JS** — operators cannot keep a full night.
3. **Default cap 100; `?cap=` overrides; no-cap is explicit** — OBS URL.

## Decision

Use option 3.

1. Default cap is 100 newest rows.
2. Oldest DOM child is dropped (`lastElementChild`).
3. Query `cap` is a positive integer (example `cap=50`).
4. `cap=0`, `cap=none`, `cap=off`, or `cap=unlimited` means no cap.
5. Missing or invalid `cap` uses 100.
6. `chat_flush` still clears all rows.

## Consequences

### Positive

* Default OBS chat stays bounded.
* Operators can keep every line with one query flag.

### Negative and risks

* No cap can still grow the DOM for a long stream.

### Neutral

* Review notes: `docs/obs-html-review.md` item 15.
* Operator URL: `docs/wiki/Chat-URL.md`.
