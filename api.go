// Package main implements the HTTP client communicating with the VoiceStudio synthesis API.
package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// RenderCompleteMsg is dispatched when a frame synthesis finishes.
type RenderCompleteMsg struct {
	FrameID  int
	Duration float64
	RawPath  string
	Err      error
}

// renderFrameAPICall sends a multipart/form-data POST request to the VoiceStudio API.
func renderFrameAPICall(cfg VoiceConfig, text, outPath string) error {
	if cfg.APIURL == "" {
		return fmt.Errorf("synthesis API URL is not configured (check ~/.config/vo-studio/config.json or :config)")
	}

	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	fields := map[string]string{
		"profile_id": cfg.ProfileID,
		"language":   cfg.Language,
		"instruct":   cfg.Instruct,
		"speed":      cfg.Speed,
		"text":       text,
	}

	for key, val := range fields {
		if val != "" {
			if err := w.WriteField(key, val); err != nil {
				return fmt.Errorf("failed to write multipart field %s: %w", key, err)
			}
		}
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("failed to close multipart writer: %w", err)
	}

	req, err := http.NewRequest("POST", cfg.APIURL, &b)
	if err != nil {
		return fmt.Errorf("failed to create synthesis request: %w", err)
	}

	if cfg.APIToken != "" {
		token := cfg.APIToken
		if !strings.HasPrefix(token, "Bearer ") {
			token = "Bearer " + token
		}
		req.Header.Set("Authorization", token)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	client := &http.Client{Timeout: 300 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("synthesis request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	audioBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read audio response body: %w", err)
	}
	if len(audioBytes) < 1000 {
		return fmt.Errorf("abnormally small audio payload (%d bytes)", len(audioBytes))
	}

	if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
		return fmt.Errorf("failed to create audio output directory: %w", err)
	}

	return os.WriteFile(outPath, audioBytes, 0644)
}

// renderFrameCmd dispatches a background task to synthesize audio for a frame.
func renderFrameCmd(f *Frame, cfg VoiceConfig) tea.Cmd {
	return func() tea.Msg {
		// Apply optional frame-level overrides
		effectiveCfg := cfg
		if f.VoiceOverride.ProfileID != "" {
			effectiveCfg.ProfileID = f.VoiceOverride.ProfileID
		}
		if f.VoiceOverride.Speed != "" {
			effectiveCfg.Speed = f.VoiceOverride.Speed
		}
		if f.VoiceOverride.Instruct != "" {
			effectiveCfg.Instruct = f.VoiceOverride.Instruct
		}

		if f.IsSilent {
			if err := generateSilenceCmd(f.PaddedPath, f.SlotDuration); err != nil {
				return RenderCompleteMsg{FrameID: f.ID, Err: err}
			}
			return RenderCompleteMsg{FrameID: f.ID, Duration: f.SlotDuration, RawPath: f.PaddedPath, Err: nil}
		}

		if err := renderFrameAPICall(effectiveCfg, f.Text, f.RawPath); err != nil {
			// Retry once after 2 seconds
			time.Sleep(2 * time.Second)
			if retryErr := renderFrameAPICall(effectiveCfg, f.Text, f.RawPath); retryErr != nil {
				return RenderCompleteMsg{FrameID: f.ID, Err: retryErr}
			}
		}

		if f.SlotDuration > 0 {
			if err := padFrameCmd(f.RawPath, f.PaddedPath, f.SlotDuration); err != nil {
				return RenderCompleteMsg{FrameID: f.ID, Err: fmt.Errorf("padding failed: %w", err)}
			}
		}

		dur := GetAudioDuration(f.RawPath)
		return RenderCompleteMsg{FrameID: f.ID, Duration: dur, RawPath: f.RawPath, Err: nil}
	}
}

