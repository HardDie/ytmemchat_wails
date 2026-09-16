# UC-12: Find latest live or upcoming stream

**Module:** `internal/youtube` (`LookupLatestBroadcast`) and Wails `LookupLatestStream`  
**Status:** Ported  
**Actors:** Operator in the Wails window  
**Goal:** Fill stream ID from a previous video without copying a new watch URL  
**Preconditions:** A Data API v3 key is set (form or saved). A previous video ID from the same channel is set.  

## Main scenario (happy path)

1. The window calls `LookupLatestStream` with the current stream ID and API key (empty key uses the saved key).
2. `videos.list` (`snippet`, `liveStreamingDetails`) loads the channel ID.
3. If that video is still live (`activeLiveChatId`, or started with no end time), that ID is returned as kind `live`.
4. Otherwise `search.list` `eventType=live` for the channel; if a video exists, it is returned as `live`.
5. Otherwise `search.list` `eventType=upcoming` `order=date`; the newest scheduled video is returned as `upcoming`.
6. The window writes the video ID into the form. Settings are not saved until Save or Start.

## Alternative scenarios and errors

* **1a. Empty stream ID:** error; lookup does not run.
* **1b. Empty API key (form and saved):** error asking for a YouTube API key. The no-key client cannot list upcoming streams.
* **2a. Unknown video / invalid key / quota:** mapped API errors; the key is never logged or returned.
* **4–5a. Channel has only completed VODs:** `ErrNoBroadcast`.

## Postconditions

* On success the form stream ID is a live or upcoming video, not a recording. Disk is unchanged until Save/Start.
* `videos.list` is counted in the default unit bucket; each `search.list` is counted in the separate search bucket (`internal/youtube/quota`).
