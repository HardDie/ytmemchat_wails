//go:build nomain

package main

func (a *App) setupWailsHooks() {}

func (a *App) PickCommandsFile() (string, error) { return "", nil }

func (a *App) PickMediaDirectory() (string, error) { return "", nil }

func (a *App) PickAlertMediaFile(_ string) (string, error) { return "", nil }
