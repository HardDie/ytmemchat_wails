# UC-20: Overlay speech badge

**Module:** `http.go` (Home pane)  
**Status:** Implemented  
**Actors:** Operator in the Wails window  
**Goal:** The Home **Overlay speech** badge shows whether the local HTTP listener is bound  
**Preconditions:** The app window is open on Home  

Tracks `OBSStatus.listening` from `GetOBSStatus`.

| Badge | When |
|---|---|
| OBS ready | the local HTTP listener accepted its port |
| OBS offline | the listener is not bound, or status has not loaded |

Green is OBS ready. Red is OBS offline.

**OBS ready** means this app is serving `http://127.0.0.1:<port>/obs/…`.

**Interrupt speech** is enabled only while this badge is OBS ready.

The window reads this status when it opens and after every settings save, including a failed save. It does not poll the listener on a timer.

The same flag drives [UC-21](uc-21-obs-browser-sources-badge.md) and [UC-22](uc-22-test-payload-badge.md). YouTube ingest is [UC-19](uc-19-youtube-live-chat-badge.md).

## Main scenario (happy path)

1. Open the app on a free port. The badge is **OBS ready**. Open the Chat URL: the page loads.
2. **Interrupt speech** is enabled.
3. On Configuration, set the HTTP port to one that is already taken and save. The badge becomes **OBS offline**. The bind error is shown. **Interrupt speech** is disabled. The Chat URL does not load.
4. Save a free port again. The badge returns to **OBS ready**. The Chat URL loads. **Interrupt speech** is enabled.
5. Click Stop on the YouTube card. This badge stays **OBS ready**. The YouTube badge becomes **Stopped**.

## Alternative scenarios and errors

* **`GetOBSStatus` fails on open:** the badge shows **OBS offline**. The window error line shows the failure.
* **Listener exits after startup, with no settings save:** `GetOBSStatus` can still report listening, because the window does not poll it. A failed request to the Chat URL means the badge is stale until the next save or restart.
* **Sibling badges:** Browser Sources and the Test **Payload** badge must show the same listening fact. See UC-21 and UC-22.

## Postconditions

* The badge matches the last `GetOBSStatus.listening` the window loaded.
* Stop cancels YouTube ingest only. It leaves this badge unchanged.
