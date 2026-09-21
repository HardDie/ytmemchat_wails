# 3. Persistent settings as JSON under `os.UserConfigDir`

* **Status:** Accepted
* **Date:** 2026-09-15
* **Authors:** @oleg

---

## Context

Wails v2 has no built-in settings store.

1. The console app uses a committed-style `.env`.
2. It panics on missing variables.
3. A desktop app must survive first launch.
4. It must store an optional API key.
5. It must survive force-quit.

## Considered options

1. **`.env` next to the binary** — matches the console app; fails after install (read-only `.app` / Program Files).
2. **webview `localStorage`** — simple; poor for secrets; can be wiped with webview data.
3. **JSON file under `os.UserConfigDir()/ytmemchat/config.json`** — OS-standard app config dir; Go owns read/write.
4. **SQLite** — useful for searchable data; overkill for a settings form.
5. **OS keychain for the API key** — better secret storage; extra platform code.

## Decision

Use option 3.

1. Load in `OnStartup`.
2. Save on every successful `SaveSettings`, not only `OnShutdown`.
3. Atomic write (temp file + rename), file mode `0600`.
4. Defaults live in code.
5. Missing file is first launch.
6. Stream ID is required to Start. API key is optional.
7. API key uses the OS keychain when available ([012](012-os-keychain-api-key.md)).

## Consequences

### Positive

* Same path helper on macOS, Windows, and Linux.
* Svelte is only a form over `GetSettings` / `SaveSettings`.
* No panic on missing config (unlike console `config.Get()`).

### Negative and risks

* If the keychain is unavailable, the API key stays in the JSON file (`0600`).
* Force-quit still loses unsaved form edits if the user never clicked save.

### Neutral

* Window size may be stored in the same JSON later.
* API key keychain is [012](012-os-keychain-api-key.md).
