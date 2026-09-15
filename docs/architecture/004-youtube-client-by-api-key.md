# 4. YouTube client chosen by API key; no fallback on invalid key

* **Status:** Accepted
* **Date:** 2026-09-15
* **Authors:** @oleg

---

## Context

The console app has two `youtube.Client` implementations: Data API v3 (API key) and a no-key live chat HTML client (`youtubev1`). Console `main.go` currently hardcodes v3. Streamers may not have a Google Cloud key. A typed-but-wrong key must not silently switch to scraping.

## Considered options

1. **Always API v3** — requires a key for everyone.
2. **Always no-key client** — quota-free; more fragile; worse for Super Chat typing if the HTML client lags behind.
3. **Empty key → no-key client; non-empty key → v3 only.** Invalid key → error in the Wails window; never fall back.

## Decision

Use option 3. Trim the saved key at Start. Both implementations emit the same `ChatMessage` type. “Not live”, quota, and network errors are distinct from “invalid API key”; none of them may start `youtube/nokey` when a key was provided.

## Consequences

### Positive

* First run works with only a video ID.
* A bad key is visible and fixable (clear or replace the key).
* Downstream chat/alerts/TTS do not care which client produced the message.

### Negative and risks

* The no-key client depends on YouTube HTML and can break without a code change.
* Users might think an empty key is “anonymous API v3”; the settings UI must state the two modes.

### Neutral

* Packages in this repo: `internal/youtube` (v3 + types) and `internal/youtube/nokey`.
