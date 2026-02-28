package audio

import (
	"os"
	"testing"
)

func TestWriteWAV(t *testing.T) {
	samples := make([]int16, 16000) // 1 second of silence at 16kHz
	path := t.TempDir() + "/test.wav"
	err := WriteWAV(path, samples, 16000, 1, 16)
	if err != nil {
		t.Fatalf("WriteWAV error: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat error: %v", err)
	}
	expectedSize := int64(44 + 32000) // header + data
	if info.Size() != expectedSize {
		t.Errorf("expected file size %d, got %d", expectedSize, info.Size())
	}
}
