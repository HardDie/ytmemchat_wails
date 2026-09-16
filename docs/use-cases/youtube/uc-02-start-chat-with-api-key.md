# UC-02: Start chat with a valid API key

**Module:** `internal/youtube`  
**Status:** Ported  
**Actors:** Pipeline at Start when settings have a non-empty API key  
**Goal:** Open a Data API v3 live chat iterator and skip history  
**Preconditions:** Stream ID and a valid API key; YouTube Data API v3 enabled  

## Main scenario (happy path)

1. The pipeline calls `youtube.New(apiKey)` then `GetMessageIterator(ctx, streamID)`.
2. `videos.list` resolves `activeLiveChatId`.
3. An initial `liveChatMessages.list` is discarded (history skip) and `nextPageToken` is kept.
4. Later polls emit new `ChatMessage` values, including formatted Super Chats.

## Alternative scenarios and errors

* **1a. Empty key:** `New` returns `ErrEmptyAPIKey` (caller should have used nokey instead).
* **2a. Video not live / no active chat:** `ErrNotLive`. Do not switch to nokey.
* **2b. Quota exceeded:** `ErrQuotaExceeded`. Do not switch to nokey.
* **2c. Invalid key:** see UC-03.

## Postconditions

* The iterator stops when the context is cancelled or YouTube reports `liveChatEnded`.
* Each `videos.list` / `liveChatMessages.list` HTTP response is counted in `internal/youtube/quota` (this process, Pacific day). No extra Data API call is made to read remaining quota.
