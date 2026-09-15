//go:build integration

package tts

import (
	"os/exec"
	"testing"
)

func skipIfMissing(t *testing.T, bin string) {
	t.Helper()
	if _, err := exec.LookPath(bin); err != nil {
		t.Skipf("%s not found", bin)
	}
}

func integrationVoice(t *testing.T) string {
	t.Helper()
	voices, err := GetAvailableVoices()
	if err != nil || len(voices) == 0 {
		t.Skip("no voices listed")
	}
	return voices[0].Name
}

func testGetAvailableVoices(t *testing.T) {
	t.Helper()
	voices, err := GetAvailableVoices()
	if err != nil {
		t.Fatal(err)
	}
	if len(voices) == 0 {
		t.Fatal("expected at least one installed voice")
	}
	if voices[0].Name == "" {
		t.Fatalf("empty voice name: %+v", voices[0])
	}
}

func testSynthesizeToBuffer(t *testing.T) {
	t.Helper()
	voice := integrationVoice(t)
	data, format, err := SynthesizeToBuffer("hello", voice)
	if err != nil {
		t.Fatal(err)
	}
	if format != "wav" {
		t.Fatalf("format = %q, want wav", format)
	}
	if len(data) < 44 {
		t.Fatalf("WAV too small: %d bytes", len(data))
	}
}

func testSynthesizeAudio(t *testing.T) {
	t.Helper()
	out := make(chan Speech, 1)
	eng := New(Config{VoiceName: integrationVoice(t), Out: out})
	if err := eng.SynthesizeAudio("hello overlay"); err != nil {
		t.Fatal(err)
	}
	got := <-out
	if len(got.WAV) < 44 {
		t.Fatalf("WAV too small: %d", len(got.WAV))
	}
	if got.Volume != 1 {
		t.Fatalf("volume = %v, want 1", got.Volume)
	}
}
