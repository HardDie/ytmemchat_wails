# Setting up commands

A command is a chat trigger (for example `@jump`) that plays a clip on the OBS **overlay**. The desktop window does not show the clip. Chat text still appears on the chat page.

Alerts are off on first launch. Finish [Getting Started](Getting-Started) first if you have not pasted a stream ID yet. Field-by-field notes for the editor are on [Commands](Commands).

## What you need

- ytmemchat running (the local server starts with the app)
- A folder of media files on this machine
- [OBS Studio](https://obsproject.com/) on the **same computer** (the URLs use `127.0.0.1`)

## 1. Turn alerts on

1. Open **Configuration**.
2. Turn **Alerts** on.
3. Leave **Command token** at `@` unless you want a different single character.
4. Set **commands.yaml** to a path on this machine. **Browse** picks an existing `.yaml` or `.yml`. You can also type a new path; the file is created when you save commands later.
5. Set **Media folder** to the directory that holds the clips. **Browse** picks a folder.
6. Click **Save changes**.

![Configuration pane with Alerts on](https://raw.githubusercontent.com/HardDie/ytmemchat_wails/main/docs/screenshots/config.png)

If the YAML path is empty, matching is skipped and Start still works. A missing, unreadable, or invalid file **fails Start**.

## 2. Add commands

Open **Commands**. Each row is one trigger.

![Commands pane](https://raw.githubusercontent.com/HardDie/ytmemchat_wails/main/docs/screenshots/commands.png)

| Field | Meaning |
|---|---|
| **Name** | Word after the token. `jump` matches `@jump`. Unique ignoring case. |
| **File** | Path inside the media folder (`jump.webm` or `clips/jump.mp4`). The folder icon stores that relative path. |
| **Volume** | Playback gain. Blank means `1`. When set: `0` or more, at most two decimal digits. |
| **Scale** | Picture size. Same rules as volume. Blank means `1`. Audio-only files ignore scale. |

**Add command**, then **Save YAML**. Saving reloads the matcher. You do not need YouTube **Start**. **Reload** throws away unsaved edits.

The folder icon rejects files outside the media folder. Put clips in that folder (subfolders are fine).

Typical files:

| Kind | Extensions | On the overlay |
|---|---|---|
| Video | `webm`, `mp4`, `mov` | White frame, sound, random spot |
| Picture | `gif`, `png`, `jpg`, `jpeg`, `webp` | White frame, no sound, random spot |
| Sound | `mp3`, `wav`, `ogg`, `aac`, `flac` | Sound only (nothing drawn unless you debug) |

## 3. Add the overlay in OBS

Commands play on the **Overlay** URL from Home, not on the chat page.

```text
http://127.0.0.1:8080/obs/overlay
```

1. In OBS, add a **Browser Source**.
2. Paste that URL. No trailing slash (`/obs/overlay/` is a 404). Do not use `http://127.0.0.1:8080/`.
3. Set **Width** and **Height** to the area where clips may appear. Full canvas (`1920×1080`) is the usual choice. Clips are placed at random inside this rectangle, slightly rotated, up to about 400×400 pixels before **Scale**.
4. Leave **Custom CSS** empty.
5. Turn on **Control audio via OBS** so clip sound is in the mixer and on the stream.

The page background is already transparent. A video or picture gets a white rounded frame. A sound-only command draws nothing.

The app adds a `v` query so OBS loads the page for this build. Leave `v` off the URL you paste. Details shared with the chat page are on [Chat URL](Chat-URL#overlay-url).

Keep this source in scenes where alerts should play. **Shutdown source when not visible** (OBS default) stops the page when the source is hidden or you switch to a scene that does not include it. **Refresh browser when scene becomes active** reloads the page and cuts a clip that is still playing.

Chat is a second Browser Source (`/obs/chat`). It lists messages. It does not play command media. Query parameters for that page are on [Chat URL](Chat-URL).

If the port on Configuration is not `8080`, copy the Overlay URL from Home after you save.

## How a match works

Each new chat line is handled in this order:

1. The line is always sent to the chat page.
2. If **Alerts** is on, ytmemchat looks for the token. The word right after the **first** token is the command name. `please @jump now` is `jump`. Matching ignores case (`@Jump` is `jump`).
3. A name that is in `commands.yaml` plays that file on the overlay. **Text to speech is skipped** for that line.
4. A token with an unknown name is not an alert. TTS may still read the whole line. See [Text to speech](Text-to-speech).

Several commands can be on screen at once. Their audio can overlap. **Interrupt speech** does not stop a command clip.

On-screen time:

| Kind | How long it stays |
|---|---|
| Video | At least 5 seconds, or the file length if that is longer. Fade-out starts half a second before it goes. A short clip plays its sound once, then stays until that time. |
| Picture (including GIF) | About 6 seconds, then fade. |
| Sound | At least 5 seconds, or the file length if that is longer. No picture. |

A file the overlay cannot load disappears. Check that the name in **File** exists under the media folder.

## Try it without YouTube

On **Commands**, open the row menu and click **Test on overlay**. That sends the clip with no chat line. The overlay Browser Source must be open (Home shows **Listening**).

On **Test**, type `@jump` and click **Send**. The chat page shows author `test`, and the overlay plays the clip. **Start** is not required. See [Test](Test).
