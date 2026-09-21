package configuration

import (
	"strings"
	"testing"

	"github.com/HardDie/ytmemchat_wails/internal/tts"
)

func TestTTSVoicesFrom_mergesLanguages(t *testing.T) {
	got := ttsVoicesFrom([]tts.VoiceInfo{
		{Name: "Alex", Language: "en_US", Gender: "Male", Details: "hi"},
		{Name: "Alex", Language: "en_GB"},
		{Name: "", Language: "xx"},
	})
	if len(got) != 1 {
		t.Fatalf("%+v", got)
	}
	if got[0].Name != "Alex" || got[0].Gender != "Male" {
		t.Fatalf("%+v", got[0])
	}
	if !strings.Contains(got[0].Languages, "en_US") || !strings.Contains(got[0].Languages, "en_GB") {
		t.Fatalf("languages %q", got[0].Languages)
	}
}
