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
	APIURL                string `json:"api_url"`
	APIToken              string `json:"token"`
	DefaultSpeakerProfile string `json:"default_speaker_profile"`
	DefaultAudioEditor    string `json:"default_audio_editor"`
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

	if p.Voice.APIURL == "" {
		p.Voice.APIURL = cfg.APIURL
	}
	if p.Voice.APIToken == "" {
		p.Voice.APIToken = cfg.APIToken
	}
	if p.Voice.ProfileID == "" && cfg.DefaultSpeakerProfile != "" {
		p.Voice.ProfileID = cfg.DefaultSpeakerProfile
	}
	if p.ExternalAudioEditor == "" {
		if cfg.DefaultAudioEditor != "" {
			p.ExternalAudioEditor = cfg.DefaultAudioEditor
		} else {
			p.ExternalAudioEditor = "audacity"
		}
	}
}
