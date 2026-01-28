// Write WAV file

package audio

import (
	"encoding/binary"
	"os"
)

func WriteWav(path string, chunks []WavDataChunk) error {
	if len(chunks) == 0 {
		return nil
	}
	metadata := chunks[0].Metadata

	// total data size
	var totalDataSize uint32
	for _, chunk := range chunks {
		totalDataSize += uint32(len(chunk.Samples))
	}

	// Create file in path
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write riff data
	if _, err := file.Write([]byte("RIFF")); err != nil {
		return err
	}

	// ChunkSize: 36 + SubChunk2Size
	// 4 (WAVE) + 24 (fmt chunk) + 8 (data chunk header) = 36
	fileSize := 36 + totalDataSize
	if err := binary.Write(file, binary.LittleEndian, fileSize); err != nil {
		return err
	}

	// Write wave data
	if _, err := file.Write([]byte("WAVE")); err != nil {
		return err
	}

	// Write fmt data
	if _, err := file.Write([]byte("fmt ")); err != nil {
		return err
	}

	// Subchunk1Size: 16 for PCM
	if err := binary.Write(file, binary.LittleEndian, uint32(16)); err != nil {
		return err
	}

	// AudioFormat: PCM = 1
	if err := binary.Write(file, binary.LittleEndian, metadata.Format); err != nil {
		return err
	}

	// NumChannels
	if err := binary.Write(file, binary.LittleEndian, metadata.Channels); err != nil {
		return err
	}

	// SampleRate
	if err := binary.Write(file, binary.LittleEndian, metadata.SampleRate); err != nil {
		return err
	}

	// ByteRate = SampleRate * NumChannels * BitsPerSample/8
	byteRate := metadata.SampleRate * uint32(metadata.Channels) * uint32(metadata.Bitdepth) / 8
	if err := binary.Write(file, binary.LittleEndian, byteRate); err != nil {
		return err
	}

	// BlockAlign = NumChannels * BitsPerSample/8
	blockAlign := metadata.Channels * metadata.Bitdepth / 8
	if err := binary.Write(file, binary.LittleEndian, blockAlign); err != nil {
		return err
	}

	// BitsPerSample
	if err := binary.Write(file, binary.LittleEndian, metadata.Bitdepth); err != nil {
		return err
	}

	// Subchunk2ID: "data"
	if _, err := file.Write([]byte("data")); err != nil {
		return err
	}

	// Subchunk2Size: numBytes in data
	if err := binary.Write(file, binary.LittleEndian, totalDataSize); err != nil {
		return err
	}

	// Write data chunks
	for _, chunk := range chunks {
		if _, err := file.Write(chunk.Samples); err != nil {
			return err
		}
	}

	return nil
}
