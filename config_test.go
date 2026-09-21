package main

import (
	"path/filepath"
	"testing"
)

func TestGlobalConfig_LoadSaveAndCascade(t *testing.T) {
	tempDir := t.TempDir()
	cfgPath := filepath.Join(tempDir, "config.json")

	cfg, err := LoadGlobalConfig(cfgPath)
	if err != nil {
		t.Fatalf("LoadGlobalConfig failed: %v", err)
	}

	cfg.APIURL = "http://localhost:9031/generate"
	cfg.APIToken = "tok_test_abc"
	cfg.DefaultSpeakerProfile = "prof_default"
	cfg.DefaultAudioEditor = "reaper"

	if err := SaveGlobalConfig(cfgPath, cfg); err != nil {
		t.Fatalf("SaveGlobalConfig failed: %v", err)
	}

	reloaded, err := LoadGlobalConfig(cfgPath)
	if err != nil {
		t.Fatalf("Failed to reload global config: %v", err)
	}

	if reloaded.APIURL != "http://localhost:9031/generate" {
		t.Errorf("Expected APIURL 'http://localhost:9031/generate', got '%s'", reloaded.APIURL)
	}
	if reloaded.APIToken != "tok_test_abc" {
		t.Errorf("Expected APIToken 'tok_test_abc', got '%s'", reloaded.APIToken)
	}

	// Test cascade: project with empty URL and Token inherits global config
	proj := &Project{
		Voice: VoiceConfig{
			ProfileID: "project_prof",
		},
	}
	proj.ApplyGlobalConfig(reloaded)

	if proj.Voice.APIURL != "http://localhost:9031/generate" {
		t.Errorf("Project failed to inherit APIURL")
	}
	if proj.Voice.APIToken != "tok_test_abc" {
		t.Errorf("Project failed to inherit APIToken")
	}
	if proj.Voice.ProfileID != "project_prof" {
		t.Errorf("Project profile should not be overwritten")
	}
	if proj.ExternalAudioEditor != "reaper" {
		t.Errorf("Expected inherited audio editor 'reaper', got '%s'", proj.ExternalAudioEditor)
	}
}

func TestGlobalConfig_NestedVoiceBlockCascade(t *testing.T) {
	globalCfg := &GlobalConfig{
		Voice: VoiceConfig{
			APIURL:    "http://100.96.85.54:9031/generate",
			APIToken:  "tok_nested_123",
			ProfileID: "7341026d",
			Language:  "Indonesian",
			Instruct:  "male, middle-aged",
			Speed:     "0.75",
		},
		DefaultAudioEditor: "audacity",
	}

	// New project with template placeholder values
	proj := &Project{
		Voice: VoiceConfig{
			ProfileID: "default",
			Language:  "English",
			Speed:     "1.0",
		},
	}

	proj.ApplyGlobalConfig(globalCfg)

	if proj.Voice.APIURL != "http://100.96.85.54:9031/generate" {
		t.Errorf("Expected cascaded APIURL, got '%s'", proj.Voice.APIURL)
	}
	if proj.Voice.APIToken != "tok_nested_123" {
		t.Errorf("Expected cascaded APIToken, got '%s'", proj.Voice.APIToken)
	}
	if proj.Voice.ProfileID != "7341026d" {
		t.Errorf("Expected cascaded ProfileID '7341026d', got '%s'", proj.Voice.ProfileID)
	}
	if proj.Voice.Language != "Indonesian" {
		t.Errorf("Expected cascaded Language 'Indonesian', got '%s'", proj.Voice.Language)
	}
	if proj.Voice.Instruct != "male, middle-aged" {
		t.Errorf("Expected cascaded Instruct 'male, middle-aged', got '%s'", proj.Voice.Instruct)
	}
	if proj.Voice.Speed != "0.75" {
		t.Errorf("Expected cascaded Speed '0.75', got '%s'", proj.Voice.Speed)
	}
	if proj.ExternalAudioEditor != "audacity" {
		t.Errorf("Expected cascaded audio editor 'audacity', got '%s'", proj.ExternalAudioEditor)
	}
}
