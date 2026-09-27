//go:build windows

package tts

import (
	"encoding/binary"
	"testing"
	"unicode/utf16"
)

func TestParseWindowsVoiceJSON_flatArray(t *testing.T) {
	raw := []byte(`[{"Name":"Microsoft David Desktop","Culture":"en-US","Gender":"Male","Description":"David"},{"Name":"Microsoft Zira Desktop","Culture":"en-US","Gender":"Female","Description":"Zira"}]`)
	got, err := parseWindowsVoiceJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].Name != "Microsoft David Desktop" || got[0].Gender != "Male" || got[0].Language != "en-US" {
		t.Fatalf("got %+v", got)
	}
	if got[1].Name != "Microsoft Zira Desktop" || got[1].Gender != "Female" {
		t.Fatalf("zira %+v", got[1])
	}
}

func TestParseWindowsVoiceJSON_singleObject(t *testing.T) {
	raw := []byte(`{"Name":"Microsoft Zira Desktop","Culture":"en-US","Gender":"Female","Description":"Zira"}`)
	got, err := parseWindowsVoiceJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Name != "Microsoft Zira Desktop" {
		t.Fatalf("got %+v", got)
	}
}

func TestParseWindowsVoiceJSON_legacyCultureAndGender(t *testing.T) {
	raw := []byte(`[{"Name":"Microsoft David Desktop","Culture":{"Name":"en-US","LCID":1033},"Gender":1,"Description":"David"}]`)
	got, err := parseWindowsVoiceJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Language != "en-US" || got[0].Gender != "Male" || got[0].Details != "David" {
		t.Fatalf("got %+v", got)
	}
}

func TestParseWindowsVoiceJSON_utf16(t *testing.T) {
	text := `[{"Name":"Microsoft Zira Desktop","Culture":"en-US","Gender":"Female","Description":"Zira"}]`
	raw := utf16LEWithBOM(text)
	got, err := parseWindowsVoiceJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Name != "Microsoft Zira Desktop" {
		t.Fatalf("got %+v", got)
	}
}

func TestParseWindowsVoiceJSON_empty(t *testing.T) {
	for _, raw := range []string{"", "null", "[]"} {
		got, err := parseWindowsVoiceJSON([]byte(raw))
		if err != nil {
			t.Fatal(raw, err)
		}
		if len(got) != 0 {
			t.Fatalf("%q got %+v", raw, got)
		}
	}
}

func utf16LEWithBOM(s string) []byte {
	u := utf16.Encode([]rune(s))
	b := make([]byte, 2+len(u)*2)
	b[0], b[1] = 0xFF, 0xFE
	for i, r := range u {
		binary.LittleEndian.PutUint16(b[2+i*2:], r)
	}
	return b
}
