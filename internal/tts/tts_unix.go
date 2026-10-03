//go:build !windows

package tts

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func synthesize(text, voiceName string) ([]byte, string, error) {
	tempFile, err := os.CreateTemp("", "tts_audio_*.wav")
	if err != nil {
		return nil, "", fmt.Errorf("failed to create temp file: %w", err)
	}
	tempFilePath := tempFile.Name()
	_ = tempFile.Close()
	defer os.Remove(tempFilePath)

	cmd, err := synthCommand(runtime.GOOS, text, voiceName, tempFilePath)
	if err != nil {
		return nil, "", err
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

// synthCommand builds the engine command that writes text to outPath.
// Text goes on stdin, never in argv.
// A chat line such as "-o/path" would otherwise be read as an option.
func synthCommand(goos, text, voiceName, outPath string) (*exec.Cmd, error) {
	var cmd *exec.Cmd
	switch goos {
	case "darwin":
		cmd = exec.Command("say",
			"-v", voiceName,
			"-o", outPath,
			"--data-format=LEF32@32000",
			"-f", "-")
	case "linux":
		cmd = exec.Command("espeak", "-v", voiceName, "-w", outPath, "--stdin")
	default:
		return nil, fmt.Errorf("unsupported OS for native TTS synthesis: %s", goos)
	}
	cmd.Stdin = strings.NewReader(text)
	return cmd, nil
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
