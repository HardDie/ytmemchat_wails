package main

import (
	"context"
	"fmt"
)

// App is the Wails bindings façade (settings and start/stop later).
type App struct {
	ctx context.Context
}

// NewApp returns the bound application struct.
func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Greet is the template demo binding. It will be replaced by settings methods.
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}
