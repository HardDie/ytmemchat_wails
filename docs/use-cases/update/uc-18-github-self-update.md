# UC-18: Check GitHub for a newer app build

**Module:** `internal/update`  
**Status:** Implemented  
**Actors:** Operator (Update pane)  
**Goal:** See if a newer GitHub Release exists, verify its archive, then quit and replace this install  
**Preconditions:** Network to api.github.com. A tagged release with SHA256SUMS.txt.  

## Main scenario (happy path)

1. Operator clicks **Check**.
2. App reads `releases/latest` and compares the tag to `AppVersion`.
3. Operator clicks **Download**. SHA-256 must match `SHA256SUMS.txt`.
4. **Quit and install** starts a helper, quits the window, replaces the binary or `.app`, and launches it.

## Alternative scenarios and errors

* **Current is `dev` or a commit:** not treated as “behind”. Download is still allowed.
* **No archive for this OS:** Check succeeds; Download stays disabled.
* **Checksum mismatch:** Download fails; the file is discarded.
* **Install dir not writable:** show GitHub link; do not Apply.
* **No releases yet:** Check returns an HTTP error.

## Postconditions

* OBS HTTP stops when the app quits.
* Tests use httptest, not the real GitHub API.
