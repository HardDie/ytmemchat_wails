// Package quota estimates YouTube Data API v3 spend from HTTP that this
// process already makes. It does not call Google Cloud or the Data API to
// read remaining quota, so the estimate never adds extra units.
//
// [WrapClient] is the only integration the v3 client needs. [Default] is the
// process-wide tracker (Start and Find latest share it). Counters reset at
// midnight Pacific Time. They are not Google’s remaining quota: other tools
// on the same Cloud project, and spend from a previous run today, are
// invisible. The no-key HTML client must not be wrapped.
package quota
