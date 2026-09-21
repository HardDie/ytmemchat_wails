# Configuration

Every setting on the **Configuration** pane is stored on this machine. Chat and overlay URLs stay on **Home**. Click **Save changes** to persist. **Start** on Home also saves the form first.

Alerts and TTS are off until you turn them on. Chat still works with both off; see [Getting Started](Getting-Started).

![Configuration pane](https://raw.githubusercontent.com/HardDie/ytmemchat_wails/main/docs/screenshots/config.png)

## Alerts run before TTS

Each new chat line is handled in this order:

1. The line is always sent to the **chat** page (`/obs/chat`).
2. If **Alerts** is on and the text matches a command, the **overlay** plays that media. TTS is skipped for that line.
3. If there was no alert match and **Text to speech** is on, the overlay speaks the line.

So a viewer who types `@jump` sees the message in chat and gets the jump clip. The message is **not** read aloud. A line with no matching command can be spoken.

Alerts and TTS both play on the overlay page, not in the operator window. Add an OBS **Browser Source** for:

```text
http://127.0.0.1:8080/obs/overlay
```

Click **Interact** on that source once and use the page’s enable-audio control so the browser is allowed to autoplay sound. Chat Browser Source query parameters are on [Chat URL](Chat-URL).

---

## Connection

### Stream / video ID

Required to **Start**. This is the YouTube video ID (`v=` in the watch URL, or the last segment of a `/live/…` URL), not the full URL. The video must be live with chat enabled when you Start.

**Find latest** (same row) replaces the ID with that channel’s current live stream, or the next upcoming one if nothing is live. VODs are skipped. It needs a [YouTube API key](YouTube-API-key) and a video ID already in the field (any recent video from the channel is enough). It does not save until you click Save or Start.

### YouTube API key

Optional. Empty uses the public live chat page. See [YouTube API key](YouTube-API-key) for what this token is, how to create one, and the default quota. With a key, **Home** shows spent units today (this app’s count; `liveChatMessages.list` is 5 units). When the OS keychain works, the key is stored there; otherwise it stays in the settings file.

### HTTP port

Local listen port for OBS pages. Default `8080`. You can enter `8080` or `:8080`. Saving a different port restarts the listener; copy the new URLs from Home.

---

## Alerts

Off on first launch. When off, command matching is skipped and every line can go to TTS (if TTS is on).

When on, extra fields appear:

### Command token

A single character that marks a command in chat. Default `@`. Example: token `@` and command name `jump` match `@jump` (also `please @jump now`). The word right after the first token in the message is the command name. Matching is case-insensitive (`@Jump` is `jump`).

If the token is present but the name is not in `commands.yaml`, that is not an alert. TTS may still speak the line.

### commands.yaml

Path to the YAML file of command names. Edit it in the **Commands** pane after this path is saved. See [Commands](Commands).

### Media folder

Directory of those files. They are served at `/obs/media/`. The overlay loads `/obs/media/<file>`. If this folder is empty, media URLs 404.

Typical files: sound (`mp3`, `wav`), video (`webm`, `mp4`), or a GIF. Keep files inside this tree; the picker rejects paths outside it.

---

## Text to speech

Off on first launch. When on, lines that **did not** match an alert are synthesized and played on the overlay.

### Voice

The operating-system voice. The list comes from this machine. Leave empty for the OS default.

| OS | Engine |
|---|---|
| macOS | built-in `say` |
| Windows | PowerShell `System.Speech` |
| Linux | `espeak` (install it if the list is empty) |

Emoji is stripped before speech. A line that is only emoji is not spoken. Speech plays in OBS, not on the ytmemchat host speakers.

If Alerts is off, every chat line can be spoken (including `@jump` as ordinary text). Turn alerts on first if you want commands to play media instead of being read.

---

## Interrupt shortcut

Stops **current and queued TTS** on the overlay. It does not stop a meme alert that is already playing.

On by default. Default combination is `Ctrl+Shift+I` (Control, not Command, so it does not steal macOS app shortcuts).

### Key combination

Click **Set**, then press a shortcut that includes at least one of Ctrl, Cmd, Alt, or Shift. Escape cancels recording. **Default** restores `Ctrl+Shift+I`. Save to apply.

The shortcut is global: it works while OBS is fullscreen and this window is in the background. Linux needs X11 (not a pure Wayland session). If registration fails, Configuration and Home show the error.

Home also has **Interrupt speech** for the same action. `POST /api/interrupt` does too when the HTTP API is on.

---

## HTTP API

Off on first launch. When on, the local server exposes operator endpoints (not for OBS):

```bash
curl -X POST http://127.0.0.1:8080/api/webhook \
     -H 'Content-Type: application/json' \
     -d '{"message": "@jump"}'

curl -X POST http://127.0.0.1:8080/api/interrupt
```

`/api/webhook` injects a fake chat line through the same alerts-then-TTS path (author `webhook`). YouTube Start is not required. The **Test** pane Send button is the same path (author `test`) and does **not** need this toggle.

---

## Settings file

The pane footer shows the JSON path. Typical locations:

| OS | Path |
|---|---|
| macOS | `~/Library/Application Support/ytmemchat/config.json` |
| Windows | `%AppData%\ytmemchat\config.json` |
| Linux | `~/.config/ytmemchat/config.json` |

Do not put secrets in the git repo. The file mode is `0600` because it can hold the API key.
