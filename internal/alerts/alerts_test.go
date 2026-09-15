package alerts

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFindToken(t *testing.T) {
	tests := map[string]struct {
		token, str, want string
	}{
		"simple":                 {"@", "@some", "some"},
		"only symbol":            {"@", "@", ""},
		"last symbol":            {"@", "some @", ""},
		"in the beginning":       {"@", "@some more text", "some"},
		"in the middle":          {"@", "check @some more", "some"},
		"in the end":             {"@", "check @some", "some"},
		"not found":              {"@", "check some", ""},
		"token in word":          {"@", "check @so@me", "so@me"},
		"in the middle, no word": {"@", "check @ more", ""},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if got := findToken(tc.token, tc.str); got != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
		})
	}
}

func writeYAML(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "commands.yaml")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestNew_rejectsNilOutAndEmptyToken(t *testing.T) {
	path := writeYAML(t, "commands: []\n")
	if _, err := New(Config{Token: "@", CommandsFilePath: path}); err == nil {
		t.Fatal("nil Out")
	}
	out := make(chan Clip, 1)
	if _, err := New(Config{Token: "", CommandsFilePath: path, Out: out}); err == nil {
		t.Fatal("empty token")
	}
}

func TestNew_missingFile(t *testing.T) {
	out := make(chan Clip, 1)
	_, err := New(Config{
		Token:            "@",
		CommandsFilePath: filepath.Join(t.TempDir(), "missing.yaml"),
		Out:              out,
	})
	if err == nil {
		t.Fatal("expected read error")
	}
}

func TestNew_invalidYAML(t *testing.T) {
	out := make(chan Clip, 1)
	path := writeYAML(t, "{[")
	_, err := New(Config{Token: "@", CommandsFilePath: path, Out: out})
	if err == nil {
		t.Fatal("expected yaml error")
	}
}

func TestNew_duplicateNames(t *testing.T) {
	out := make(chan Clip, 1)
	path := writeYAML(t, `
commands:
  - name: Jump
    file: a.mp3
  - name: jump
    file: b.mp3
`)
	_, err := New(Config{Token: "@", CommandsFilePath: path, Out: out})
	if err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("err = %v", err)
	}
}

func TestAlert_matchAndDefaults(t *testing.T) {
	out := make(chan Clip, 1)
	path := writeYAML(t, `
commands:
  - name: jump
    file: mario_jump.mp3
  - name: Dance
    file: cat.gif
    volume: 0.5
    scale: 1.2
`)
	a, err := New(Config{
		Token:            "@",
		MediaPath:        "/media",
		CommandsFilePath: path,
		Out:              out,
	})
	if err != nil {
		t.Fatal(err)
	}
	if a.MediaPath() != "/media" {
		t.Fatalf("media = %q", a.MediaPath())
	}
	if a.Alert("hello world") {
		t.Fatal("no token")
	}
	if a.Alert("please @unknown") {
		t.Fatal("unknown command")
	}
	if !a.Alert("go @JUMP now") {
		t.Fatal("expected jump")
	}
	clip := <-out
	if clip.Filename != "mario_jump.mp3" || clip.Volume != 1 || clip.Scale != 1 {
		t.Fatalf("%+v", clip)
	}
	if !a.Alert("@dance") {
		t.Fatal("expected dance")
	}
	clip = <-out
	if clip.Filename != "cat.gif" || clip.Volume != 0.5 || clip.Scale != 1.2 {
		t.Fatalf("%+v", clip)
	}
}
