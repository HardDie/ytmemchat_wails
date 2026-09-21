//go:build nomain

package configuration

func (c *Configuration) PickCommandsFile() (string, error) { return "", nil }

func (c *Configuration) PickMediaDirectory() (string, error) { return "", nil }
