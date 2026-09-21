package secret

import "sync"

// Memory is an in-process vault for tests.
type Memory struct {
	mu        sync.Mutex
	value     string
	has       bool
	down      bool
	setErr    error
	deleteErr error
}

// Available reports whether this memory vault should act like a working keychain.
func (m *Memory) Available() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return !m.down
}

// Down marks the vault as unavailable (file fallback).
func (m *Memory) Down() {
	m.mu.Lock()
	m.down = true
	m.mu.Unlock()
}

// FailSet makes the next Set calls return err. Available stays true.
func (m *Memory) FailSet(err error) {
	m.mu.Lock()
	m.setErr = err
	m.mu.Unlock()
}

// Get returns the in-memory secret.
func (m *Memory) Get() (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.down {
		return "", ErrUnavailable
	}
	if !m.has {
		return "", ErrNotFound
	}
	return m.value, nil
}

// Set stores value.
func (m *Memory) Set(value string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.down {
		return ErrUnavailable
	}
	if m.setErr != nil {
		return m.setErr
	}
	m.value = value
	m.has = true
	return nil
}

// Delete clears the item.
func (m *Memory) Delete() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.deleteErr != nil {
		return m.deleteErr
	}
	m.value = ""
	m.has = false
	return nil
}
