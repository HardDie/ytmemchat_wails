// Package youtube normalizes YouTube Live Chat into [ChatMessage] values.
//
// [New] is the Data API v3 client (requires a non-empty API key). The no-key
// HTML client lives in package nokey and also implements [Client].
// Callers must not fall back from v3 to nokey when the key is rejected;
// use [ErrInvalidAPIKey] to detect that case.
//
// [LookupLatestBroadcast] uses videos.list then search.list (live, then upcoming)
// so a previous video ID can be replaced without copying a new watch URL.
//
// The v3 HTTP client is wrapped with [quota.WrapClient] so this process can
// estimate Data API spend locally. That estimate does not call Google for
// remaining quota and does not add extra units.
package youtube
