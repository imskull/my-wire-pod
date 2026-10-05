package barks

import (
	"context"
	_ "embed"
	"errors"
	"math/rand"
	"sync"
	"time"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
	"github.com/kercre123/wire-pod/chipper/pkg/vars"
	"github.com/kercre123/wire-pod/chipper/pkg/wirepod/sdkapp"
)

//go:embed dog-bark.pcm
var dogBarkPCM []byte // 16 kHz, mono, signed 16-bit little-endian PCM

const (
	minimumDelay  = 20 * time.Minute
	maximumDelay  = 60 * time.Minute
	voiceCooldown = 5 * time.Minute
	dayStart      = 9
	dayEnd        = 21
)

var (
	startOnce sync.Once
	stateMu   sync.Mutex
	active    int
	lastVoice = time.Now()
	stopBark  context.CancelFunc
)

// VoiceRequestStarted prevents an idle bark from overlapping a voice request.
// Call the returned function when the request has finished.
func VoiceRequestStarted() func() {
	stateMu.Lock()
	active++
	lastVoice = time.Now()
	if stopBark != nil {
		stopBark()
	}
	stateMu.Unlock()
	return func() {
		stateMu.Lock()
		active--
		lastVoice = time.Now()
		stateMu.Unlock()
	}
}

func eligible(now, lastVoice time.Time, active int) bool {
	return now.Hour() >= dayStart && now.Hour() < dayEnd &&
		active == 0 && now.Sub(lastVoice) >= voiceCooldown
}

func randomDelay(random *rand.Rand) time.Duration {
	return minimumDelay + time.Duration(random.Int63n(int64(maximumDelay-minimumDelay)))
}

// Start enables occasional daytime barks when configured in apiConfig.json.
func Start() {
	if !vars.APIConfig.PastInitialSetup || !vars.APIConfig.Server.RandomBarks {
		return
	}
	startOnce.Do(func() {
		logger.Println("Idle dog barks enabled (daytime, approximately every 20-60 minutes)")
		go run()
	})
}

func run() {
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	for {
		time.Sleep(randomDelay(random))
		stateMu.Lock()
		ready := eligible(time.Now(), lastVoice, active)
		if !ready {
			stateMu.Unlock()
			continue
		}
		ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
		stopBark = cancel
		stateMu.Unlock()

		for _, bot := range vars.BotInfo.Robots {
			if err := playForRobot(ctx, bot.Esn); err != nil && !errors.Is(err, context.Canceled) {
				logger.Println("Idle dog bark failed for", bot.Esn, ":", err)
			}
		}
		stateMu.Lock()
		stopBark = nil
		stateMu.Unlock()
		cancel()
	}
}

func playForRobot(ctx context.Context, esn string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := sdkapp.PlayPCM(ctx, esn, dogBarkPCM, 16000, 60); err != nil {
		return err
	}
	logger.Println("Played idle dog bark for", esn)
	return nil
}
