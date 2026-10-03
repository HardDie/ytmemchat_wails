//go:build integration

package tts

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// requireTool fails the test when the OS TTS engine is missing.
// These tests must not skip: CI should go red if synthesis cannot run.
func requireTool(t *testing.T, bin string) {
	t.Helper()
	if _, err := exec.LookPath(bin); err != nil {
		t.Fatalf("%s not found: %v", bin, err)
	}
}

func integrationVoice(t *testing.T) string {
	t.Helper()
	voices, err := GetAvailableVoices()
	if err != nil {
		t.Fatalf("list voices: %v", err)
	}
	if len(voices) == 0 {
		t.Fatal("no voices listed")
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

// testSynthesizeHostile speaks a chat line built to create pwn through the engine.
// The line must be spoken as plain text, and pwn must not exist afterwards.
func testSynthesizeHostile(t *testing.T, line func(pwn string) string) {
	t.Helper()
	voice := integrationVoice(t)
	pwn := filepath.Join(t.TempDir(), "pwn")
	data, _, err := SynthesizeToBuffer(line(pwn), voice)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) < 44 {
		t.Fatalf("WAV too small: %d bytes", len(data))
	}
	if _, err := os.Stat(pwn); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("chat text created %s (stat err %v)", pwn, err)
	}
}
