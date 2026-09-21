// Package hotkey parses operator keyboard chords and registers a global
// interrupt shortcut with the OS.
//
// Chord parsing is CGO-free. OS registration uses golang.design/x/hotkey
// in files tagged `!nomain && !integration`. Unit and integration tests
// compile a no-op binder so Ubuntu CI does not need a display.
package hotkey
