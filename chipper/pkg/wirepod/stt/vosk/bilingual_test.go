package wirepod_vosk

import (
	"testing"

	"github.com/kercre123/wire-pod/chipper/pkg/vars"
)

func TestChooseBilingualResult(t *testing.T) {
	oldIntents := vars.IntentList
	vars.IntentList = []vars.JsonIntent{{Name: "intent_system_charger", Keyphrases: []string{"go home"}}}
	defer func() { vars.IntentList = oldIntents }()

	tests := []struct {
		name    string
		english string
		chinese string
		want    string
		lang    string
	}{
		{"English command has priority", `{"text":"go home","result":[{"word":"go","conf":0.90}]}`, `{"text":"给我","result":[{"word":"给我","conf":0.85}]}`, "go home", "en-US"},
		{"Misheard Chinese overrides English fragment", `{"text":"go home","result":[{"word":"go","conf":0.60}]}`, `{"text":"给我讲故事","result":[{"word":"讲","conf":0.95}]}`, "给我讲故事", "zh-CN"},
		{"Chinese chat", `{"text":"can you tell me","result":[{"word":"can","conf":0.52}]}`, `{"text":"讲一个故事","result":[{"word":"讲","conf":0.90}]}`, "讲一个故事", "zh-CN"},
		{"English chat", `{"text":"tell me a story","result":[{"word":"tell","conf":0.91}]}`, `{"text":"太热","result":[{"word":"太热","conf":0.55}]}`, "tell me a story", "en-US"},
		{"No English transcription", `{"text":""}`, `{"text":"你好","result":[{"word":"你好","conf":0.70}]}`, "你好", "zh-CN"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, lang := chooseBilingualResult(parseVoskResult(tc.english), parseVoskResult(tc.chinese))
			if got != tc.want || lang != tc.lang {
				t.Fatalf("got %q (%s), want %q (%s)", got, lang, tc.want, tc.lang)
			}
		})
	}
}
