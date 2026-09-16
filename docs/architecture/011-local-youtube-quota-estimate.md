# 11. Local YouTube Data API quota estimate

* **Status:** Accepted
* **Date:** 2026-09-16
* **Authors:** @oleg

---

## Context

A default Google Cloud project gets **10,000 units per day** for most YouTube Data API v3 methods, plus a separate **100 `search.list` calls per day**. Quota resets at midnight Pacific Time. The Data API does not return remaining units, and Cloud Monitoring / Cloud Quotas need extra credentials (not the YouTube API key). Operators still want a sense of spend so a long Start session does not surprise them.

Counting by adding extra YouTube or Cloud API calls would itself spend quota (or require OAuth). Sprinkling `Record()` next to every `.Do()` in the v3 client is easy to miss when a new method is added.

## Considered options

1. **Query Google Cloud for remaining quota** — accurate, but needs project ID and OAuth/service account; a YouTube-restricted API key cannot call it.
2. **Probe YouTube with extra list calls** — does not return remaining units; every attempt costs at least 1 unit.
3. **Explicit `Record` at each v3 `.Do()` site** — clear, but a new call path can skip the counter.
4. **HTTP transport in `internal/youtube/quota`** — one wrap in the v3 client constructor; every request that gets an HTTP response is classified from the URL. No extra Google traffic.

## Decision

Use option 4.

* The Data API v3 client wraps its `http.Client` (`quota.WrapClient`). Because `option.WithHTTPClient` skips `option.WithAPIKey`, the v3 constructor then attaches the key on that client (`key` query param). The no-key HTML client is not wrapped.
* Count **after an HTTP response**, including 4xx/5xx (Google bills invalid requests at least 1 unit). Do **not** count transport errors with no response (the request may never have reached Google).
* Use Google’s billed costs for methods this app actually calls: `videos.list` costs **1**; `liveChatMessages.list` costs **5** (Cloud Console usage; the public calculator table currently lists 1). `search.list` goes to the separate search bucket (1 each, default 100/day). Unknown `/youtube/v3/…` paths count **1** in the default bucket.
* Counters are **per Pacific day**, shared across Start and Find latest via `quota.Default`, and persisted as `quota.json` next to `config.json` so a restart does not drop today’s total. They are not Google’s remaining quota: other tools on the same Cloud project are invisible.

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
