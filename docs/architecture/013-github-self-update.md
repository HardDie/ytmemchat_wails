# 13. GitHub Releases for in-app update

* **Status:** Accepted
* **Date:** 2026-09-21
* **Authors:** @oleg

---

## Context

1. Releases are tagged archives ([008](008-github-actions-test-and-release.md)).
2. Operators need to replace a running binary or macOS `.app`.
3. Binaries are unsigned.

## Considered options

1. **Download only from the GitHub Releases page** — no in-app path.
2. **Sparkle / WinSparkle** — extra native stack; signing assumed.
3. **Check `releases/latest`, SHA-256, helper after quit** — same archives as CI.

## Decision

Use option 3.

1. Package `internal/update`. Binding `bindings/update`.
2. Operator **Check**, **Download**, **Quit and install**.
3. No check on startup.
4. Compare the latest tag to `AppVersion`.
5. `dev` or a commit is never “behind”. Download is still allowed.
6. SHA-256 must match `SHA256SUMS.txt`. Mismatch discards the file.
7. A helper replaces the binary or `.app` after quit, then starts it.
8. If the install dir is not writable: GitHub link. Do not Apply.
9. Tests use httptest, not api.github.com.

## Consequences

### Positive

* Same assets as tag CI.
* Checksum before replace.

### Negative and risks

* Unsigned macOS may need Gatekeeper again.
* The helper must wait for this process to exit.

### Neutral

* Operator steps: wiki [Update](../wiki/Update.md).
* Scenario: [UC-18](../use-cases/update/uc-18-github-self-update.md).
