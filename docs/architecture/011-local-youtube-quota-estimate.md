# 11. Local YouTube Data API quota estimate

* **Status:** Accepted
* **Date:** 2026-09-16
* **Authors:** @oleg

---

## Context

A default Google Cloud project has daily Data API budgets.

1. **10,000 units per day** for most YouTube Data API v3 methods.
2. A separate **100 `search.list` calls per day**.
3. Quota resets at midnight Pacific Time.
4. The Data API does not return remaining units.
5. Cloud Monitoring / Cloud Quotas need extra credentials (not the YouTube API key).
6. Operators still want a sense of spend so a long Start session does not surprise them.

Counting by adding extra YouTube or Cloud API calls is a bad fit.

1. Extra calls spend quota (or require OAuth).
2. Sprinkling `Record()` next to every `.Do()` in the v3 client is easy to miss when a new method is added.

## Considered options

1. **Query Google Cloud for remaining quota**
   1. Accurate, but needs project ID and OAuth/service account.
   2. A YouTube-restricted API key cannot call it.
2. **Probe YouTube with extra list calls** — does not return remaining units; every attempt costs at least 1 unit.
3. **Explicit `Record` at each v3 `.Do()` site** — clear, but a new call path can skip the counter.
4. **HTTP transport in `internal/youtube/quota`**
   1. One wrap in the v3 client constructor.
   2. Every request that gets an HTTP response is classified from the URL.
   3. No extra Google traffic.

## Decision

Use option 4.

1. The Data API v3 client wraps its `http.Client` (`quota.WrapClient`).
2. `option.WithHTTPClient` skips `option.WithAPIKey`.
3. The v3 constructor then attaches the key on that client (`key` query param).
4. The no-key HTML client is not wrapped.
5. Count **after an HTTP response**, including 4xx/5xx (Google bills invalid requests at least 1 unit).
6. Do **not** count transport errors with no response (the request may never have reached Google).
7. Billed costs this app uses:
   1. `videos.list` **1**
   2. `liveChatMessages.list` **5** (Cloud Console usage; the public calculator table currently lists 1)
8. `search.list` goes to the separate search bucket (1 each, default 100/day).
9. Unknown `/youtube/v3/…` paths count **1** in the default bucket.
10. Counters are **per Pacific day**, shared across Start and Find latest via `quota.Default`.
11. Persist as `quota.json` next to `config.json` so a restart does not drop today’s total.
12. They are not Google’s remaining quota. Other tools on the same Cloud project are invisible.

## Consequences

### Positive

* No extra quota spend to “check” quota.
* Lookup, chat connect, and polling are counted without touching each `.Do()` site.
* Tests can inject a private `Tracker` so they do not share `Default`.

### Negative and risks

* The number can be **below** real project spend (other apps on the same Cloud project).
* The default 10,000 / 100 limits are Google’s documented defaults, not a project’s approved quota if the owner requested more.
* A response from a proxy that Google never billed would over-count (unusual).

### Neutral

* Home reads `quota.Default.Snapshot()` via `GetRunStatus` (polled). Display only; it must not add Data API calls.
