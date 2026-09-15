//go:build integration && windows

package tts

import "testing"

func TestGetAvailableVoices_system(t *testing.T) {
	skipIfMissing(t, "powershell")
	testGetAvailableVoices(t)
}

func TestSynthesizeToBuffer_system(t *testing.T) {
	skipIfMissing(t, "powershell")
	testSynthesizeToBuffer(t)
}

func TestSynthesizeAudio_sendsSpeech(t *testing.T) {
	skipIfMissing(t, "powershell")
	testSynthesizeAudio(t)
}
