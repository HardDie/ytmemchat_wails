package tts

import (
	"fmt"
	"runtime"
	"strings"

	"golang.org/x/text/language"
	"golang.org/x/text/language/display"
)

// VoiceInfo is metadata for one installed OS voice.
type VoiceInfo struct {
	// Name is the identifier passed to Speak and synthesis (for example "Alex").
	Name string
	// Language is one or more locale codes, comma-separated (for example "en_US" or "en-gb, en-uk, en").
	Language string
	// Gender is "Male", "Female", or empty when unknown.
	Gender string
	// Details is the raw system description.
	Details string
}

// GetAvailableVoices lists TTS voices installed on this machine.
func GetAvailableVoices() ([]VoiceInfo, error) {
	voices, err := getAvailableVoices()
	if err != nil {
		return nil, fmt.Errorf("failed to get voice info on %s: %w", runtime.GOOS, err)
	}
	return voices, nil
}

// FormatLanguages turns locale codes into English names with the code in
// parentheses, for example "Russian (ru_RU)". Unknown tags are left as-is.
func FormatLanguages(raw string) string {
	var parts []string
	for _, piece := range strings.Split(raw, ",") {
		code := strings.TrimSpace(piece)
		if code == "" {
			continue
		}
		parts = append(parts, formatLanguage(code))
	}
	return strings.Join(parts, ", ")
}

func formatLanguage(code string) string {
	tag, err := language.Parse(strings.ReplaceAll(code, "_", "-"))
	if err != nil {
		return code
	}
	name := display.English.Tags().Name(tag)
	if name == "" || strings.EqualFold(name, code) {
		return code
	}
	return name + " (" + code + ")"
}

func normalizeGender(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "m", "male":
		return "Male"
	case "f", "female":
		return "Female"
	default:
		return strings.TrimSpace(raw)
	}
}
