//go:build integration && linux

package tts

import "testing"

func TestGetAvailableVoices_system(t *testing.T) {
	skipIfMissing(t, "espeak")
	testGetAvailableVoices(t)
}

func TestSynthesizeToBuffer_system(t *testing.T) {
	skipIfMissing(t, "espeak")
	testSynthesizeToBuffer(t)
}

func TestSynthesizeAudio_sendsSpeech(t *testing.T) {
	skipIfMissing(t, "espeak")
	testSynthesizeAudio(t)
}

// A line starting with -w was read by espeak as its output file option.
func TestSynthesizeToBuffer_dashTextIsNotAnOption(t *testing.T) {
	skipIfMissing(t, "espeak")
	testSynthesizeHostile(t, func(pwn string) string { return "-w" + pwn })
}
