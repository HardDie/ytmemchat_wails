# UC-17: Store the YouTube API key in the OS keychain

**Module:** `internal/secret`  
**Status:** Implemented  
**Actors:** Operator (Configuration pane), `config.Store`  
**Goal:** Keep the Data API key out of `config.json` when the OS vault works  
**Preconditions:** Settings can be saved. Vault may or may not be available.  

## Main scenario (happy path)

1. Operator saves a non-empty API key.
2. `Store` writes it to the vault and writes JSON with empty `apiKey`.
3. Load returns the key in memory. Config hint: API key stored in the OS keychain.

## Alternative scenarios and errors

* **Vault down:** key stays in `config.json` (mode `0600`). Hint: stored in the settings file (keychain unavailable).
* **JSON still has a key and vault works:** move it into the vault, rewrite JSON with empty `apiKey`.
* **Vault Set fails:** leave the file as-is. Hint is file.

## Postconditions

* The in-memory settings still have the key for YouTube Start.
* Tests do not need a real OS keychain (`secret.Memory` / `secret.Unavailable`).
