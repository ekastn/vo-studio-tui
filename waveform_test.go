package main

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// createSyntheticWAV creates a minimal 16-bit mono WAV file for testing.
func createSyntheticWAV(sampleRate int, durationSec float64) []byte {
	numSamples := int(float64(sampleRate) * durationSec)
	dataSize := uint32(numSamples * 2) // 16-bit = 2 bytes per sample
	byteRate := uint32(sampleRate * 2)

	var buf bytes.Buffer
	// RIFF header
	buf.WriteString("RIFF")
	binary.Write(&buf, binary.LittleEndian, uint32(36+dataSize))
	buf.WriteString("WAVE")

	// fmt chunk
	buf.WriteString("fmt ")
	binary.Write(&buf, binary.LittleEndian, uint32(16))     // Chunk size
	binary.Write(&buf, binary.LittleEndian, uint16(1))      // Audio format: PCM
	binary.Write(&buf, binary.LittleEndian, uint16(1))      // Channels: 1
	binary.Write(&buf, binary.LittleEndian, uint32(sampleRate))
	binary.Write(&buf, binary.LittleEndian, byteRate)
	binary.Write(&buf, binary.LittleEndian, uint16(2))      // Block align
	binary.Write(&buf, binary.LittleEndian, uint16(16))     // Bits per sample

	// data chunk
	buf.WriteString("data")
	binary.Write(&buf, binary.LittleEndian, dataSize)

	// Samples (sine wave or simple pattern)
	for i := 0; i < numSamples; i++ {
		val := int16(16000)
		if i%2 == 0 {
			val = -16000
		}
		binary.Write(&buf, binary.LittleEndian, val)
	}

	return buf.Bytes()
}

func TestWaveform_DurationAndExtraction(t *testing.T) {
	tempDir := t.TempDir()
	wavPath := filepath.Join(tempDir, "test.wav")

	wavBytes := createSyntheticWAV(24000, 2.5) // 2.5 seconds
	if err := os.WriteFile(wavPath, wavBytes, 0644); err != nil {
		t.Fatalf("Failed to write test wav: %v", err)
	}

	// Test duration reader
	dur, err := readWAVDuration(wavPath)
	if err != nil {
		t.Fatalf("readWAVDuration failed: %v", err)
	}
	if dur < 2.49 || dur > 2.51 {
		t.Errorf("Expected duration ~2.5s, got %.3fs", dur)
	}

	// Test waveform extraction
	bins, err := ExtractWaveform(wavPath, 16)
	if err != nil {
		t.Fatalf("ExtractWaveform failed: %v", err)
	}
	if len(bins) != 16 {
		t.Fatalf("Expected 16 bins, got %d", len(bins))
	}
	for i, b := range bins {
		if b <= 0.0 || b > 1.0 {
			t.Errorf("Bin %d out of normalized range: %f", i, b)
		}
	}

	// Test waveform rendering
	rendered := RenderWaveform(bins, 0.5, 40)
	if !strings.Contains(rendered, "Waveform") {
		t.Errorf("Expected rendered waveform to contain header")
	}
}
