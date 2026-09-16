// Package quota estimates YouTube Data API v3 spend from HTTP that this
// process already makes. It does not call Google Cloud or the Data API to
// read remaining quota, so the estimate never adds extra units.
//
// [WrapClient] is the only integration the v3 client needs. [Default] is the
// process-wide tracker (Start and Find latest share it). liveChatMessages.list
// is billed at [LiveChatMessagesListCost] (5) to match Cloud Console usage.
// Counters reset at midnight Pacific Time and are persisted next to
// config.json as [FileName]. They are not Google’s remaining quota: other
// tools on the same Cloud project are invisible. The no-key HTML client
// must not be wrapped.
package quota
