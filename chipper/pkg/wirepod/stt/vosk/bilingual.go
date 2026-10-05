package wirepod_vosk

import (
	"encoding/json"
	"strings"
	"unicode"

	"github.com/kercre123/wire-pod/chipper/pkg/vars"
)

type voskResult struct {
	Text   string `json:"text"`
	Result []struct {
		Conf float64 `json:"conf"`
		Word string  `json:"word"`
	} `json:"result"`
}

func parseVoskResult(raw string) voskResult {
	var result voskResult
	_ = json.Unmarshal([]byte(raw), &result)
	result.Text = strings.TrimSpace(result.Text)
	return result
}

func averageConfidence(result voskResult) float64 {
	if len(result.Result) == 0 {
		return 0
	}
	var sum float64
	for _, word := range result.Result {
		sum += word.Conf
	}
	return sum / float64(len(result.Result))
}

func hasHan(text string) bool {
	for _, character := range text {
		if unicode.Is(unicode.Han, character) {
			return true
		}
	}
	return false
}

func matchesEnglishIntent(text string) bool {
	text = strings.ToLower(strings.TrimSpace(text))
	for _, intent := range vars.IntentList {
		for _, phrase := range intent.Keyphrases {
			phrase = strings.ToLower(phrase)
			if text == phrase || (!intent.RequireExactMatch && strings.Contains(text, phrase)) {
				return true
			}
		}
	}
	return false
}

// Keep English commands when recognition is plausible. A large Chinese
// confidence lead takes precedence because the English intent list contains
// short error-correction fragments that can match misheard Chinese speech.
func chooseBilingualResult(english, chinese voskResult) (string, string) {
	if !hasHan(chinese.Text) {
		return english.Text, "en-US"
	}
	if english.Text == "" {
		return chinese.Text, "zh-CN"
	}
	englishConfidence := averageConfidence(english)
	chineseConfidence := averageConfidence(chinese)
	if chineseConfidence > englishConfidence+0.12 {
		return chinese.Text, "zh-CN"
	}
	if matchesEnglishIntent(english.Text) {
		return english.Text, "en-US"
	}
	if chineseConfidence > englishConfidence+0.05 {
		return chinese.Text, "zh-CN"
	}
	return english.Text, "en-US"
}
