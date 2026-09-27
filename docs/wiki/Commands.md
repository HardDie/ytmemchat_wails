# Commands

The **Commands** pane edits the YAML file used for overlay alerts. It is not shown on stream.

Set the file path and media folder on [Configuration](Configuration) first (Alerts on), then return here. The full path, including the OBS Browser Source, is [Setting up commands](Setting-up-commands).

![Commands pane](https://raw.githubusercontent.com/HardDie/ytmemchat_wails/main/docs/screenshots/commands.png)

A chat line matches when it contains the command token from Configuration (default `@`) plus a **Name** from this list. Matching is case-insensitive (`@Jump` is `jump`). A match plays the media on the overlay and **skips TTS**. See [Configuration](Configuration#alerts-run-before-tts).

If the YAML path is empty, matching is skipped and Start still works. A missing default `commands.yaml` also skips matching. A missing custom file fails Start. A file that cannot be read or parsed fails Start too. Saving here applies the file. YouTube Start is not required.

## Each command

| Field | Meaning |
|---|---|
| **Name** | Trigger after the token (`jump` for `@jump`). Must be unique ignoring case. |
| **File** | Media file relative to the media folder (`jump.webm`, or `clips/jump.mp4`). The folder icon is the only way to set it. The stored path has the folder prefix removed. |
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

The footer buttons are icons. Hover one to see its name.

**Add command**, **Reload**, **Save YAML**, **Export**, and **Import** sit on the left. **Play random** and **Copy commands** sit on the right.

**Add command** appends a row. **Save YAML** writes the file. **Reload** discards unsaved edits. Sort by Command or Filename only changes the on-screen order until you save.

**Search** sits next to those buttons. It hides rows whose Name or File does not contain the typed text. Matching ignores case. The cross inside the field clears it. Hidden rows stay in the file.

**Copy commands** copies the names on screen. Each line starts with the token from Configuration. A blank name is skipped.

**Export** packs the saved YAML and the media files those commands use.

`commands.yaml` is at the root of the zip. A custom path still uses that name.

Each used file keeps its path relative to the media folder. Files no command uses stay out.

Save YAML first. Export reads the file on disk. A missing file cancels the export.

**Import** asks for a zip, then for a folder.

The zip must have `commands.yaml` at its root. Every file in the zip is unpacked into that folder.

The media folder in Configuration becomes that folder. A custom commands path is cleared, so the unpacked `commands.yaml` is the one in use.

Cancel either dialog and the config stays as it was.

**Play random** sends one command that has a file, chosen at random, using the volume and scale in the editor. Commands with a blank file are skipped. The play icon on a row sends that one command. Both need Alerts on and the overlay listening.

The trash icon asks before that row is removed. **Save YAML** writes the change.

The folder icon on File needs a media folder on Configuration. Paths outside that folder are rejected. If Name is empty, the filename without its extension becomes the Name. A Name already typed stays as it is.
