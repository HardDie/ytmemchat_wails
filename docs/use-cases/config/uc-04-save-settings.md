# UC-04: Save and reload settings

**Module:** `internal/config`  
**Status:** Ported  
**Actors:** Settings UI (Wails `GetSettings` / `SaveSettings`)  
**Goal:** Persist user settings on disk and get the same values back after restart  
**Preconditions:** Process can create the config directory (or a test uses a temp path)  

## Main scenario (happy path)

1. The UI calls `SaveSettings` with stream ID, optional API key, port, TTS, alerts, and webhook fields.
2. The package validates port and (if alerts enabled) a single-character token. Empty stream ID is allowed on save.
3. JSON is written atomically to `os.UserConfigDir()/ytmemchat/config.json` (or the store path) with mode `0600`.
4. Later, `Store.Load` returns the same trimmed fields. Missing file yields defaults (port `8080`, alert token `@`, TTS on, webhook off).

## Alternative scenarios and errors

* **2a. Invalid port:** Save returns `ErrInvalidPort`; the previous file is left unchanged.
* **2b. Alerts enabled and token is not one character:** Save returns `ErrAlertToken`.
* **4a. File missing (first launch):** Load returns [Defaults] without error. `CanStart` still fails with `ErrStreamIDRequired` until a stream ID is set.
* **4b. Corrupt JSON:** Load returns an error; it does not panic and does not overwrite the file.
* **Start with empty API key:** `HasAPIKey` is false (no-key YouTube client). Whitespace-only key counts as empty.
* **Start with empty stream ID:** `CanStart` returns `ErrStreamIDRequired`. Errors never include the API key.

## Postconditions

* On successful save, `config.json` exists and Load round-trips. No `.env` is required. The process does not panic on missing config.
