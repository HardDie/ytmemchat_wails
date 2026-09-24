# Chat URL

Copy the **Chat** URL from Home into an OBS **Browser Source**. With the default port:

```text
http://127.0.0.1:8080/obs/chat
```

Do not use a trailing slash (`/obs/chat/` is a 404). Do not use the index URL (`http://127.0.0.1:8080/`) as an OBS source. Size the source to the chat area (a sidebar such as `400×800`, or full canvas `1920×1080`).

Add query parameters after `?`, joined with `&`. Restart or refresh the Browser Source after you change the URL.

## Parameters

| Parameter | Default | Values | What it changes |
|---|---|---|---|
| `transparent` | off (dark `#18181b` page background) | `1` or `true` | Makes the page background transparent so gameplay shows through. Message rows still have a light frosted panel. |
| `fontSize` | `16px` | A CSS size **with a unit**, for example `20px` or `1.25rem` | Message text size (author name and body). Avatars stay 32px. |
| `textColor` | `efeff1` (light gray) | Hex, with or without `#` | Message body color. Author names stay the accent purple; they are not this parameter. |
| `cap` | `100` | Positive integer, or `0` / `none` / `off` / `unlimited` | How many newest chat rows to keep. `0` and the named values mean no cap. |

`transparent=0`, empty values, and unknown names are ignored.

## Examples

Dark panel (default), larger type:

```text
http://127.0.0.1:8080/obs/chat?fontSize=22px
```

Over gameplay, white text:

```text
http://127.0.0.1:8080/obs/chat?transparent=1&textColor=ffffff
```

All three:

```text
http://127.0.0.1:8080/obs/chat?transparent=1&fontSize=20px&textColor=ffe08a
```

Keep every line (no cap):

```text
http://127.0.0.1:8080/obs/chat?cap=none
```

`textColor=fff` and `textColor=#ffffff` are both valid.

## OBS notes

- Put this URL in the Browser Source **URL** field. Leave **Custom CSS** empty unless you are adding extra rules; OBS already uses a transparent body in its default CSS.
- `transparent=1` is what drops ytmemchat’s own dark page fill. Without it, the source is an opaque rectangle even if OBS CSS is transparent.
- After editing the URL, click **Refresh cache of current page** on the source if the old look sticks.
- Test pane **Flush chat** sends `type: chat_flush` on the chat socket. On-screen lines are removed. Overlay alerts stay. See [Test](Test).
- The **Overlay** URL (`/obs/overlay`) has no query parameters. That page is already transparent; use **Interact** once so alert and TTS audio can autoplay.

Then continue with [Getting Started](Getting-Started) (Start chat) or [Configuration](Configuration) (port and overlay).
