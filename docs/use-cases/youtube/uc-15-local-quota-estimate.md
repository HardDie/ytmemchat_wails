# UC-15: Estimate Data API quota spend locally

**Module:** `internal/youtube/quota` (wrapped from `internal/youtube` v3 client)  
**Status:** Ported  
**Actors:** Data API v3 client (Start chat or Find latest)  
**Goal:** Count units this process spent without extra Google requests  
**Preconditions:** A non-empty API key so the v3 client is used  

## Main scenario (happy path)

1. `youtube.New` / `LookupLatestBroadcast` wraps the v3 HTTP client with `quota.WrapClient` and attaches the API key on that transport (`WithHTTPClient` does not apply `WithAPIKey`).
2. Each YouTube Data API HTTP response is classified (`videos.list`, `liveChatMessages.list`, `search.list`, or unknown v3).
3. The tracker adds the published cost to the default unit bucket or the search bucket.
4. `Tracker.Snapshot` returns today’s Pacific-Time totals for this process. No Cloud or extra YouTube call is made.
5. Home shows spent units when a YouTube API key is set (`GetRunStatus`, refreshed every few seconds while the window is open).
6. Saving, Start, or a successful Find latest with a **different** video ID calls `Tracker.Reset` so the line is spend for the current stream.

## Alternative scenarios and errors

* **1a. No API key:** the no-key client is not wrapped; Home hides the quota line.
* **2a. HTTP 4xx/5xx from Google:** still counted (invalid requests cost at least 1 unit).
* **2b. Transport error, no response:** not counted.
* **3a. Midnight Pacific Time:** counters reset on the next `Record` or `Snapshot`.
* **3b. Unknown `/youtube/v3/…` path:** 1 unit in the default bucket.
* **6a. Same stream ID:** counters are left unchanged (port/TTS edits do not reset).

## Postconditions

* Spent units are an estimate of this process only. They are not leftover quota from Google Cloud.
* After a stream ID change the estimate is zero until the next v3 HTTP response.
