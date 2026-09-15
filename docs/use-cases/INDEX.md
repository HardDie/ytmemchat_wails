# Use cases

Scenarios for **ported** modules. Do not add `uc-*.md` until that module lives in this repository. This index may list planned IDs so numbering stays stable.

**Status:** Planned (not in this repo yet) · In progress · Ported

Copy [\_TEMPLATE.md](_TEMPLATE.md) when a module is ported.

| ID | Name | Module | Status | File |
|---|---|---|---|---|
| UC-01 | Start chat without an API key | `internal/youtube/nokey` | Ported | [youtube/uc-01-start-chat-without-api-key.md](youtube/uc-01-start-chat-without-api-key.md) |
| UC-02 | Start chat with a valid API key | `internal/youtube` | Ported | [youtube/uc-02-start-chat-with-api-key.md](youtube/uc-02-start-chat-with-api-key.md) |
| UC-03 | Invalid API key does not fall back | `internal/youtube` | Ported | [youtube/uc-03-invalid-api-key.md](youtube/uc-03-invalid-api-key.md) |
| UC-04 | Save and reload settings | `internal/config` | Ported | [config/uc-04-save-settings.md](config/uc-04-save-settings.md) |
| UC-05 | Show live chat in OBS | `internal/obs` | Ported | [obs/uc-05-obs-live-chat.md](obs/uc-05-obs-live-chat.md) |
| UC-06 | Play alert / TTS on overlay | `internal/obs` | Ported | [obs/uc-06-obs-overlay.md](obs/uc-06-obs-overlay.md) |
| UC-07 | Match an alert command | `internal/alerts` | Ported | [alerts/uc-07-alert-command.md](alerts/uc-07-alert-command.md) |
| UC-08 | Speak a non-alert message | `internal/tts` | Ported | [tts/uc-08-speak-message.md](tts/uc-08-speak-message.md) |
| UC-09 | Inject a test chat message | `internal/obs` (`/api/webhook`) | Ported | [obs/uc-09-inject-webhook.md](obs/uc-09-inject-webhook.md) |
| UC-10 | Interrupt TTS | `internal/obs` (`/api/interrupt`) | Ported | [obs/uc-10-interrupt-tts.md](obs/uc-10-interrupt-tts.md) |
| UC-11 | Start and stop the pipeline | `app.go` | Ported | [app/uc-11-start-stop.md](app/uc-11-start-stop.md) |
| UC-12 | Find latest live or upcoming stream | `internal/youtube` | Ported | [youtube/uc-12-find-latest-stream.md](youtube/uc-12-find-latest-stream.md) |
| UC-13 | Edit commands.yaml from the window | `internal/alerts` | Ported | [alerts/uc-13-edit-commands-yaml.md](alerts/uc-13-edit-commands-yaml.md) |

## Layout (after porting)

```text
docs/use-cases/
├── INDEX.md
├── _TEMPLATE.md
├── config/
│   └── uc-04-save-settings.md
├── youtube/
│   ├── uc-01-start-chat-without-api-key.md
│   ├── uc-02-start-chat-with-api-key.md
│   ├── uc-03-invalid-api-key.md
│   └── uc-12-find-latest-stream.md
├── obs/
│   ├── uc-05-obs-live-chat.md
│   ├── uc-06-obs-overlay.md
│   ├── uc-09-inject-webhook.md
│   └── uc-10-interrupt-tts.md
├── alerts/
│   ├── uc-07-alert-command.md
│   └── uc-13-edit-commands-yaml.md
├── tts/
│   └── uc-08-speak-message.md
└── app/
    └── uc-11-start-stop.md
```
