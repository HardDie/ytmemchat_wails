package tts

import (
	"strings"
	"testing"
)

func TestRemoveEmojis(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "simple emoji", input: "Hello World! 😊", expected: "Hello World!"},
		{name: "multiple emojis", input: "🔥 Live Chat 🔥 is crazy 🚀", expected: "Live Chat is crazy"},
		{name: "emoji with skin tone", input: "High five! ✋🏾", expected: "High five!"},
		{name: "flag and complex symbols", input: "Welcome from 🇺🇸! ✌️", expected: "Welcome from !"},
		{name: "text only", input: "Just a normal message.", expected: "Just a normal message."},
		{name: "only emoji", input: "🔥🚀", expected: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := removeEmojis(tt.input)
			want := strings.Join(strings.Fields(tt.expected), " ")
			if got != want {
				t.Errorf("removeEmojis() = %q, want %q", got, want)
			}
		})
	}
}

func TestSynthesizeAudio_rejectsEmptyAndNilOut(t *testing.T) {
	t.Run("nil out", func(t *testing.T) {
		eng := New(Config{VoiceName: "Alex"})
		if err := eng.SynthesizeAudio("hello"); err == nil {
			t.Fatal("expected error for nil Out")
		}
	})
	t.Run("emoji only", func(t *testing.T) {
		out := make(chan Speech, 1)
		eng := New(Config{VoiceName: "Alex", Out: out})
		if err := eng.SynthesizeAudio("🔥"); err == nil {
			t.Fatal("expected error for empty speakable text")
		}
	})
}

func TestNew_defaultVolume(t *testing.T) {
	eng := New(Config{})
	if eng.volume != 1 {
		t.Fatalf("volume = %v, want 1", eng.volume)
	}
	v := 0.25
	eng = New(Config{Volume: &v})
	if eng.volume != 0.25 {
		t.Fatalf("volume = %v, want 0.25", eng.volume)
	}
}

func TestParseSayVoiceList(t *testing.T) {
	input := "Alex                en_US    # Most people recognize me by my voice.\n" +
		"Milena              ru_RU    # Здравствуйте! Меня зовут Милена.\n" +
		"Bad News            en_US    # this line should be skipped (space in name)\n"
	got := parseSayVoiceList(input)
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2: %+v", len(got), got)
	}
	if got[0].Name != "Alex" || got[0].Language != "en_US" {
		t.Fatalf("Alex row: %+v", got[0])
	}
	if got[1].Name != "Milena" || got[1].Language != "ru_RU" {
		t.Fatalf("Milena row: %+v", got[1])
	}
}

func TestParseEspeakVoiceList(t *testing.T) {
	input := "Pty Language Age/Gender VoiceName File Other Languages\n" +
		" 5  en-gb          M  english                  gmw/en         (en-uk 2)(en 2)\n"
	got := parseEspeakVoiceList(input)
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1: %+v", len(got), got)
	}
	if got[0].Name != "en-gb" || got[0].Language != "en-gb, en-uk, en" || got[0].Gender != "Male" {
		t.Fatalf("parsed row: %+v", got[0])
	}
	if got[0].Details != "english" {
		t.Fatalf("details %+v", got[0])
	}
}

func TestFormatLanguages(t *testing.T) {
	got := FormatLanguages("ru_RU")
	if !strings.Contains(got, "Russian") || !strings.Contains(got, "ru_RU") {
		t.Fatalf("got %q", got)
	}
	got = FormatLanguages("en-gb, en")
	if !strings.Contains(got, "English") {
		t.Fatalf("got %q", got)
	}
}
