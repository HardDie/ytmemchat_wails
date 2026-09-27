//go:build windows

package tts

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

// createNoWindow is the Win32 CREATE_NO_WINDOW flag.
// A GUI app otherwise opens a console for powershell.exe.
const createNoWindow = 0x08000000

func hiddenCommand(name string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: createNoWindow,
	}
	return cmd
}

func speak(text, voiceName string) error {
	text = strings.ReplaceAll(text, "'", "''")
	voiceName = strings.ReplaceAll(voiceName, "'", "''")
	powershellScript := fmt.Sprintf(`
		Add-Type -AssemblyName System.Speech;
		$synth = New-Object System.Speech.Synthesis.SpeechSynthesizer;
		%s
		$synth.Speak('%s');
	`, voiceSelectLine(voiceName), text)
	cmd := hiddenCommand("powershell", "-NoProfile", "-Command", powershellScript)
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
		%s
		$synth.SetOutputToWaveFile('%s');
		$synth.Speak('%s');
		$synth.Dispose();
	`, voiceSelectLine(safeVoice), tempFilePath, safeText)
	cmd := hiddenCommand("powershell", "-NoProfile", "-Command", powershellScript)
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

func voiceSelectLine(voiceName string) string {
	if strings.TrimSpace(voiceName) == "" {
		return ""
	}
	return fmt.Sprintf("$synth.SelectVoice('%s');", voiceName)
}

func getAvailableVoices() ([]VoiceInfo, error) {
	// Culture is a CultureInfo object and Gender is an enum.
	// Select-Object would emit those as JSON objects and numbers, which do not parse.
	powershellCommand := `
		Add-Type -AssemblyName System.Speech
		$synth = New-Object System.Speech.Synthesis.SpeechSynthesizer
		$voices = @(
			$synth.GetInstalledVoices() | ForEach-Object {
				$info = $_.VoiceInfo
				[PSCustomObject]@{
					Name = [string]$info.Name
					Culture = [string]$info.Culture.Name
					Gender = [string]$info.Gender
					Description = [string]$info.Description
				}
			}
		)
		if ($voices.Count -eq 0) { '[]' } else { $voices | ConvertTo-Json -Compress }
	`
	cmd := hiddenCommand("powershell", "-NoProfile", "-Command", powershellCommand)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute PowerShell for voice list: %w", err)
	}
	return parseWindowsVoiceJSON(output)
}
