package wirepod_ttr

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/fforchino/vector-go-sdk/pkg/vector"
	"github.com/fforchino/vector-go-sdk/pkg/vectorpb"
)

// DoSayText_WindowsChinese uses the Chinese voice installed on the WirePod PC.
// Vector's own SayText voice pronounces Chinese text as English.
func DoSayText_WindowsChinese(robot *vector.Vector, input string) error {
	if strings.TrimSpace(input) == "" {
		return nil
	}
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	wavFile, err := os.CreateTemp("", "wirepod-zh-*.wav")
	if err != nil {
		return err
	}
	wavPath := wavFile.Name()
	wavFile.Close()
	defer os.Remove(wavPath)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	script := filepath.Join(filepath.Dir(exe), "chinese-tts.ps1")
	encodedText := base64.StdEncoding.EncodeToString([]byte(input))
	cmd := exec.CommandContext(ctx, "powershell.exe", "-NoProfile", "-NonInteractive", "-WindowStyle", "Hidden", "-ExecutionPolicy", "Bypass", "-File", script, encodedText, wavPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("speech synthesis: %w: %s", err, strings.TrimSpace(string(output)))
	}
	wav, err := os.ReadFile(wavPath)
	if err != nil {
		return err
	}
	pcm, err := pcmFromWave16k(wav)
	if err != nil {
		return err
	}

	stream, err := robot.Conn.ExternalAudioStreamPlayback(context.Background())
	if err != nil {
		return err
	}
	if err := stream.Send(&vectorpb.ExternalAudioStreamRequest{
		AudioRequestType: &vectorpb.ExternalAudioStreamRequest_AudioStreamPrepare{
			AudioStreamPrepare: &vectorpb.ExternalAudioStreamPrepare{AudioFrameRate: 16000, AudioVolume: 100},
		},
	}); err != nil {
		return err
	}
	for len(pcm) > 0 {
		chunk := make([]byte, 1024)
		n := copy(chunk, pcm)
		pcm = pcm[n:]
		if err := stream.Send(&vectorpb.ExternalAudioStreamRequest{
			AudioRequestType: &vectorpb.ExternalAudioStreamRequest_AudioStreamChunk{
				AudioStreamChunk: &vectorpb.ExternalAudioStreamChunk{AudioChunkSizeBytes: 1024, AudioChunkSamples: chunk},
			},
		}); err != nil {
			return err
		}
		time.Sleep(32 * time.Millisecond)
	}
	if err := stream.Send(&vectorpb.ExternalAudioStreamRequest{
		AudioRequestType: &vectorpb.ExternalAudioStreamRequest_AudioStreamComplete{
			AudioStreamComplete: &vectorpb.ExternalAudioStreamComplete{},
		},
	}); err != nil {
		return err
	}
	return stream.CloseSend()
}

func pcmFromWave16k(wav []byte) ([]byte, error) {
	if len(wav) < 12 || string(wav[:4]) != "RIFF" || string(wav[8:12]) != "WAVE" {
		return nil, errors.New("invalid WAV header")
	}
	var formatOK bool
	var pcm []byte
	for offset := 12; offset+8 <= len(wav); {
		size := int(binary.LittleEndian.Uint32(wav[offset+4 : offset+8]))
		start := offset + 8
		if size > len(wav)-start {
			return nil, errors.New("invalid WAV chunk length")
		}
		switch string(wav[offset : offset+4]) {
		case "fmt ":
			if size < 16 {
				return nil, errors.New("invalid WAV format chunk")
			}
			formatOK = binary.LittleEndian.Uint16(wav[start:start+2]) == 1 &&
				binary.LittleEndian.Uint16(wav[start+2:start+4]) == 1 &&
				binary.LittleEndian.Uint32(wav[start+4:start+8]) == 16000 &&
				binary.LittleEndian.Uint16(wav[start+14:start+16]) == 16
		case "data":
			pcm = wav[start : start+size]
		}
		offset = start + size + size%2
	}
	if !formatOK || len(pcm) == 0 || len(pcm)%2 != 0 {
		return nil, errors.New("WAV must contain 16 kHz mono 16-bit PCM")
	}
	return pcm, nil
}
