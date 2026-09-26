# UC-21: OBS Browser Sources badge

**Module:** `http.go` (Home pane)  
**Status:** Implemented  
**Actors:** Operator in the Wails window  
**Goal:** The Home **OBS Browser Sources** badge shows whether the local HTTP listener is bound  
**Preconditions:** The app window is open on Home. `GetOBSStatus` has returned once, so the card is visible  

Tracks the same `OBSStatus.listening` flag as [UC-20](uc-20-overlay-speech-badge.md).

| Badge | When |
|---|---|
| Listening | the listener is bound |
| Offline | the listener is not bound |

Green is Listening. Red is Offline.

The card appears after the first successful `GetOBSStatus`. A bind error is the `error` string on that card. Chat and Overlay URLs on the card are the addresses to request when checking the badge.

**Listening** means this app is serving those URLs.

The window reads this status when it opens and after every settings save, including a failed save. It does not poll the listener on a timer.

The same flag drives [UC-20](uc-20-overlay-speech-badge.md) and [UC-22](uc-22-test-payload-badge.md).

## Main scenario (happy path)

1. Open the app on a free port. The badge is **Listening**. Open the Chat URL and the Overlay URL: both pages load.
2. On Configuration, set the HTTP port to one that is already taken and save. The badge becomes **Offline**. The bind error is shown on this card. Those URLs do not load.
3. Save a free port again. The badge returns to **Listening**. Both URLs load.
4. Click Stop on the YouTube card. This badge stays **Listening**.

## Alternative scenarios and errors

* **`GetOBSStatus` has not returned:** the card is hidden. There is no badge yet.
* **`GetOBSStatus` fails on open:** the card stays hidden. The window error line shows the failure.
* **Listener exits after startup, with no settings save:** `GetOBSStatus` can still report listening, because the window does not poll it. A failed request to the Chat URL means the badge is stale until the next save or restart.
* **Sibling badges:** Overlay speech and the Test **Payload** badge must show the same listening fact. See UC-20 and UC-22.

## Postconditions

* The badge matches the last `GetOBSStatus.listening` the window loaded.
* Stop cancels YouTube ingest only. It leaves this badge unchanged.
