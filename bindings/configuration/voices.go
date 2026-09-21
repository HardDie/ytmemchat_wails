package configuration

import (
	"strings"

	"github.com/HardDie/ytmemchat_wails/internal/tts"
)

// GetTTSVoices lists installed voices with language and gender. Empty if the engine is missing.
func (c *Configuration) GetTTSVoices() []TTSVoice {
	voices, err := tts.GetAvailableVoices()
	if err != nil {
		return []TTSVoice{}
	}
	return ttsVoicesFrom(voices)
}

func ttsVoicesFrom(list []tts.VoiceInfo) []TTSVoice {
	seen := make(map[string]int)
	out := make([]TTSVoice, 0, len(list))
	for _, v := range list {
		if v.Name == "" {
			continue
		}
		langs := tts.FormatLanguages(v.Language)
		if i, ok := seen[v.Name]; ok {
			out[i].Languages = mergeLanguageLabels(out[i].Languages, langs)
			if out[i].Gender == "" {
				out[i].Gender = v.Gender
			}
			if out[i].Details == "" {
				out[i].Details = v.Details
			}
			continue
		}
		seen[v.Name] = len(out)
		out = append(out, TTSVoice{
			Name:      v.Name,
			Languages: langs,
			Gender:    v.Gender,
			Details:   v.Details,
		})
	}
	return out
}

func mergeLanguageLabels(existing, next string) string {
	if next == "" {
		return existing
	}
	if existing == "" {
		return next
	}
	seen := map[string]struct{}{}
	var parts []string
	for _, p := range strings.Split(existing+", "+next, ", ") {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		parts = append(parts, p)
	}
	return strings.Join(parts, ", ")
}
