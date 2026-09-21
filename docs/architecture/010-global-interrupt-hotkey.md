# 10. OS-level interrupt hotkey

* **Status:** Accepted
* **Date:** 2026-09-15
* **Authors:** @oleg

---

## Context

1. OBS is usually fullscreen.
2. A Wails window keydown handler cannot stop TTS.
3. Operators need a shortcut that works while another app is focused.

## Considered options

1. **Window-only shortcut** — fails when OBS has focus.
2. **HTTP `/api/interrupt` only** — already exists; too slow in a panic.
3. **OS global hotkey** (Carbon / `RegisterHotKey` / X11) in `internal/hotkey`, chord stored in config.json.

## Decision

Use option 3.

1. Default chord is `Ctrl+Shift+I` (Control, not Command, so it does not steal macOS app shortcuts).
2. The operator can record a different combination in Config.
3. Registration is CGO in `internal/hotkey` (`!nomain && !integration`).
4. Chord parse stays CGO-free in the same package.
5. Tests compile a no-op binder (`nomain` or `integration`).
6. Pin `golang.design/x/hotkey` v0.4.1.
7. Later versions use a macOS event tap that requires Accessibility permission.
8. Linux is X11 only.

## Consequences

### Positive

* Interrupt works with OBS fullscreen and ytmemchat in the background.
* No Accessibility prompt on macOS.

### Negative and risks

* The chord can collide with another app’s global shortcut. Show the register error in Config/Home.
* Pure Wayland sessions cannot grab keys this way.
* Linux `init()` panics without an X11 display, so headless `wails build` (GitHub Actions) must use Xvfb.
