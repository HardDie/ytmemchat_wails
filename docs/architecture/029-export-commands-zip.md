# 29. Export commands as a zip

* **Status:** Accepted
* **Date:** 2026-09-27
* **Authors:** @oleg

---

## Context

1. Operators want to share a command collection.
2. The YAML path may be custom.
3. The media folder holds used and unused files.
4. Import reads that same zip into a chosen folder.

## Considered options

1. **Copy the media folder** — simple; unused files travel with the share.
2. **Zip the YAML plus used files** — the share matches the collection.
3. **YAML only** — no clips; the other machine cannot play commands.

## Decision

Use option 2.

1. **Export** on the Commands pane opens a save dialog.
2. Cancel leaves disk unchanged.
3. The zip root holds `commands.yaml`.
   1. The bytes are the saved file.
   2. A custom path still uses the name `commands.yaml`.
4. Each command `file` is stored at that relative path.
   1. The path is relative to the media folder.
   2. The same file is stored once.
5. Files not named by a command are omitted.
6. A path outside the media folder is rejected.
7. A missing file fails the export.
   1. No zip is left behind.
8. Unsaved editor rows are not packed.
9. The button sits with Add, Reload, and Save YAML.
10. `pkg/archive` writes the zip.
    1. It does not know commands or the media folder.
    2. Commands passes the saved YAML, the name `commands.yaml`, and the used files.
11. **Import** checks for `commands.yaml` at the zip root, then asks for a folder.
    1. Files unpack into that folder at their archive paths.
    2. The saved media folder becomes that path.
    3. A custom commands path is cleared.
    4. Cancel leaves the config unchanged.

## Consequences

### Positive

* A share contains only the clips the collection uses.
* Unzipping into a media folder restores the default layout.

### Negative and risks

* Unsaved edits are not in the zip.
* A command file named `commands.yaml` cannot be packed.

### Neutral

* Import uses the same zip layout.
