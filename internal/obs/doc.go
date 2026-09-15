// Package obs serves OBS Browser Source pages and operator HTTP APIs.
// It does not import youtube, alerts, tts, or Wails. The HTTP/WebSocket
// listener is meant to run for the life of the Wails process; Start/Stop
// in app.go only runs the YouTube iterator and calls PublishChat /
// PublishOverlay. app.go maps chat lines and Clip/Speech values onto
// [ChatEvent] and [OverlayEvent].
package obs
