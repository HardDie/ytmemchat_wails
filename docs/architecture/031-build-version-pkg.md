# 31. Build version is a shared package

* **Status:** Accepted
* **Date:** 2026-09-30
* **Authors:** @oleg

---

## Context

1. `AppVersion` lived in `bindings/sidebar`.
2. The value was a link-time string.
3. That string is an exact git tag, or the short commit hash.
4. Other programs need the same rule.

## Considered options

1. **Keep the stamp in `bindings/sidebar`** — this app only.
2. **Copy the git command into each app** — the rule drifts.
3. **`pkg/version` reads git and holds the stamp** — one implementation.

## Decision

Use option 3.

1. Package `pkg/version`.
2. `Describe` returns the exact tag of HEAD.
3. Untagged HEAD returns the 12-character commit hash.
4. `Build` is the link-time stamp (`-X …/pkg/version.Build`).
5. `String` returns `Build`. A blank stamp is `dev`.
6. `bindings/sidebar.AppVersion` calls `String`.
7. `make` and release CI stamp `Build` from `Describe`.

## Consequences

### Positive

* Another program imports `pkg/version` and stamps the same symbol.

### Negative and risks

* The `-X` path includes this module path.

### Neutral

* macOS and Windows bundle versions stay the last git tag.
* That value is still `info.productVersion`.
