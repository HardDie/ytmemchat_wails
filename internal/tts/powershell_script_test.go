//go:build windows

package tts

import (
	"strings"
	"testing"
)

func TestPSSynthEnv_passesValuesVerbatim(t *testing.T) {
	tests := []struct {
		name  string
		text  string
		voice string
		out   string
	}{
		{name: "ascii quote", text: "it's '; calc; '", voice: "Microsoft Zira Desktop", out: `C:\Temp\a.wav`},
		{name: "typographic quotes", text: "don’t; Start-Process calc; ’ ‘ ‚ ‛", voice: "O’Brien", out: `C:\Users\O'Brien\a.wav`},
		{name: "subexpression", text: "$(Start-Process calc) `n", voice: "", out: `C:\Temp\$x.wav`},
		{name: "unicode", text: "Привет, мир", voice: "Microsoft Irina Desktop", out: `C:\Temp\b.wav`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := psSynthEnv(tt.text, tt.voice, tt.out)
			want := []string{
				psEnvText + "=" + tt.text,
				psEnvVoice + "=" + tt.voice,
				psEnvOut + "=" + tt.out,
			}
			if strings.Join(env, "\n") != strings.Join(want, "\n") {
				t.Fatalf("env = %q, want %q", env, want)
			}
		})
	}
}

func TestPSSynthScript_readsEveryInputFromEnv(t *testing.T) {
	for _, name := range []string{psEnvText, psEnvVoice, psEnvOut} {
		if !strings.Contains(psSynthScript, "$env:"+name) {
			t.Errorf("script does not read $env:%s", name)
		}
	}
	// A format verb or a quoted literal would mean input is pasted into code again.
	for _, bad := range []string{"%s", "%v", "'"} {
		if strings.Contains(psSynthScript, bad) {
			t.Errorf("script contains %q", bad)
		}
	}
}
