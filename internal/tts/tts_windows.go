//go:build windows

package tts

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func speak(text, voiceName string) error {
	text = strings.ReplaceAll(text, "'", "''")
	voiceName = strings.ReplaceAll(voiceName, "'", "''")
	powershellScript := fmt.Sprintf(`
		Add-Type -AssemblyName System.Speech;
		$synth = New-Object System.Speech.Synthesis.SpeechSynthesizer;
		$synth.SelectVoice('%s');
		$synth.Speak('%s');
	`, voiceName, text)
	cmd := exec.Command("powershell", "-NoProfile", "-Command", powershellScript)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("PowerShell execution failed (voice %s): %w. output: %s", voiceName, err, string(output))
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

	safeText := strings.ReplaceAll(text, "'", "''")
	safeVoice := strings.ReplaceAll(voiceName, "'", "''")
	powershellScript := fmt.Sprintf(`
		Add-Type -AssemblyName System.Speech;
		$synth = New-Object System.Speech.Synthesis.SpeechSynthesizer;
		$synth.SelectVoice('%s');
		$synth.SetOutputToWaveFile('%s');
		$synth.Speak('%s');
		$synth.Dispose();
	`, safeVoice, tempFilePath, safeText)
	cmd := exec.Command("powershell", "-NoProfile", "-Command", powershellScript)
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, "", fmt.Errorf("powershell audio write failed: %w: %s", err, string(out))
	}
	audioData, err := os.ReadFile(tempFilePath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read temp audio file: %w", err)
	}
	if len(audioData) == 0 {
		return nil, "", fmt.Errorf("synthesized audio file is empty")
	}
	return audioData, "wav", nil
}

type psVoiceInfo struct {
	Name        string
	Culture     string
	Gender      string
	Description string
}

func getAvailableVoices() ([]VoiceInfo, error) {
	powershellCommand := `
		Add-Type -AssemblyName System.Speech;
		$synth = New-Object System.Speech.Synthesis.SpeechSynthesizer;
		$synth.GetInstalledVoices() |
		Select-Object -ExpandProperty VoiceInfo |
		Select-Object Name, Culture, Gender, Description |
		ConvertTo-Json -Compress
	`
	cmd := exec.Command("powershell", "-NoProfile", "-Command", powershellCommand)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute PowerShell for voice list: %w", err)
	}
	return parseWindowsVoiceJSON(output)
}

func parseWindowsVoiceJSON(output []byte) ([]VoiceInfo, error) {
	var psVoices []psVoiceInfo
	if err := json.Unmarshal(output, &psVoices); err != nil {
		var one psVoiceInfo
		if errOne := json.Unmarshal(output, &one); errOne != nil {
			return nil, fmt.Errorf("failed to parse JSON output from PowerShell: %w. raw: %s", err, string(output))
		}
		psVoices = []psVoiceInfo{one}
	}
	voices := make([]VoiceInfo, 0, len(psVoices))
	for _, pv := range psVoices {
		voices = append(voices, VoiceInfo{
			Name:     pv.Name,
			Language: pv.Culture,
			Gender:   normalizeGender(pv.Gender),
			Details:  pv.Description,
		})
	}
	return voices, nil
}
