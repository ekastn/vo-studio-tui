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
