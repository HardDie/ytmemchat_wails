# UC-15: Estimate Data API quota spend locally

**Module:** `internal/youtube/quota` (wrapped from `internal/youtube` v3 client)  
**Status:** Ported  
**Actors:** Data API v3 client (Start chat or Find latest)  
**Goal:** Count units this process spent without extra Google requests  
**Preconditions:** A non-empty API key so the v3 client is used  

## Main scenario (happy path)

1. `youtube.New` / `LookupLatestBroadcast` wraps the v3 HTTP client with `quota.WrapClient` and attaches the API key on that transport (`WithHTTPClient` does not apply `WithAPIKey`).
2. Each YouTube Data API HTTP response is classified (`videos.list`, `liveChatMessages.list`, `search.list`, or unknown v3).
3. The tracker adds the billed cost (`liveChatMessages.list` is 5 units; other methods this app uses are 1).
4. `Tracker.Snapshot` returns today’s Pacific-Time totals. The app restores/saves `quota.json` next to config.json. No Cloud or extra YouTube call is made.
5. Home shows spent units when a YouTube API key is set (`GetRunStatus`, refreshed every few seconds while the window is open).

## Alternative scenarios and errors

* **1a. No API key:** the no-key client is not wrapped; Home hides the quota line.
* **2a. HTTP 4xx/5xx from Google:** still counted (invalid requests cost at least 1 unit).
* **2b. Transport error, no response:** not counted.
* **3a. Midnight Pacific Time:** counters reset on the next `Record` or `Snapshot` (and the file is rewritten).
* **3b. Unknown `/youtube/v3/…` path:** 1 unit in the default bucket.

## Postconditions

* Spent units are an estimate of this app’s Data API calls today. They are not leftover quota from Google Cloud.
* Changing stream ID does not reset the daily total (Google’s quota is per project per day).
