# 30. GitHub self-update is a shared package

* **Status:** Accepted
* **Date:** 2026-09-27
* **Authors:** @oleg

---

## Context

1. In-app update is [013](013-github-self-update.md).
2. It lived in `internal/update`.
3. Owner, repo, and binary name were constants.
4. Similar apps need the same Check, Download, and Apply flow.

## Considered options

1. **Keep `internal/update`** — this app only.
2. **Copy the package into each app** — the copies drift.
3. **`pkg/update` takes owner, repo, and binary name** — one implementation.

## Decision

Use option 3.

1. Package `pkg/update`.
2. `Config` carries owner, repo, binary name, and the current version.
3. The package does not mention ytmemchat.
4. `bindings/update` passes HardDie, ytmemchat_wails, and ytmemchat.
5. Archive suffixes stay those from tag CI.
   1. `darwin-universal.zip`
   2. `windows-amd64.zip`
   3. `linux-amd64.tar.gz`
   4. `linux-arm64.tar.gz`
6. The archive must contain `Name`, `Name.exe`, or `Name.app`.

## Consequences

### Positive

* Another app calls `New` with its own identity.

### Negative and risks

* A different archive layout will not match.

### Neutral

* Operator steps stay on wiki [Update](../wiki/Update.md).
* [013](013-github-self-update.md) still describes Check, checksum, and the helper.
