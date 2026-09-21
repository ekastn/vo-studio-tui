// Package main implements global user configuration and credential cascading for vo-studio-tui.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// GlobalConfig holds user-wide connection parameters and tool defaults.
type GlobalConfig struct {
	APIURL                string      `json:"api_url"`
	APIToken              string      `json:"token"`
	DefaultSpeakerProfile string      `json:"default_speaker_profile,omitempty"`
	DefaultAudioEditor    string      `json:"default_audio_editor,omitempty"`
	Voice                 VoiceConfig `json:"voice,omitempty"`
}

// DefaultGlobalConfig returns recommended starting configuration values.
func DefaultGlobalConfig() *GlobalConfig {
	return &GlobalConfig{
		APIURL:                "http://localhost:9031/generate",
		APIToken:              "",
		DefaultSpeakerProfile: "default",
		DefaultAudioEditor:    "audacity",
	}
}

// LoadGlobalConfig reads the user config file, creating a default one if missing.
func LoadGlobalConfig(configPath string) (*GlobalConfig, error) {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		cfg := DefaultGlobalConfig()
		_ = SaveGlobalConfig(configPath, cfg)
		return cfg, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read global config: %w", err)
	}

	var cfg GlobalConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse global config: %w", err)
	}
	return &cfg, nil
}

// SaveGlobalConfig serializes the global config to disk.
func SaveGlobalConfig(configPath string, cfg *GlobalConfig) error {
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode global config: %w", err)
	}
	return os.WriteFile(configPath, data, 0644)
}

// ApplyGlobalConfig cascades global credentials and defaults into the project if not explicitly set.
func (p *Project) ApplyGlobalConfig(cfg *GlobalConfig) {
	if cfg == nil {
		return
	}

	apiURL := cfg.APIURL
	if apiURL == "" && cfg.Voice.APIURL != "" {
		apiURL = cfg.Voice.APIURL
	}
	if p.Voice.APIURL == "" {
		p.Voice.APIURL = apiURL
	}

	token := cfg.APIToken
	if token == "" && cfg.Voice.APIToken != "" {
		token = cfg.Voice.APIToken
	}
	if p.Voice.APIToken == "" {
		p.Voice.APIToken = token
	}

	profile := cfg.DefaultSpeakerProfile
	if profile == "" && cfg.Voice.ProfileID != "" {
		profile = cfg.Voice.ProfileID
	}
	if (p.Voice.ProfileID == "" || p.Voice.ProfileID == "default") && profile != "" {
		p.Voice.ProfileID = profile
	}

	speed := cfg.Voice.Speed
	if (p.Voice.Speed == "" || p.Voice.Speed == "1.0") && speed != "" {
		p.Voice.Speed = speed
	}

	language := cfg.Voice.Language
	if (p.Voice.Language == "" || p.Voice.Language == "English") && language != "" {
		p.Voice.Language = language
	}

	instruct := cfg.Voice.Instruct
	if p.Voice.Instruct == "" && instruct != "" {
		p.Voice.Instruct = instruct
	}

	if p.ExternalAudioEditor == "" {
		if cfg.DefaultAudioEditor != "" {
			p.ExternalAudioEditor = cfg.DefaultAudioEditor
		} else {
			p.ExternalAudioEditor = "audacity"
		}
	}
}
