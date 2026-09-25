# Text to speech

Text to speech reads a chat line aloud **in OBS**, on the overlay page. The ytmemchat window does not speak, and the app does not play the line on the computer’s speakers by itself.

TTS is off on first launch. A line that matched a [command](Setting-up-commands) is not spoken.

## What you need

- ytmemchat running
- [OBS Studio](https://obsproject.com/) on the **same computer**, with a Browser Source pointed at the **overlay** URL
- A voice on this operating system (below)

The chat Browser Source is not enough. Chat only draws messages. Speech is a separate page.

## 1. Add the overlay in OBS

On **Home**, copy **Overlay**. With the default port:

```text
http://127.0.0.1:8080/obs/overlay
```

1. In OBS, add a **Browser Source** (or reuse the one from [Setting up commands](Setting-up-commands#3-add-the-overlay-in-obs)).
2. Paste that URL. No trailing slash. Do not use `http://127.0.0.1:8080/`.
3. Set **Width** and **Height** large enough that the source stays in the scene. Full canvas (`1920×1080`) matches command clips. Speech itself draws nothing, but a hidden or zero-size source does not stay loaded.
4. Leave **Custom CSS** empty.
5. Turn on **Control audio via OBS**. That puts overlay audio in the OBS mixer. Raise that source if the stream is silent. Mute it and speech stops on the stream.

The page is already transparent. Nothing is drawn while someone is speaking.

The app adds `v` so OBS loads this build’s page. Do not type `v` yourself.

**Shutdown source when not visible** unloads the page when this source is hidden or the current scene does not include it. Speech for that moment is dropped. Put the overlay on every scene that should talk, or leave the source visible.

**Refresh browser when scene becomes active** reloads the page and cuts the line that is playing.

If Home says **WEB SOCKET DISCONNECTED** on a preview, or the overlay status banner says that, ytmemchat is not running or the URL / port does not match Configuration.

One overlay source covers both TTS and command clips. Chat stays a second source: [Chat URL](Chat-URL).

## 2. Turn TTS on

1. Open **Configuration**.
2. Turn **Text to speech** on.
3. Pick a **Voice**, or leave **System default**.
4. Click **Save changes**.

![Configuration pane](https://raw.githubusercontent.com/HardDie/ytmemchat_wails/main/docs/screenshots/config.png)

| OS | Engine |
|---|---|
| macOS | built-in `say` |
| Windows | PowerShell `System.Speech` |
| Linux | `espeak` |

On Linux, install `espeak` if the voice list is empty (`espeak` on the package manager for your distro). A missing voice or engine fails that line only. Chat and commands still run. The failure is logged; the overlay stays quiet for that message.

There is no TTS volume field. Playback gain is `1`. Use the OBS mixer on this Browser Source to make it louder or quieter.

## What gets spoken

For each new chat line:

1. The text is always sent to the chat page.
2. If **Alerts** is on and the line matches a command, the overlay plays that clip and **does not** speak. See [Setting up commands](Setting-up-commands#how-a-match-works).
3. Otherwise, if TTS is on, the **message text** is spoken. The author name is not added.

Emoji is removed first. A line that is only emoji is not spoken. Extra spaces are collapsed.

If Alerts is off, `@jump` is ordinary text and can be read aloud. Turn alerts on when that token should play a clip instead.

Speech is one line at a time. The next line waits until the current one finishes. A busy chat can lag behind. Command clips are not in this queue: a meme can play while someone is still being read, and a meme does not wait for speech.

## Stop a line

**Interrupt speech** on Home stops the utterance that is playing. Lines already waiting still play afterward. It does not stop a command clip.

The same action is the global shortcut on Configuration (default `Ctrl+Shift+I`, Control, not Command). It works while OBS is fullscreen and this window is in the background. Linux needs X11, not a pure Wayland session. If the shortcut fails to register, Configuration and Home show the error.

`POST /api/interrupt` does the same thing when **HTTP API** is on. The Home button does not need that toggle.

Closing ytmemchat stops overlay audio. The source then shows that the socket is down until you open the app again.

## Try it

On **Test**, type `hello` and click **Send**. Author on the chat page is `test`. The overlay speaks `hello` if TTS is on and the overlay source is running. YouTube **Start** is not required. See [Test](Test).

A real stream speaks only **new** lines after you click **Start**. History from before Start is not read.
