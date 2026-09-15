# UC-01: Start chat without an API key

**Module:** `internal/youtube/nokey`  
**Status:** Ported  
**Actors:** Pipeline at Start when settings have an empty API key  
**Goal:** Open a live chat iterator using HTML/innertube, skip history  
**Preconditions:** A live video ID is set; no API key (or only whitespace)  

## Main scenario (happy path)

1. The pipeline calls `nokey.New().GetMessageIterator(ctx, streamID)`.
2. The client GET `live_chat?v=` and extracts a continuation token.
3. History is not replayed; polling starts from that continuation.
4. New text messages are pushed as `youtube.ChatMessage` on the iterator.

## Alternative scenarios and errors

* **2a. Page has no continuation (offline / wrong id):** GetMessageIterator returns an error. The v3 client is not used as a fallback (no key was set).
* **3a. get_live_chat HTTP error:** the poller logs and retries; context cancel stops it.

## Postconditions

* Callers consume the same `youtube.Client` / `ChatMessage` types as the API v3 client.
