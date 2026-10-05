package sdkapp

import (
	"context"
	"errors"
	"time"

	"github.com/fforchino/vector-go-sdk/pkg/vectorpb"
)

// PlayPCM plays 16-bit mono PCM through the SDK connection, which is also used
// by the web app. Cancellation stops playback when a voice request begins.
func PlayPCM(ctx context.Context, serial string, pcm []byte, sampleRate, volume uint32) error {
	if len(pcm) == 0 || len(pcm)%2 != 0 {
		return errors.New("PCM audio must contain 16-bit samples")
	}
	if sampleRate != 8000 && sampleRate != 16000 {
		return errors.New("unsupported PCM sample rate")
	}
	robot, index, err := getRobot(serial)
	if err != nil {
		return err
	}
	robots[index].ConnTimer = 0
	stream, err := robot.Vector.Conn.ExternalAudioStreamPlayback(ctx)
	if err != nil {
		return err
	}
	if err := stream.Send(&vectorpb.ExternalAudioStreamRequest{
		AudioRequestType: &vectorpb.ExternalAudioStreamRequest_AudioStreamPrepare{
			AudioStreamPrepare: &vectorpb.ExternalAudioStreamPrepare{AudioFrameRate: sampleRate, AudioVolume: volume},
		},
	}); err != nil {
		return err
	}
	for len(pcm) > 0 {
		if err := ctx.Err(); err != nil {
			return err
		}
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
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Duration(512*1000/sampleRate) * time.Millisecond):
		}
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
