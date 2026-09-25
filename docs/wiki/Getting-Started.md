# Getting Started

This is the simplest way to run ytmemchat: paste a live stream ID, leave everything else alone, and watch chat. You do not need a YouTube API key. Alerts and text-to-speech are off on first launch.

The desktop window is only the operator console. Chat is a local web page that OBS (or a browser) loads.

## What you need

- ytmemchat (a [GitHub Release](https://github.com/HardDie/ytmemchat_wails/releases) for your OS, or a binary you built). On macOS the release is unsigned: [Running on macOS](Running-on-macOS).
- A YouTube video that is **live right now**, with live chat enabled
- Optional: [OBS Studio](https://obsproject.com/), if you want that chat on stream

## 1. Open the app

Unpack the release and start ytmemchat. The window opens on **Home**. The local chat server starts with the app, even before you click Start. macOS may block the first launch; see [Running on macOS](Running-on-macOS).

![Home pane: Start/Stop, interrupt, and OBS Browser Source URLs](https://raw.githubusercontent.com/HardDie/ytmemchat_wails/main/docs/screenshots/home.png)

## 2. Paste the stream ID

1. Open **Configuration**.
2. In **Stream / video ID**, paste only the video ID, not the full URL.

| Watch URL | Stream ID |
|---|---|
| `https://www.youtube.com/watch?v=AbCdEf12345` | `AbCdEf12345` |
| `https://www.youtube.com/live/AbCdEf12345` | `AbCdEf12345` |

3. Leave **YouTube API key** empty.
4. Leave **Alerts** and **Text to speech** off.
5. Leave **HTTP port** at `8080` unless that port is already in use.
6. Click **Save changes**.

![Configuration pane: Connection card with Stream / video ID](https://raw.githubusercontent.com/HardDie/ytmemchat_wails/main/docs/screenshots/config.png)

The screenshot above is a populated demo. On first launch **Alerts** is Off, so you will not see `commands.yaml` until you turn alerts on later.

Skip **Find latest** for this setup. It needs an API key.

## 3. Open the chat page

On **Home**, copy the **Chat** URL. With the default port it is:

```text
http://127.0.0.1:8080/obs/chat
```

Open that URL in a browser to preview, or add an OBS **Browser Source** pointed at it. Size the source to the area you want for chat (a sidebar such as `400×800`, or full canvas `1920×1080`).

To drop the dark background when the source sits over gameplay, use:

```text
http://127.0.0.1:8080/obs/chat?transparent=1
```

Font size, text color, and other chat query parameters are on [Chat URL](Chat-URL).

You can ignore the **Overlay** URL for this guide. Do not use the index URL (`http://127.0.0.1:8080/`) as an OBS source.

## 4. Start live chat

On **Home**, click **Start**. The badge should move from **Connecting** to **Connected**. The detail line reads **No API key**.

New messages appear on the chat page as viewers send them. ytmemchat does not replay history: lines that were already in chat before you clicked Start stay off the overlay.

## Which messages you will see

Without an API key (token), ytmemchat reads the **public** live chat page — the same feed an anonymous viewer gets in the browser.

YouTube does not put every line on that page. If YouTube marks a message as spam or otherwise “bad”, it never reaches ytmemchat, so it will not appear in our chat overlay. That is YouTube filtering the public feed, not the overlay dropping a message.

Studio chat (and a [YouTube API key](YouTube-API-key)) can show a fuller feed. For this getting-started path, expect some messages that you see as the channel owner to be missing here. A new Google key has a small daily limit; it is still enough for **Find latest**. See [YouTube API key](YouTube-API-key#default-quota-why-a-long-stream-may-run-out).

## If Start fails

| What you see | What to check |
|---|---|
| **Stream ID is required to Start** | Save a video ID on Configuration first. |
| **This video is not a current live stream with chat**, or `failed to extract chat continuation` | The video must be **live now** with chat on. Upcoming premieres, ended streams, and VODs do not work. |
| Chat page is empty after **Connected** | Only **new** messages are shown. Send a test line from another account, or wait for a viewer. |
| Chat shows **WEB SOCKET DISCONNECTED** | ytmemchat is not running, or the Browser Source URL / port is wrong. |
| A message you saw in YouTube Studio never appears | See [Which messages you will see](#which-messages-you-will-see). |

## Next

When you want meme alerts, follow [Setting up commands](Setting-up-commands). Speech needs the overlay source: [Text to speech](Text-to-speech). A [YouTube API key](YouTube-API-key) and the interrupt shortcut are on [Configuration](Configuration). Edit alert clips on [Commands](Commands). Use [Test](Test) to send a fake line or flush OBS chat without YouTube Start.
