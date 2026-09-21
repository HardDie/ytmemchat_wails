// Package secret stores the YouTube API key in the OS keychain when possible.
// Tests use [Memory] or [Unavailable]. The real vault is compiled only without
// the nomain and integration tags.
package secret

import "fmt"

const (
	// Service is the keychain service name.
	Service = "ytmemchat"
	// User is the keychain account for the YouTube Data API key.
	User = "youtube-api-key"
)

var (
	// ErrNotFound is returned when the vault has no item.
	ErrNotFound = fmt.Errorf("secret: not found")
	// ErrUnavailable is returned when the OS keychain cannot be used.
	ErrUnavailable = fmt.Errorf("secret: keychain unavailable")
)

// Vault is a named secret (Get/Set/Delete). Implementations must not log values.
type Vault interface {
	// Available is true when the backend can store and retrieve items.
	Available() bool
	// Get returns the stored secret, [ErrNotFound], or [ErrUnavailable].
	Get() (string, error)
	// Set writes the secret.
	Set(value string) error
	// Delete removes the item. Missing items are not an error.
	Delete() error
}
