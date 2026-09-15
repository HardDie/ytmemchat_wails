package config

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

var (
	// ErrStreamIDRequired is returned by [Settings.CanStart] when stream ID is empty.
	ErrStreamIDRequired = errors.New("config: stream ID is required")
	// ErrInvalidPort is returned when Server.Port is not a valid TCP port.
	ErrInvalidPort = errors.New("config: invalid listen port")
	// ErrAlertToken is returned when alerts are enabled and Token is not one character.
	ErrAlertToken = errors.New("config: alert token must be a single character")
)

// CanStart reports whether the pipeline may start. Stream ID is required;
// API key is optional. Errors never include the API key.
func (s Settings) CanStart() error {
	s = s.TrimSpace()
	if s.Youtube.StreamID == "" {
		return ErrStreamIDRequired
	}
	return s.Validate()
}

// Validate checks fields that must be well-formed even when saving an
// incomplete form (empty stream ID is allowed). Errors never include the API key.
func (s Settings) Validate() error {
	s = s.TrimSpace()
	if _, err := ParsePort(s.Server.Port); err != nil {
		return err
	}
	if s.Alerts.Enabled && utf8.RuneCountInString(s.Alerts.Token) != 1 {
		return ErrAlertToken
	}
	return nil
}

// ParsePort accepts "8080" or ":8080" and returns the port number.
func ParsePort(port string) (int, error) {
	port = strings.TrimSpace(port)
	port = strings.TrimPrefix(port, ":")
	if port == "" {
		return 0, fmt.Errorf("%w: empty", ErrInvalidPort)
	}
	n, err := strconv.Atoi(port)
	if err != nil || n < 1 || n > 65535 {
		return 0, fmt.Errorf("%w", ErrInvalidPort)
	}
	return n, nil
}

// ListenAddr returns an address suitable for net.Listen ("host:port" form with empty host).
func (s Server) ListenAddr() string {
	p := strings.TrimSpace(s.Port)
	if strings.HasPrefix(p, ":") {
		return p
	}
	return ":" + p
}
