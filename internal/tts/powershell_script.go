//go:build windows

package tts

// Env var names read by [psSynthScript].
const (
	psEnvText  = "YTMEMCHAT_TTS_TEXT"
	psEnvVoice = "YTMEMCHAT_TTS_VOICE"
	psEnvOut   = "YTMEMCHAT_TTS_OUT"
)

// psSynthScript writes one utterance to a WAV file with System.Speech.
// It is a constant: text, voice, and path come from env vars.
// Chat text is never parsed as PowerShell, so no quote escaping is needed.
// PowerShell also treats ‘ ’ ‚ ‛ as quotes, which escaping ' alone misses.
const psSynthScript = `
Add-Type -AssemblyName System.Speech
$synth = New-Object System.Speech.Synthesis.SpeechSynthesizer
if ($env:` + psEnvVoice + `) { $synth.SelectVoice($env:` + psEnvVoice + `) }
$synth.SetOutputToWaveFile($env:` + psEnvOut + `)
$synth.Speak($env:` + psEnvText + `)
$synth.Dispose()
`

// psSynthEnv returns the env entries [psSynthScript] reads.
func psSynthEnv(text, voiceName, outPath string) []string {
	return []string{
		psEnvText + "=" + text,
		psEnvVoice + "=" + voiceName,
		psEnvOut + "=" + outPath,
	}
}
