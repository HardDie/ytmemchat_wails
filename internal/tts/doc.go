// Package tts synthesizes chat text with the host OS voice engine and
// emits WAV bytes for the OBS overlay to play.
//
// Drivers: macOS `say`, Linux `espeak`, Windows PowerShell System.Speech.
// This package does not start HTTP or import Wails. Overlay code maps [Speech]
// onto the WebSocket `tts` payload when `internal/obs` is ported.
package tts
