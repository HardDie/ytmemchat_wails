# UC-07: Match an alert command

**Module:** `internal/alerts`  
**Status:** Ported  
**Actors:** Pipeline, for each chat line  
**Goal:** If the message contains a configured command after the token, emit a media [Clip] and skip TTS  
**Preconditions:** `commands.yaml` loaded; `Out` is a non-nil buffered or received channel; token is non-empty  

## Main scenario (happy path)

1. Chat text is passed to `Alerts.Alert`.
2. `findToken` takes the first word after the token (for example `@jump` → `jump`).
3. The name is matched case-insensitively to YAML `commands[].name`.
4. A `Clip` is sent (`Filename`, `Volume` default 1, `Scale` default 1).
5. Alert returns true; the pipeline does not call TTS.

## Alternative scenarios and errors

* **2a. No token / empty word after token:** returns false; nothing sent.
* **3a. Unknown command:** returns false; nothing sent (TTS may run).
* **Load: duplicate names:** `New` fails (case-insensitive).
* **Load: missing or invalid YAML:** `New` returns a read or decode error.
* **Load: nil Out or empty token:** `New` returns an error.

## Postconditions

* Overlay mapping (`type: alert`, `/obs/media/<file>`) is not in this package; obs consumes `Clip` later.
