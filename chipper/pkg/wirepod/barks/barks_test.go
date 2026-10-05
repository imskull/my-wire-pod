package barks

import (
	"math/rand"
	"testing"
	"time"
)

func TestEligibleBarkTime(t *testing.T) {
	day := time.Date(2026, 10, 5, 12, 0, 0, 0, time.Local)
	for _, tc := range []struct {
		name   string
		now    time.Time
		last   time.Time
		active int
		want   bool
	}{
		{"idle daytime", day, day.Add(-6 * time.Minute), 0, true},
		{"recent speech", day, day.Add(-2 * time.Minute), 0, false},
		{"speaking now", day, day.Add(-6 * time.Minute), 1, false},
		{"before daytime", day.Add(-4 * time.Hour), day.Add(-6 * time.Minute), 0, false},
		{"after daytime", day.Add(9 * time.Hour), day.Add(-6 * time.Minute), 0, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := eligible(tc.now, tc.last, tc.active); got != tc.want {
				t.Fatalf("eligible = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestBarkDelayBoundsAndAudio(t *testing.T) {
	random := rand.New(rand.NewSource(1))
	for i := 0; i < 100; i++ {
		delay := randomDelay(random)
		if delay < minimumDelay || delay >= maximumDelay {
			t.Fatalf("delay outside bounds: %s", delay)
		}
	}
	if len(dogBarkPCM) < 16000 || len(dogBarkPCM)%2 != 0 {
		t.Fatalf("embedded bark is empty or invalid: %d bytes", len(dogBarkPCM))
	}
}
