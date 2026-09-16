# UC-15: Estimate Data API quota spend locally

**Module:** `internal/youtube/quota` (wrapped from `internal/youtube` v3 client)  
**Status:** Ported  
**Actors:** Data API v3 client (Start chat or Find latest)  
**Goal:** Count units this process spent without extra Google requests  
**Preconditions:** A non-empty API key so the v3 client is used  

## Main scenario (happy path)

1. `youtube.New` / `LookupLatestBroadcast` builds the v3 HTTP client through `quota.WrapClient` and `quota.Default`.
2. Each YouTube Data API HTTP response is classified (`videos.list`, `liveChatMessages.list`, `search.list`, or unknown v3).
3. The tracker adds the published cost to the default unit bucket or the search bucket.
4. `Tracker.Snapshot` returns today’s Pacific-Time totals for this process. No Cloud or extra YouTube call is made.

## Alternative scenarios and errors

* **1a. No API key:** the no-key client is not wrapped; snapshot stays unchanged.
* **2a. HTTP 4xx/5xx from Google:** still counted (invalid requests cost at least 1 unit).
* **2b. Transport error, no response:** not counted.
* **3a. Midnight Pacific Time:** counters reset on the next `Record` or `Snapshot`.
* **3b. Unknown `/youtube/v3/…` path:** 1 unit in the default bucket.

## Postconditions

* Spent units are an estimate of this process only. They are not leftover quota from Google Cloud.
