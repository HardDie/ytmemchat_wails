package secret

// Unavailable is a vault that is never usable. Get and Set fail.
type Unavailable struct{}

// Available is always false.
func (Unavailable) Available() bool { return false }

// Get returns [ErrUnavailable].
func (Unavailable) Get() (string, error) { return "", ErrUnavailable }

// Set returns [ErrUnavailable].
func (Unavailable) Set(string) error { return ErrUnavailable }

// Delete is a no-op.
func (Unavailable) Delete() error { return nil }
