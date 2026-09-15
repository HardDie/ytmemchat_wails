# 10. OS-level interrupt hotkey

* **Status:** Accepted
* **Date:** 2026-09-15
* **Authors:** @oleg

---

## Context

OBS is usually fullscreen, so a Wails window keydown handler cannot stop TTS. Operators need a shortcut that works while another app is focused.

## Considered options

1. **Window-only shortcut** — fails when OBS has focus.
2. **HTTP `/api/interrupt` only** — already exists; too slow in a panic.
3. **OS global hotkey** (Carbon / `RegisterHotKey` / X11) registered from package main, chord stored in config.json.

## Decision

Use option 3. Default chord is `Ctrl+Shift+I` (Control, not Command, so it does not steal macOS app shortcuts). The operator can record a different combination in Config. Registration is CGO in package main (`!nomain`); `internal/hotkey` only parses chords so tests stay CGO-free. Pin `golang.design/x/hotkey` v0.4.1 because later versions use a macOS event tap that requires Accessibility permission. Linux is X11 only.

## Consequences

### Positive

* Interrupt works with OBS fullscreen and ytmemchat in the background.
* No Accessibility prompt on macOS.

### Negative and risks

* The chord can collide with another app’s global shortcut; show the register error in Config/Home.
* Pure Wayland sessions cannot grab keys this way.
