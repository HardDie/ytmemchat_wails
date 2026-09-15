package tts

import (
	"fmt"
	"runtime"
)

// VoiceInfo is metadata for one installed OS voice.
type VoiceInfo struct {
	// Name is the identifier passed to Speak and synthesis (for example "Alex").
	Name string
	// Language is a locale such as "en_US" or "en-US".
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
