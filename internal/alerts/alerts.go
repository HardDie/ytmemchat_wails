package alerts

import (
	"fmt"
	"strings"
)

// Clip is one media alert for the OBS overlay.
type Clip struct {
	// Filename is a file inside the configured media directory.
	Filename string
	// Volume is overlay playback gain from 0 to 1 (default 1).
	Volume float64
	// Scale is the visual size multiplier (default 1).
	Scale float64
}

// Config holds YAML path, matching settings, and the sink for [Clip] values.
type Config struct {
	// Token is the command prefix (for example "@").
	Token string
	// MediaPath is the directory of alert media files (used later by obs).
	MediaPath string
	// CommandsFilePath is the YAML file of command names to files.
	CommandsFilePath string
	// Out receives [Clip] values. New requires a non-nil channel.
	Out chan<- Clip
}

// Alerts matches chat lines to YAML commands and sends [Clip] values to Out.
type Alerts struct {
	token    string
	mediaDir string
	commands map[string]command
	out      chan<- Clip
}

// New loads CommandsFilePath and rejects duplicate names (case-insensitive).
func New(cfg Config) (*Alerts, error) {
	if cfg.Out == nil {
		return nil, fmt.Errorf("alerts: nil Out channel")
	}
	if cfg.Token == "" {
		return nil, fmt.Errorf("alerts: empty token")
	}
	parsed, err := parseCommands(cfg.CommandsFilePath)
	if err != nil {
		return nil, fmt.Errorf("parseCommands(): %w", err)
	}
	a := &Alerts{
		token:    cfg.Token,
		mediaDir: cfg.MediaPath,
		commands: make(map[string]command),
		out:      cfg.Out,
	}
	for _, it := range parsed.Commands {
		key := strings.ToLower(it.Name)
		if _, ok := a.commands[key]; ok {
			return nil, fmt.Errorf("duplicate command: %s", it.Name)
		}
		a.commands[key] = it
	}
	return a, nil
}

// MediaPath returns the configured media directory.
func (a *Alerts) MediaPath() string {
	return a.mediaDir
}

// Alert looks for a command after the token. On a match it sends a [Clip]
// and returns true so the caller can skip TTS.
func (a *Alerts) Alert(msg string) bool {
	name := findToken(a.token, msg)
	if name == "" {
		return false
	}
	cmd, ok := a.commands[strings.ToLower(name)]
	if !ok {
		return false
	}
	a.out <- Clip{
		Filename: cmd.File,
		Volume:   valueOr(cmd.Volume, 1),
		Scale:    valueOr(cmd.Scale, 1),
	}
	return true
}

func valueOr(v *float64, def float64) float64 {
	if v == nil {
		return def
	}
	return *v
}

func findToken(token, str string) string {
	_, str, _ = strings.Cut(str, token)
	if str == "" {
		return ""
	}
	return strings.Split(str, " ")[0]
}
