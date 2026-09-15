package tts

import (
	"regexp"
	"strings"
)

var sayVoiceLine = regexp.MustCompile(`^(\w+)\s+([a-z]{2}_[A-Z]{2})[^\n]*#\s*([^\n]+)`)

func parseSayVoiceList(output string) []VoiceInfo {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	var voices []VoiceInfo
	for _, line := range lines {
		matches := sayVoiceLine.FindStringSubmatch(line)
		if len(matches) != 4 {
			continue
		}
		details := strings.TrimSpace(matches[3])
		gender := ""
		switch {
		case strings.Contains(details, "Male"):
			gender = "Male"
		case strings.Contains(details, "Female"):
			gender = "Female"
		}
		voices = append(voices, VoiceInfo{
			Name:     matches[1],
			Language: matches[2],
			Gender:   normalizeGender(gender),
			Details:  details,
		})
	}
	return voices
}

var espeakOtherLang = regexp.MustCompile(`\(([^\s)]+)`)

func parseEspeakVoiceList(output string) []VoiceInfo {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	var voices []VoiceInfo
	for i, line := range lines {
		if i < 1 || strings.Contains(line, "Language") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}
		langs := []string{fields[1]}
		if len(fields) >= 6 {
			rest := strings.Join(fields[5:], " ")
			for _, m := range espeakOtherLang.FindAllStringSubmatch(rest, -1) {
				if m[1] != "" && m[1] != fields[1] {
					langs = append(langs, m[1])
				}
			}
		}
		voices = append(voices, VoiceInfo{
			Name:     fields[1],
			Language: strings.Join(uniqueStrings(langs), ", "),
			Gender:   normalizeGender(fields[2]),
			Details:  fields[3],
		})
	}
	return voices
}

func uniqueStrings(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}
