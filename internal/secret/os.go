//go:build !nomain && !integration

package secret

import (
	"errors"

	"github.com/zalando/go-keyring"
)

type osVault struct{}

// OS is the platform keychain (macOS Keychain, Windows Credential Manager, libsecret).
func OS() Vault {
	return osVault{}
}

func (osVault) Available() bool {
	_, err := keyring.Get(Service, User)
	return err == nil || errors.Is(err, keyring.ErrNotFound)
}

func (osVault) Get() (string, error) {
	s, err := keyring.Get(Service, User)
	if errors.Is(err, keyring.ErrNotFound) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", err
	}
	return s, nil
}

func (osVault) Set(value string) error {
	return keyring.Set(Service, User, value)
}

func (osVault) Delete() error {
	err := keyring.Delete(Service, User)
	if errors.Is(err, keyring.ErrNotFound) {
		return nil
	}
	return err
}
