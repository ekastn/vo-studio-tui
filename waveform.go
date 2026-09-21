// Package main implements PCM waveform extraction and terminal meter visualization.
package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// readWAVDuration parses the RIFF header to calculate audio duration directly in pure Go.
func readWAVDuration(path string) (float64, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	header := make([]byte, 12)
	if _, err := io.ReadFull(f, header); err != nil {
		return 0, err
	}
	if string(header[0:4]) != "RIFF" || string(header[8:12]) != "WAVE" {
		return 0, fmt.Errorf("not a RIFF/WAVE file")
	}

	var byteRate uint32
	var dataSize uint32

	chunkHdr := make([]byte, 8)
	for {
		if _, err := io.ReadFull(f, chunkHdr); err != nil {
			break
		}
		chunkID := string(chunkHdr[0:4])
		chunkLen := binary.LittleEndian.Uint32(chunkHdr[4:8])

		if chunkID == "fmt " {
			fmtData := make([]byte, chunkLen)
			if _, err := io.ReadFull(f, fmtData); err != nil {
				return 0, err
			}
			if chunkLen >= 12 {
				byteRate = binary.LittleEndian.Uint32(fmtData[8:12])
			}
		} else if chunkID == "data" {
			dataSize = chunkLen
			if byteRate > 0 && dataSize > 0 {
				return float64(dataSize) / float64(byteRate), nil
			}
			if _, err := f.Seek(int64(chunkLen), io.SeekCurrent); err != nil {
				break
			}
		} else {
			if _, err := f.Seek(int64(chunkLen), io.SeekCurrent); err != nil {
				break
			}
		}
	}

	if byteRate > 0 && dataSize > 0 {
		return float64(dataSize) / float64(byteRate), nil
	}
	return 0, fmt.Errorf("unable to read WAV chunks")
}

// GetAudioDuration queries duration via pure Go WAV reader, falling back to ffprobe.
func GetAudioDuration(path string) float64 {
	if dur, err := readWAVDuration(path); err == nil && dur > 0 {
		return dur
	}
	if _, err := os.Stat(path); err != nil {
		return 0.0
	}
	cmd := exec.Command("ffprobe", "-v", "error", "-show_entries", "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", path)
	out, err := cmd.Output()
	if err != nil {
		return 0.0
	}
	val, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	if err != nil {
		return 0.0
	}
	return val
}

// ExtractWaveform reads 16-bit PCM audio samples and normalizes peak amplitudes across bins.
func ExtractWaveform(path string, bins int) ([]float64, error) {
	if bins <= 0 {
		bins = 32
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	header := make([]byte, 12)
	if _, err := io.ReadFull(f, header); err != nil {
		return nil, err
	}
	if string(header[0:4]) != "RIFF" || string(header[8:12]) != "WAVE" {
		return nil, fmt.Errorf("not a RIFF/WAVE file")
	}

	var channels uint16 = 1
	var bitsPerSample uint16 = 16
	var dataOffset int64
	var dataSize int64

	chunkHdr := make([]byte, 8)
	for {
		if _, err := io.ReadFull(f, chunkHdr); err != nil {
			break
		}
		chunkID := string(chunkHdr[0:4])
		chunkLen := int64(binary.LittleEndian.Uint32(chunkHdr[4:8]))

		if chunkID == "fmt " {
			fmtData := make([]byte, chunkLen)
			if _, err := io.ReadFull(f, fmtData); err != nil {
				return nil, err
			}
			if chunkLen >= 16 {
				channels = binary.LittleEndian.Uint16(fmtData[2:4])
				bitsPerSample = binary.LittleEndian.Uint16(fmtData[14:16])
			}
		} else if chunkID == "data" {
			curPos, _ := f.Seek(0, io.SeekCurrent)
			dataOffset = curPos
			dataSize = chunkLen
			break
		} else {
			if _, err := f.Seek(chunkLen, io.SeekCurrent); err != nil {
				break
			}
		}
	}

	if dataOffset == 0 || dataSize == 0 || bitsPerSample != 16 {
		return nil, fmt.Errorf("unsupported WAV format for waveform extraction")
	}

	totalSamples := dataSize / int64(2*channels)
	if totalSamples <= 0 {
		return make([]float64, bins), nil
	}

	samplesPerBin := totalSamples / int64(bins)
	if samplesPerBin == 0 {
		samplesPerBin = 1
	}

	result := make([]float64, bins)
	bytesPerSampleFrame := int(2 * channels)
	buf := make([]byte, bytesPerSampleFrame)

	for b := 0; b < bins; b++ {
		var maxAmp float64
		binSampleCount := samplesPerBin
		step := int64(1)
		if binSampleCount > 100 {
			step = binSampleCount / 100
		}

		for s := int64(0); s < binSampleCount; s += step {
			sampleIndex := int64(b)*samplesPerBin + s
			if sampleIndex >= totalSamples {
				break
			}

			fileOffset := dataOffset + sampleIndex*int64(bytesPerSampleFrame)
			if _, err := f.Seek(fileOffset, io.SeekStart); err != nil {
				break
			}
			if _, err := io.ReadFull(f, buf); err != nil {
				break
			}

			sample := int16(binary.LittleEndian.Uint16(buf[0:2]))
			absVal := math.Abs(float64(sample)) / 32768.0
			if absVal > maxAmp {
				maxAmp = absVal
			}
		}
		result[b] = maxAmp
	}

	return result, nil
}

// RenderWaveform formats the peak amplitude bins into a visual Unicode block meter.
func RenderWaveform(waveform []float64, progress float64, width int) string {
	if len(waveform) == 0 {
		return ""
	}

	levels := []rune{' ', ' ', '▂', '▃', '▄', '▅', '▆', '▇', '█'}
	playheadBin := -1
	if progress >= 0.0 && progress <= 1.0 {
		playheadBin = int(float64(len(waveform)) * progress)
		if playheadBin >= len(waveform) {
			playheadBin = len(waveform) - 1
		}
	}

	var sb strings.Builder
	for i, amp := range waveform {
		lvlIndex := int(amp * float64(len(levels)-1))
		if lvlIndex < 1 && amp > 0.01 {
			lvlIndex = 1
		}
		if lvlIndex >= len(levels) {
			lvlIndex = len(levels) - 1
		}
		ch := string(levels[lvlIndex])

		if i == playheadBin {
			sb.WriteString(lipgloss.NewStyle().Foreground(warningColor).Bold(true).Render(ch))
		} else if playheadBin >= 0 && i < playheadBin {
			sb.WriteString(lipgloss.NewStyle().Foreground(activeTabColor).Render(ch))
		} else {
			sb.WriteString(lipgloss.NewStyle().Foreground(inactiveColor).Render(ch))
		}
	}

	title := lipgloss.NewStyle().Bold(true).Foreground(fgBrightColor).Render("Audio Waveform:")
	meter := fmt.Sprintf("[ %s ]", sb.String())
	return fmt.Sprintf("%s\n%s", title, meter)
}
