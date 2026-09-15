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
