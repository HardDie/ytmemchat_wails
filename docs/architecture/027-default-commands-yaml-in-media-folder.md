# 27. Default commands.yaml lives in the media folder

* **Status:** Accepted
* **Date:** 2026-09-25
* **Authors:** @oleg

---

## Context

1. Alerts ask for two paths.
   1. `commands.yaml`
   2. Media folder
2. The operator must create or pick the YAML file.
3. The Commands pane stays closed until that path is saved.
4. Save YAML already creates a missing file.
5. A set path with a missing file fails Start.

## Considered options

1. **Keep both path fields** — matches today; the YAML path stays a manual step.
2. **Default file inside the media folder** — one path for the common case; custom path stays available.
3. **Create the YAML when the media folder is saved** — file exists before any command; extra empty file on disk.

## Decision

Use option 2.

1. Checkbox off is the default.
2. The path field stays hidden, same as Voice when Text to speech is off.
3. Effective path is `<media folder>/commands.yaml`.
4. An empty media folder means no commands path.
5. Checkbox **Use custom commands.yaml path** shows the existing path field and Browse.
6. The checkbox is not stored in `config.json`.
   1. On load it is on when the saved path is set and is not the default file.
   2. On save, off stores an empty path.
   3. On save, on stores the custom path.
7. Unchecking and saving leaves the old custom file on disk.
8. A media-folder change in default mode follows the new folder.
   1. The old YAML is not moved.
9. A legacy path that is not the default file opens with the checkbox on.
10. The file is created on Save YAML, not on Add.
    1. A new row has no name or file yet.
    2. Save rejects that row.
11. Default mode and a missing file: Start works.
    1. The matcher is skipped.
    2. Same as an empty path today.
12. Invalid YAML still fails Start.
13. Custom mode and a missing file still fails Start.
14. Checking the box and saving the default path stays default mode.
    1. That path is the default file.

## Consequences

### Positive

* First-time setup is the media folder only.
* The first saved command creates `commands.yaml` there.
* An existing custom path is not replaced.

### Negative and risks

* Moving the media folder in default mode points at a different YAML file.
* A custom path that matches the default file does not keep the checkbox on.

### Neutral

* Commands pane with no effective path asks for a media folder or a custom path.
* An existing default path with no file yet shows an empty command list.
