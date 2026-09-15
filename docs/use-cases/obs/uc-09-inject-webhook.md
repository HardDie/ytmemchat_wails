# UC-09: Inject a test chat message

**Module:** `internal/obs` (`/api/webhook`)  
**Status:** Ported  
**Actors:** Operator (curl), pipeline reading `Injected()`  
**Goal:** POST a JSON message that the pipeline can treat like a chat line  
**Preconditions:** Server started with `Webhooks: true`  

## Main scenario (happy path)

1. Operator POSTs `/api/webhook` with `{"message": "@jump"}`.
2. Handler responds 204 and sends [InjectedMessage] (`Author` `webhook`, `Text` from JSON) on `Injected()`.
3. `app.go` maps that onto the same chat/alerts/TTS path as YouTube (`dispatchChat`) while OBS HTTP is up. YouTube Start is not required.

## Alternative scenarios and errors

* **Webhooks disabled:** route is not registered (404).
* **Invalid JSON:** 400.
* **GET:** 405.
* **Inject queue full:** 503; the previous file/channel is not overwritten.

## Postconditions

* Handler does not import `youtube`. It does not call alerts or TTS itself; the process-level inject loop does.
