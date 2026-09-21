// Package update is the Wails binding for the Update pane.
package update

import (
	"context"

	"github.com/HardDie/ytmemchat_wails/bindings/sidebar"
	intupdate "github.com/HardDie/ytmemchat_wails/internal/update"
)

// Update exposes GitHub release check, download, and quit-and-install.
type Update struct {
	app api
	c   *intupdate.Client
}

type api interface {
	DialogContext() context.Context
}

// New wraps the application core for the Update pane.
func New(app api) *Update {
	return &Update{app: app, c: intupdate.New(sidebar.New().AppVersion())}
}

// Check returns the latest GitHub release compared to this binary.
func (u *Update) Check() (intupdate.Status, error) {
	return u.c.Check()
}

// Download fetches and checksums the archive for this OS.
func (u *Update) Download() error {
	return u.c.Download()
}

// ApplyAndQuit starts a helper to replace this install, then quits the window.
func (u *Update) ApplyAndQuit() error {
	if err := u.c.Apply(); err != nil {
		return err
	}
	quit(u.app.DialogContext())
	return nil
}
