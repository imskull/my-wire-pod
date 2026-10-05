package wirepod_ttr

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestPCMFromWave16k(t *testing.T) {
	fmtChunk := make([]byte, 16)
	binary.LittleEndian.PutUint16(fmtChunk[0:2], 1)
	binary.LittleEndian.PutUint16(fmtChunk[2:4], 1)
	binary.LittleEndian.PutUint32(fmtChunk[4:8], 16000)
	binary.LittleEndian.PutUint16(fmtChunk[14:16], 16)
	pcm := []byte{1, 2, 3, 4}
	wav := []byte("RIFF\x00\x00\x00\x00WAVE")
	wav = append(wav, []byte("fmt ")...)
	wav = binary.LittleEndian.AppendUint32(wav, uint32(len(fmtChunk)))
	wav = append(wav, fmtChunk...)
	wav = append(wav, []byte("data")...)
	wav = binary.LittleEndian.AppendUint32(wav, uint32(len(pcm)))
	wav = append(wav, pcm...)

	got, err := pcmFromWave16k(wav)
	if err != nil || !bytes.Equal(got, pcm) {
		t.Fatalf("got %v, %v; want PCM %v", got, err, pcm)
	}
	binary.LittleEndian.PutUint32(wav[24:28], 24000)
	if _, err := pcmFromWave16k(wav); err == nil {
		t.Fatal("accepted WAV with unsupported sample rate")
	}
}
