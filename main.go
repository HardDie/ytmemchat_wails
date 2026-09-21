//go:build !nomain

// Package main is the Wails desktop entrypoint: native window plus pane
// bindings. Domain logic lives in internal/. OBS HTTP starts in OnStartup.
// Start/Stop runs the YouTube iterator into the chat overlay.
package main

import (
	"embed"

	"github.com/HardDie/ytmemchat_wails/bindings/commands"
	"github.com/HardDie/ytmemchat_wails/bindings/configuration"
	"github.com/HardDie/ytmemchat_wails/bindings/home"
	"github.com/HardDie/ytmemchat_wails/bindings/sidebar"
	"github.com/HardDie/ytmemchat_wails/bindings/test"
	"github.com/HardDie/ytmemchat_wails/bindings/update"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:     "ytmemchat",
		Width:     760,
		Height:    680,
		MinWidth:  640,
		MinHeight: 520,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 18, G: 21, B: 26, A: 1},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Bind: []interface{}{
			sidebar.New(),
			home.New(app),
			configuration.New(app),
			commands.New(app),
			test.New(app),
			update.New(app),
		},
	})
	if err != nil {
		println("Error:", err.Error())
	}
}
