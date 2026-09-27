//go:build windows

package tts

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"unicode/utf16"
)

// psVoiceInfo is one voice from Windows PowerShell.
// Culture and Gender are strings in the script, but Windows PowerShell 5.1
// ConvertTo-Json also emits Culture as an object and Gender as a number.
type psVoiceInfo struct {
	Name        string
	Culture     string
	Gender      string
	Description string
}

func (v *psVoiceInfo) UnmarshalJSON(data []byte) error {
	var raw struct {
		Name        string          `json:"Name"`
		Culture     json.RawMessage `json:"Culture"`
		Gender      json.RawMessage `json:"Gender"`
		Description string          `json:"Description"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	v.Name = raw.Name
	v.Description = raw.Description
	v.Culture = jsonStringOrName(raw.Culture)
	v.Gender = jsonGender(raw.Gender)
	return nil
}

func parseWindowsVoiceJSON(output []byte) ([]VoiceInfo, error) {
	output = bytes.TrimSpace(decodePowerShellStdout(output))
	if len(output) == 0 || bytes.Equal(output, []byte("null")) {
		return []VoiceInfo{}, nil
	}
	var psVoices []psVoiceInfo
	if err := json.Unmarshal(output, &psVoices); err != nil {
		var one psVoiceInfo
		if errOne := json.Unmarshal(output, &one); errOne != nil {
			raw := output
			if len(raw) > 512 {
				raw = raw[:512]
			}
			return nil, fmt.Errorf("failed to parse JSON output from PowerShell: %w. raw: %s", err, string(raw))
		}
		psVoices = []psVoiceInfo{one}
	}
	voices := make([]VoiceInfo, 0, len(psVoices))
	for _, pv := range psVoices {
		if pv.Name == "" {
			continue
		}
		voices = append(voices, VoiceInfo{
			Name:     pv.Name,
			Language: pv.Culture,
			Gender:   normalizeGender(pv.Gender),
			Details:  pv.Description,
		})
	}
	return voices, nil
}

// decodePowerShellStdout turns Windows PowerShell pipe output into UTF-8.
// powershell.exe writes UTF-16 LE when stdout is not a console.
func decodePowerShellStdout(b []byte) []byte {
	b = bytes.TrimPrefix(b, []byte{0xEF, 0xBB, 0xBF})
	if bytes.HasPrefix(b, []byte{0xFF, 0xFE}) {
		return utf16LEBytes(b[2:])
	}
	if utf16LEASCII(b) {
		return utf16LEBytes(b)
	}
	return b
}

func utf16LEASCII(b []byte) bool {
	if len(b) < 4 || len(b)%2 != 0 {
		return false
	}
	samples := 0
	for i := 0; i+1 < len(b) && samples < 8; i += 2 {
		if b[i] == 0 && b[i+1] == 0 {
			continue
		}
		if b[i+1] != 0 || b[i] >= 0x80 {
			return false
		}
		samples++
	}
	return samples > 0
}

func utf16LEBytes(b []byte) []byte {
	if len(b)%2 == 1 {
		b = b[:len(b)-1]
	}
	u := make([]uint16, len(b)/2)
	for i := range u {
		u[i] = binary.LittleEndian.Uint16(b[i*2:])
	}
	return []byte(string(utf16.Decode(u)))
}

func jsonStringOrName(raw json.RawMessage) string {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return ""
	}
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return s
	}
	var obj struct {
		Name string `json:"Name"`
	}
	if json.Unmarshal(raw, &obj) == nil {
		return obj.Name
	}
	return ""
}

func jsonGender(raw json.RawMessage) string {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return ""
	}
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return s
	}
	var n int
	if json.Unmarshal(raw, &n) == nil {
		switch n {
		case 1:
			return "Male"
		case 2:
			return "Female"
		case 3:
			return "Neutral"
		default:
			return ""
		}
	}
	return ""
}
