//go:build nomain || integration

package secret

// OS is a no-op vault in unit and integration tests (no display / keychain).
func OS() Vault {
	return Unavailable{}
}
