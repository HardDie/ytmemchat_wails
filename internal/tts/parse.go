package tts

import (
	"fmt"
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
			Gender:   gender,
			Details:  details,
		})
	}
	return voices
}

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
		voices = append(voices, VoiceInfo{
			Name:     fields[1],
			Language: fields[2],
			Gender:   fields[3],
			Details:  fmt.Sprintf("Age: %s", fields[4]),
		})
	}
	return voices
}
