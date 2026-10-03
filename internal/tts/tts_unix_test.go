//go:build !windows

package tts

import (
	"io"
	"slices"
	"testing"
)

func TestSynthCommand_textOnStdinOnly(t *testing.T) {
	const text = "-o/tmp/pwn.aiff --help -w/tmp/pwn.wav"
	for _, goos := range []string{"darwin", "linux"} {
		t.Run(goos, func(t *testing.T) {
			cmd, err := synthCommand(goos, text, "Alex", "/tmp/out.wav")
			if err != nil {
				t.Fatal(err)
			}
			if slices.Contains(cmd.Args, text) {
				t.Fatalf("text is in argv: %q", cmd.Args)
			}
			if cmd.Stdin == nil {
				t.Fatal("nil stdin")
			}
			got, err := io.ReadAll(cmd.Stdin)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != text {
				t.Fatalf("stdin = %q, want %q", got, text)
			}
		})
	}
}

func TestSynthCommand_unsupportedOS(t *testing.T) {
	if _, err := synthCommand("plan9", "hi", "", "/tmp/out.wav"); err == nil {
		t.Fatal("want error")
	}
}
