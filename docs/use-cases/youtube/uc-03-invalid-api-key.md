# UC-03: Invalid API key does not fall back

**Module:** `internal/youtube`  
**Status:** Implemented  
**Actors:** Pipeline at Start when the user saved a non-empty API key that YouTube rejects  
**Goal:** Fail with `ErrInvalidAPIKey` so the UI can show “invalid API key”  
**Preconditions:** API key field is non-empty after trim  

## Main scenario (happy path)

1. The pipeline calls `youtube.New(key)` (succeeds locally) then `GetMessageIterator`.
2. YouTube returns `keyInvalid` / “API key not valid” (HTTP 400/401) or a disabled-API 403.
3. The error wraps `ErrInvalidAPIKey` (`youtube.IsInvalidAPIKey` is true).
4. The pipeline **does not** construct `nokey.New()`. The Wails window shows that the key is invalid.

## Alternative scenarios and errors

* **2a. Not live:** `ErrNotLive` — still no nokey fallback (key was provided).
* **2b. Quota:** `ErrQuotaExceeded` — still no nokey fallback.
* **2c. Network error:** wrapped transport error — still no nokey fallback.

## Postconditions

* Only an empty/whitespace key selects nokey. A bad key stays on the v3 path until the user clears or fixes it.
