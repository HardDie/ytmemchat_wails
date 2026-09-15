//go:build !windows

package tts

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

func speak(text, voiceName string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("say", "-v", voiceName, text)
	case "linux":
		cmd = exec.Command("espeak", "-v", voiceName, text)
	default:
		return fmt.Errorf("unsupported operating system for native TTS: %s", runtime.GOOS)
	}
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("TTS command failed on %s: %w", runtime.GOOS, err)
	}
	return nil
}

func synthesize(text, voiceName string) ([]byte, string, error) {
	tempFile, err := os.CreateTemp("", "tts_audio_*.wav")
	if err != nil {
		return nil, "", fmt.Errorf("failed to create temp file: %w", err)
	}
	tempFilePath := tempFile.Name()
	_ = tempFile.Close()
	defer os.Remove(tempFilePath)

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("say",
			"-v", voiceName,
			"-o", tempFilePath,
			"--data-format=LEF32@32000",
			text)
	case "linux":
		cmd = exec.Command("espeak", "-v", voiceName, "-w", tempFilePath, text)
	default:
		return nil, "", fmt.Errorf("unsupported OS for native TTS synthesis: %s", runtime.GOOS)
	}

	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, "", fmt.Errorf("TTS command failed on %s (tool %s): %w: %s", runtime.GOOS, cmd.Path, err, string(out))
	}

	audioData, err := os.ReadFile(tempFilePath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read synthesized audio file: %w", err)
	}
	if len(audioData) == 0 {
		return nil, "", fmt.Errorf("synthesized audio file is empty")
	}
	return audioData, "wav", nil
}

func getAvailableVoices() ([]VoiceInfo, error) {
	switch runtime.GOOS {
	case "darwin":
		cmd := exec.Command("say", "-v", "?")
		output, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("failed to execute 'say -v ?' on macOS: %w", err)
		}
		return parseSayVoiceList(string(output)), nil
	case "linux":
		cmd := exec.Command("espeak", "--voices")
		output, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("failed to execute 'espeak --voices'; is espeak installed? %w", err)
		}
		return parseEspeakVoiceList(string(output)), nil
	default:
		return nil, fmt.Errorf("voice listing not supported on %s", runtime.GOOS)
	}
}
