//go:build !nomain

package main

import "github.com/wailsapp/wails/v2/pkg/runtime"

func (a *App) setupWailsHooks() {
	a.emit = func(name string, data any) {
		if a.ctx == nil {
			return
		}
		runtime.EventsEmit(a.ctx, name, data)
	}
}
