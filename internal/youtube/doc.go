// Package youtube normalizes YouTube Live Chat into [ChatMessage] values.
//
// [New] is the Data API v3 client (requires a non-empty API key). The no-key
// HTML client lives in package nokey and also implements [Client].
// Callers must not fall back from v3 to nokey when the key is rejected;
// use [ErrInvalidAPIKey] to detect that case.
package youtube
