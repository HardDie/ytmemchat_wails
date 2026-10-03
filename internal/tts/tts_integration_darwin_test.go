//go:build integration && darwin

package tts

import "testing"

func TestGetAvailableVoices_system(t *testing.T) {
	skipIfMissing(t, "say")
	testGetAvailableVoices(t)
}

func TestSynthesizeToBuffer_system(t *testing.T) {
	skipIfMissing(t, "say")
	testSynthesizeToBuffer(t)
}

func TestSynthesizeAudio_sendsSpeech(t *testing.T) {
	skipIfMissing(t, "say")
	testSynthesizeAudio(t)
}

// A line starting with -o was read by say as its output file option.
func TestSynthesizeToBuffer_dashTextIsNotAnOption(t *testing.T) {
	skipIfMissing(t, "say")
	testSynthesizeHostile(t, func(pwn string) string { return "-o" + pwn })
}
