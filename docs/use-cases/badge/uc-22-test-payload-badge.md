# UC-22: Test payload badge

**Module:** `http.go` (Test pane)  
**Status:** Implemented  
**Actors:** Operator in the Wails window  
**Goal:** The Test **Payload** badge shows whether the local HTTP listener is bound  
**Preconditions:** The app window is open on Test  

Tracks the same `OBSStatus.listening` flag as [UC-20](uc-20-overlay-speech-badge.md) and [UC-21](uc-21-obs-browser-sources-badge.md).

| Badge | When |
|---|---|
| Ready | the listener is bound |
| OBS offline | the listener is not bound, or status has not loaded |

Green is Ready. Red is OBS offline.

**Ready** means this app is serving `http://127.0.0.1:<port>/obs/…`. YouTube Start is not required.

**Send** is enabled only while this badge is Ready and the message is non-empty. **Flush chat** is enabled only while this badge is Ready.

The window reads this status when it opens and after every settings save, including a failed save. It does not poll the listener on a timer.

## Main scenario (happy path)

1. Open the app on a free port and open Test. The badge is **Ready**. Open the Chat URL: the page loads.
2. Type a message. **Send** and **Flush chat** are enabled.
3. On Configuration, set the HTTP port to one that is already taken and save. Return to Test. The badge is **OBS offline**. The bind error is shown. **Send** and **Flush chat** are disabled. The Chat URL does not load.
4. Save a free port again. The badge returns to **Ready**. The Chat URL loads. With a message typed, **Send** is enabled.
5. On Home, click Stop. Return to Test. This badge stays **Ready**.

## Alternative scenarios and errors

* **`GetOBSStatus` fails on open:** the badge shows **OBS offline**. The window error line shows the failure.
* **Empty message:** the badge can be **Ready** while **Send** stays disabled. **Flush chat** stays enabled.
* **Listener exits after startup, with no settings save:** `GetOBSStatus` can still report listening, because the window does not poll it. A failed request to the Chat URL means the badge is stale until the next save or restart.
* **Sibling badges:** Home Overlay speech and OBS Browser Sources must show the same listening fact. See UC-20 and UC-21.

## Postconditions

* The badge matches the last `GetOBSStatus.listening` the window loaded.
* Stop cancels YouTube ingest only. It leaves this badge unchanged.
