//go:build integration && windows

package tts

import "testing"

func TestGetAvailableVoices_system(t *testing.T) {
	requireTool(t, "powershell")
	testGetAvailableVoices(t)
}

func TestSynthesizeToBuffer_system(t *testing.T) {
	requireTool(t, "powershell")
	testSynthesizeToBuffer(t)
}

func TestSynthesizeAudio_sendsSpeech(t *testing.T) {
	requireTool(t, "powershell")
	testSynthesizeAudio(t)
}

// PowerShell treats ’ as a quote, so this line once closed the Speak string
// and ran New-Item. It must now be spoken as plain text.
func TestSynthesizeToBuffer_quoteTextIsNotCode(t *testing.T) {
	requireTool(t, "powershell")
	testSynthesizeHostile(t, func(pwn string) string {
		return "x’); New-Item -ItemType File -Path ’" + pwn + "’ | Out-Null; $null=(’"
	})
}
