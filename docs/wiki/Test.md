# Test

The **Test** pane injects one chat line into the same alerts-then-TTS path as live chat. **Start** is not required. The HTTP API toggle is not required.

![Test pane: Send a fake chat line and Flush chat](https://raw.githubusercontent.com/HardDie/ytmemchat_wails/main/docs/screenshots/test.png)

Open the [Chat](Chat-URL) and [overlay](Configuration#alerts-run-before-tts) Browser Sources first so you can see the result.

## Send

1. Type a message (`@jump` or `hello`).
2. Click **Send**, or press Enter.

The author on the chat page is `test`. A token plus a name from [Commands](Commands) plays that clip on the overlay and skips TTS. Any other text can be spoken if TTS is on.

Send is disabled while OBS HTTP is down (**OBS offline**). The listener starts with the app; you do not need YouTube Start.

`POST /api/webhook` is the same path with author `webhook`. That route needs the HTTP API toggle on [Configuration](Configuration#http-api). Send does not.

## Flush chat

**Flush chat** removes every on-screen line on the chat page (`type: chat_flush`). Overlay alerts and TTS keep playing. New lines after a flush still appear.

Flush is also disabled while OBS HTTP is down.
