package audio

import (
	"encoding/binary"
	"os"
)

func WriteWAV(path string, samples []int16, sampleRate, numChannels, bitDepth int) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	dataSize := len(samples) * numChannels * (bitDepth / 8)

	f.Write([]byte("RIFF"))
	binary.Write(f, binary.LittleEndian, uint32(36+dataSize))
	f.Write([]byte("WAVE"))
	f.Write([]byte("fmt "))
	binary.Write(f, binary.LittleEndian, uint32(16))
	binary.Write(f, binary.LittleEndian, uint16(1))
	binary.Write(f, binary.LittleEndian, uint16(numChannels))
	binary.Write(f, binary.LittleEndian, uint32(sampleRate))
	binary.Write(f, binary.LittleEndian, uint32(sampleRate*numChannels*bitDepth/8))
	binary.Write(f, binary.LittleEndian, uint16(numChannels*bitDepth/8))
	binary.Write(f, binary.LittleEndian, uint16(bitDepth))
	f.Write([]byte("data"))
	binary.Write(f, binary.LittleEndian, uint32(dataSize))
	binary.Write(f, binary.LittleEndian, samples)

	return nil
}
