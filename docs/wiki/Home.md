# Home

ytmemchat reads YouTube live chat and drives an OBS overlay: on-stream messages, meme alerts, and text-to-speech.

## Start

* [Getting Started](Getting-Started) — paste a stream ID and watch chat (no API key)
* [Running on macOS](Running-on-macOS) — open the unsigned `.app` (Gatekeeper)

## OBS

* [Chat URL](Chat-URL) — OBS Browser Source query parameters (`transparent`, `fontSize`, `textColor`, `cap`). `v` is added by the app.

## Setup

* [Configuration](Configuration) — every settings field; alerts run before TTS
* [Setting up commands](Setting-up-commands) — alerts on, media folder, overlay Browser Source
  * [Commands](Commands) — edit `commands.yaml` (name, file, volume, scale)
* [Text to speech](Text-to-speech) — voice, and the overlay source OBS must load
* [YouTube API key](YouTube-API-key) — what the Google token is, how to create one, and the default quota

## Tools

* [Test](Test) — send a fake chat line, or flush OBS chat
* [Update](Update) — check GitHub Releases and replace this install

## Contributing

* [Contributing](Contributing) — make targets, tests, and package docs
