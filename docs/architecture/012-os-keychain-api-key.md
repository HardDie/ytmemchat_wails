# 12. OS keychain for the YouTube API key

* **Status:** Accepted
* **Date:** 2026-09-21
* **Authors:** @oleg

---

## Context

1. Settings JSON is mode `0600` but still plaintext.
2. The API key is the only secret in that file.
3. macOS Keychain, Windows Credential Manager, and libsecret can hold it.

## Considered options

1. **Keep the key only in `config.json`** — simplest; readable on disk.
2. **OS keychain with JSON fallback** — vault when available; file when not.

## Decision

Use option 2.

1. Package `internal/secret`. `config.Store` calls it.
2. Service `ytmemchat`, account `youtube-api-key`.
3. If JSON has a key and the vault works: move it, then save JSON with empty `apiKey`.
4. If the vault fails: leave the file as-is.
5. Config hint: keychain vs settings file.
6. OS backend is `!nomain && !integration`. Tests use Memory or Unavailable.

## Consequences

### Positive

* The key is not in JSON when the vault works.

### Negative and risks

* Linux without a secret service stays on the file.
* Unsigned macOS builds may prompt on first access or after a new binary path.
