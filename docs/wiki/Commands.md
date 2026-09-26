# Commands

The **Commands** pane edits the YAML file used for overlay alerts. It is not shown on stream.

Set the file path and media folder on [Configuration](Configuration) first (Alerts on), then return here. The full path, including the OBS Browser Source, is [Setting up commands](Setting-up-commands).

![Commands pane](https://raw.githubusercontent.com/HardDie/ytmemchat_wails/main/docs/screenshots/commands.png)

A chat line matches when it contains the command token from Configuration (default `@`) plus a **Name** from this list. Matching is case-insensitive (`@Jump` is `jump`). A match plays the media on the overlay and **skips TTS**. See [Configuration](Configuration#alerts-run-before-tts).

If the YAML path is empty, matching is skipped and Start still works. A missing, unreadable, or invalid file **fails Start**. Saving here reloads the matcher. YouTube Start is not required.

## Each command

| Field | Meaning |
|---|---|
| **Name** | Trigger after the token (`jump` for `@jump`). Must be unique ignoring case. |
| **File** | Media file relative to the media folder (`jump.webm`, or `clips/jump.mp4`). The folder icon picks a file inside that folder and stores the path without the folder prefix. |
| **Volume** | Overlay playback gain. Leave blank to omit it; playback uses `1`. When set: non-negative, at most two decimal digits. |
| **Scale** | Visual size multiplier. Same rules as volume; omitted means `1`. |

Leave volume and scale blank so those keys are not written. Playback still uses 1.

Example after **Save YAML**:

```yaml
commands:
  - name: "jump"
    file: "jump.webm"
    volume: 0.5
  - name: "dance"
    file: "dancing_cat.gif"
    scale: 1.2
```

**Add command** appends a row. **Save YAML** writes the file. **Reload** discards unsaved edits. Sort by Command or Filename only changes the on-screen order until you save.

**Play random** sends one command that has a file, chosen at random, using the volume and scale in the editor. Commands with a blank file are skipped. The per-row **Play** sends that one command. Both need Alerts on and the overlay listening.

The folder icon on File needs a media folder on Configuration. Paths outside that folder are rejected.
