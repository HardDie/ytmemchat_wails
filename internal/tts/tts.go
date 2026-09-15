package tts

import (
	"fmt"
	"regexp"
	"runtime"
	"strings"
)

// Speech is one synthesized utterance for overlay playback.
type Speech struct {
	// WAV is raw audio produced by the OS engine (WAVE format).
	WAV []byte
	// Volume is playback gain in the overlay, from 0 to 1.
	Volume float64
}

// Config holds voice settings and the sink for synthesized audio.
type Config struct {
	// VoiceName is the OS voice identifier (for example "Milena" on macOS).
	VoiceName string
	// Volume is overlay playback gain from 0 to 1. A nil value means 1.
	Volume *float64
	// Out receives [Speech] values. SynthesizeAudio requires a non-nil channel;
	// the caller must receive or buffer it.
	Out chan<- Speech
}

// TTS synthesizes chat lines and sends [Speech] to [Config.Out].
type TTS struct {
	voiceName string
	volume    float64
	out       chan<- Speech
}

// New returns a TTS engine. Volume defaults to 1 when Config.Volume is nil.
func New(cfg Config) *TTS {
	return &TTS{
		voiceName: cfg.VoiceName,
		volume:    valueOr(cfg.Volume, 1),
		out:       cfg.Out,
	}
}

// Speak plays text on the local default output (not the OBS overlay).
func Speak(text, voiceName string) error {
	return speak(text, voiceName)
}

// SynthesizeToBuffer renders text with voiceName and returns WAV bytes and the
// format name (always "wav" for current drivers).
func SynthesizeToBuffer(text, voiceName string) ([]byte, string, error) {
	audioData, format, err := synthesize(text, voiceName)
	if err != nil {
		return nil, "", fmt.Errorf("failed to synthesize audio on %s: %w", runtime.GOOS, err)
	}
	return audioData, format, nil
}

// SynthesizeAudio strips emoji from text, synthesizes WAV audio, and sends a
// [Speech] value on Out. It returns an error if text is empty after cleaning,
// Out is nil, or the OS engine fails.
func (t *TTS) SynthesizeAudio(text string) error {
	if t.out == nil {
		return fmt.Errorf("tts: nil Out channel")
	}
	cleaned := removeEmojis(text)
	if cleaned == "" {
		return fmt.Errorf("tts: no speakable text")
	}
	audioData, _, err := synthesize(cleaned, t.voiceName)
	if err != nil {
		return fmt.Errorf("synthesis failed: %w", err)
	}
	t.out <- Speech{WAV: audioData, Volume: t.volume}
	return nil
}

func valueOr[T any](v *T, def T) T {
	if v == nil {
		return def
	}
	return *v
}

// \p{So} Symbol, Other (most emoji)
// \p{Sk} Symbol, Modifier (skin tones)
// \p{Mn} Mark, Nonspacing
// \p{Cs} Surrogate
// \p{Sm} Symbol, Math
var emojiRegex = regexp.MustCompile(`[\p{So}\p{Sk}\p{Mn}\p{Cs}\p{Sm}]`)

func removeEmojis(input string) string {
	content := emojiRegex.ReplaceAllString(input, "")
	return strings.TrimSpace(strings.Join(strings.Fields(content), " "))
}
