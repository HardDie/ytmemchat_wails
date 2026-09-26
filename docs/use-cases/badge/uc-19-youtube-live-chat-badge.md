# UC-19: YouTube live chat badge

**Module:** `pipeline.go` (Home pane)  
**Status:** Implemented  
**Actors:** Operator in the Wails window  
**Goal:** The Home **YouTube live chat** badge shows ingest state  
**Preconditions:** The app window is open on Home  

Tracks `RunStatus` from `GetRunStatus`. The window applies each `pipeline` event as it arrives, and polls `GetRunStatus` every 2 seconds.

| Badge | When |
|---|---|
| Unknown | `RunStatus` has not loaded yet |
| Connecting | `connecting` is true (Start has begun; the iterator is not ready) |
| Connected | `running` is true (the iterator is live) |
| Stopped | loaded, and neither `connecting` nor `running` |

Green is Connected. Yellow is Connecting. Stopped and Unknown use the plain badge.

The line under the title is separate. While Connected it is **YouTube Data API v3** or **No API key**. An error string under the card is `RunStatus.error`. The badge stays Stopped when that error is set and the iterator is not running.

The badge is ingest state. OBS listener badges are [UC-20](uc-20-overlay-speech-badge.md), [UC-21](uc-21-obs-browser-sources-badge.md), and [UC-22](uc-22-test-payload-badge.md).

## Main scenario (happy path)

1. Leave Start unclicked after the window loads. The badge is **Stopped**.
2. Save a live stream ID and click Start. The badge becomes **Connecting**, then **Connected**. A new viewer line shows on the chat page.
3. Click Stop. The badge returns to **Stopped**. The chat page still loads. New YouTube lines stop.

## Alternative scenarios and errors

* **Empty stream ID:** save a blank stream ID and click Start. The badge stays **Stopped**. The card shows an error. It does not remain on Connecting.
* **Invalid API key:** the badge may pass through **Connecting**, then return to **Stopped** with an error. It does not stay **Connected**.
* **`GetRunStatus` fails on open:** the badge stays **Unknown**. The window error line shows the failure.
* **Iterator ends without Stop:** `running` clears. The badge becomes **Stopped**.
* **Stop while Connected:** the OBS badges on Home stay on their ready labels. Only this badge changes.

## Postconditions

* The badge matches `connecting` / `running` from the last `RunStatus` the window holds.
