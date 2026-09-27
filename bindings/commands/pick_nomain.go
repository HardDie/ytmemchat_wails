//go:build nomain

package commands

func (c *Commands) PickAlertMediaFile(_ string) (string, error) { return "", nil }

func (c *Commands) ExportAlertCommands() (string, error) { return "", nil }

func (c *Commands) ImportAlertCommands() (string, error) { return "", nil }
