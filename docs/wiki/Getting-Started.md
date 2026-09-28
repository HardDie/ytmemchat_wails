# Getting Started

This is the simplest way to run ytmemchat. Paste a live stream ID, leave everything else alone, and watch chat.

You do not need a YouTube API key. Alerts and text-to-speech are off on first launch.

The desktop window only controls the app. OBS or a browser loads the pages.

The **Chat** page lists messages. Use that page for this guide. The **Overlay** page plays clips and speech. Ignore that URL until you want alerts or speech.

## What you need

- ytmemchat from a [GitHub Release](https://github.com/HardDie/ytmemchat_wails/releases) for your OS, or a binary you built
- A YouTube video that is **live right now**, with live chat enabled
- Optional: [OBS Studio](https://obsproject.com/), if you want that chat on stream

On macOS the release is unsigned. See [Running on macOS](Running-on-macOS).

On Linux, run the binary or install it for your user. See [Installing on Linux](Installing-on-Linux).

## 1. Open the app

Unpack the release and start ytmemchat. The window opens on **Home**. The local chat server starts with the app, even before you click Start.

macOS may block the first launch. See [Running on macOS](Running-on-macOS).

![Home pane: Start/Stop, interrupt, and OBS Browser Source URLs](https://raw.githubusercontent.com/HardDie/ytmemchat_wails/main/docs/screenshots/home.png)

## 2. Paste the stream ID

1. Open **Configuration**.
2. In **Stream / video ID**, paste only the video ID, not the full URL.

| Watch URL | Stream ID |
|---|---|
| `https://www.youtube.com/watch?v=AbCdEf12345` | `AbCdEf12345` |
| `https://www.youtube.com/live/AbCdEf12345` | `AbCdEf12345` |

Leave the API key empty. Leave alerts and text-to-speech off. Leave **HTTP port** at `8080` unless that port is already in use. Then click **Save changes**.

![Configuration pane: Connection card with Stream / video ID](https://raw.githubusercontent.com/HardDie/ytmemchat_wails/main/docs/screenshots/config.png)

The screenshot above is a populated demo. On first launch **Alerts** is Off, so you will not see `commands.yaml` until you turn alerts on later.

Skip **Find latest** for this setup. It needs an API key.

## 3. Open the chat page

On **Home**, copy the **Chat** URL. With the default port it is:

```text
http://127.0.0.1:8080/obs/chat
```

Open that URL in a browser to preview, or add an OBS **Browser Source**. Size the source to the chat area, such as a sidebar `400×800` or full canvas `1920×1080`.

To drop the dark background when the source sits over gameplay, use:

```text
http://127.0.0.1:8080/obs/chat?transparent=1
```

Font size, text color, and other chat query parameters are on [Chat URL](Chat-URL). Ignore the **Overlay** URL for this guide.

The index URL (`http://127.0.0.1:8080/`) lists the chat and overlay pages. Do not use it as an OBS source.

## 4. Start live chat

On **Home**, click **Start**. The badge should move from **Connecting** to **Connected**. The detail line reads **No API key**.

New messages appear on the chat page as viewers send them. ytmemchat does not replay history. Lines already in chat before you clicked Start stay off that page.

**Stop** ends the YouTube read. The badge returns to **Stopped**. Messages already on the chat page stay. The chat and overlay pages stay up.

Home uses these words:

- **Connecting**, then **Connected** — YouTube chat is coming in. **Stopped** is the badge before Start and after Stop.
- **No API key** — the line under the badge while Connected without a key. With a key it says **YouTube Data API v3**.
- **OBS ready** and **Listening** — the local pages are up. **OBS offline** means they are not.

## Which messages you will see

Without an API key (token), ytmemchat reads the **public** live chat page. That is the same feed an anonymous viewer gets in the browser.

YouTube does not put every line on that page. If YouTube marks a message as spam or otherwise “bad”, it never reaches ytmemchat. It will not appear on the chat page. YouTube filtered that line before ytmemchat could see it.

Studio chat, or a [YouTube API key](YouTube-API-key), can show a fuller feed. On this path, some messages you see as the channel owner can be missing here.

## If Start fails

| What you see | What to check |
|---|---|
| **Stream ID is required to Start** | Save a video ID on Configuration first. |
| **This video is not a current live stream with chat**, or `failed to extract chat continuation` | The video must be **live now** with chat on. Upcoming premieres, ended streams, and VODs do not work. |
| Chat page is empty after **Connected** | Only **new** messages are shown. Send a test line from another account, or wait for a viewer. |
| Chat shows **WEB SOCKET DISCONNECTED** | ytmemchat is not running, or the Browser Source URL / port is wrong. |
| A message you saw in YouTube Studio never appears | See [Which messages you will see](#which-messages-you-will-see). |

## Next

Meme alerts are [Setting up commands](Setting-up-commands). Speech is [Text to speech](Text-to-speech).

Later pages:

- Alert clip editor: [Commands](Commands)
- YouTube API key: [YouTube API key](YouTube-API-key)
- A fake line, or flush OBS chat: [Test](Test)
