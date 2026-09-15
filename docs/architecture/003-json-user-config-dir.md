# 3. Persistent settings as JSON under `os.UserConfigDir`

* **Status:** Accepted
* **Date:** 2026-09-15
* **Authors:** @oleg

---

## Context

Wails v2 has no built-in settings store. The console app uses a committed-style `.env` and panics on missing variables. A desktop app must survive first launch, store an optional API key, and survive force-quit.

## Considered options

1. **`.env` next to the binary** — matches the console app; fails after install (read-only `.app` / Program Files).
2. **webview `localStorage`** — simple; poor for secrets; can be wiped with webview data.
3. **JSON file under `os.UserConfigDir()/ytmemchat/config.json`** — OS-standard app config dir; Go owns read/write.
4. **SQLite** — useful for searchable data; overkill for a settings form.
5. **OS keychain for the API key** — better secret storage; extra platform code.

## Decision

Use option 3. Load in `OnStartup`. Save on every successful `SaveSettings`, not only `OnShutdown`. Atomic write (temp file + rename), file mode `0600`. Defaults live in code. Missing file is first launch. Stream ID is required to Start; API key is optional.

OS keychain is a later hardening step, not the first slice.

## Consequences

### Positive

* Same path helper on macOS, Windows, and Linux.
* Svelte is only a form over `GetSettings` / `SaveSettings`.
* No panic on missing config (unlike console `config.Get()`).

### Negative and risks

* API key sits on disk (mitigated by `0600`, not by encryption).
* Force-quit still loses unsaved form edits if the user never clicked save.

### Neutral

* Window size may be stored in the same JSON later.
