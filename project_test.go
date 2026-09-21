package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadProject_ValidManifest(t *testing.T) {
	tempDir := t.TempDir()

	// Create sample project manifest
	manifestJSON := `{
  "name": "Sample Production",
  "version": 1,
  "default_slot_duration": 10.0,
  "voice": {
    "profile_id": "profile_123",
    "language": "Indonesian",
    "instruct": "calm, clear",
    "speed": "0.75"
  },
  "acts": [
    {
      "id": 1,
      "title": "Act 1: Introduction",
      "frames": [
        {
          "id": 1,
          "name": "Opening",
          "script_file": "scripts/act_01/frame_01.txt",
          "slot_duration": 10.0,
          "is_silent": false
        },
        {
          "id": 2,
          "name": "Hook",
          "script_file": "scripts/act_01/frame_02.txt",
          "slot_duration": 10.0,
          "is_silent": false
        }
      ]
    }
  ]
}`

	manifestPath := filepath.Join(tempDir, "vo-project.json")
	if err := os.WriteFile(manifestPath, []byte(manifestJSON), 0644); err != nil {
		t.Fatalf("Failed to write manifest: %v", err)
	}

	// Create script files
	scriptDir := filepath.Join(tempDir, "scripts", "act_01")
	if err := os.MkdirAll(scriptDir, 0755); err != nil {
		t.Fatalf("Failed to create script dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(scriptDir, "frame_01.txt"), []byte("Hello world\n"), 0644); err != nil {
		t.Fatalf("Failed to write script: %v", err)
	}

	proj, err := LoadProject(tempDir)
	if err != nil {
		t.Fatalf("LoadProject returned error: %v", err)
	}

	if proj.Name != "Sample Production" {
		t.Errorf("Expected project name 'Sample Production', got '%s'", proj.Name)
	}
	if len(proj.Acts) != 1 {
		t.Fatalf("Expected 1 act, got %d", len(proj.Acts))
	}
	if len(proj.Acts[0].Frames) != 2 {
		t.Fatalf("Expected 2 frames in act 0, got %d", len(proj.Acts[0].Frames))
	}

	f1 := proj.Acts[0].Frames[0]
	if f1.Text != "Hello world" {
		t.Errorf("Expected frame 1 text 'Hello world', got '%s'", f1.Text)
	}
	if f1.SlotDuration != 10.0 {
		t.Errorf("Expected slot duration 10.0, got %.1f", f1.SlotDuration)
	}
}

func TestSaveProject(t *testing.T) {
	tempDir := t.TempDir()
	proj := &Project{
		RootPath:            tempDir,
		Name:                "Saved Project",
		Version:             1,
		DefaultSlotDuration: 10.0,
		Voice: VoiceConfig{
			ProfileID: "prof_1",
			Language:  "Indonesian",
			Speed:     "0.75",
		},
		Acts: []*Act{
			{
				ID:    1,
				Title: "Act 1",
				Frames: []*Frame{
					{
						ID:           1,
						Name:         "Scene 1",
						ScriptFile:   "scripts/act_01/frame_01.txt",
						SlotDuration: 10.0,
					},
				},
			},
		},
	}

	if err := SaveProject(proj); err != nil {
		t.Fatalf("SaveProject failed: %v", err)
	}

	reloaded, err := LoadProject(tempDir)
	if err != nil {
		t.Fatalf("LoadProject failed to read saved project: %v", err)
	}

	if reloaded.Name != "Saved Project" {
		t.Errorf("Expected name 'Saved Project', got '%s'", reloaded.Name)
	}
	if len(reloaded.Acts) != 1 || len(reloaded.Acts[0].Frames) != 1 {
		t.Fatalf("Unexpected act/frame count in reloaded project")
	}
}

func TestActAndFrameManagement(t *testing.T) {
	tempDir := t.TempDir()
	proj, err := CreateNewProject(tempDir, "Manage Test", 10.0)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Add second act
	act2, err := proj.AddAct("Act 2: Exploration")
	if err != nil {
		t.Fatalf("AddAct failed: %v", err)
	}
	if len(proj.Acts) != 2 {
		t.Fatalf("Expected 2 acts, got %d", len(proj.Acts))
	}
	if act2.ID != 2 {
		t.Errorf("Expected act ID 2, got %d", act2.ID)
	}

	// Add frame to act 2
	f2, err := act2.AddFrame(proj.RootPath, "Deep Dive", 12.0)
	if err != nil {
		t.Fatalf("AddFrame failed: %v", err)
	}
	if len(act2.Frames) != 1 {
		t.Fatalf("Expected 1 frame in act 2, got %d", len(act2.Frames))
	}
	if f2.SlotDuration != 12.0 {
		t.Errorf("Expected slot duration 12.0, got %.1f", f2.SlotDuration)
	}

	// Check script file was created
	if _, err := os.Stat(f2.ScriptPath); err != nil {
		t.Errorf("Expected script file to exist at %s", f2.ScriptPath)
	}

	// Add another frame and test reordering
	f3, err := act2.AddFrame(proj.RootPath, "Wrap Up", 8.0)
	if err != nil {
		t.Fatalf("AddFrame failed: %v", err)
	}
	if len(act2.Frames) != 2 {
		t.Fatalf("Expected 2 frames in act 2, got %d", len(act2.Frames))
	}

	// Move frame 1 (f3) up
	if err := act2.MoveFrame(1, -1); err != nil {
		t.Fatalf("MoveFrame failed: %v", err)
	}
	if act2.Frames[0].ID != f3.ID {
		t.Errorf("Expected frame %d to be at index 0 after move up, got %d", f3.ID, act2.Frames[0].ID)
	}

	// Delete frame
	if err := act2.DeleteFrame(f2.ID); err != nil {
		t.Fatalf("DeleteFrame failed: %v", err)
	}
	if len(act2.Frames) != 1 {
		t.Fatalf("Expected 1 frame in act 2 after deletion, got %d", len(act2.Frames))
	}

	// Rename act
	if err := proj.RenameAct(act2.ID, "Act 2: Advanced Exploration"); err != nil {
		t.Fatalf("RenameAct failed: %v", err)
	}
	if act2.Title != "Act 2: Advanced Exploration" {
		t.Errorf("Expected renamed title, got '%s'", act2.Title)
	}

	// Delete act
	if err := proj.DeleteAct(act2.ID); err != nil {
		t.Fatalf("DeleteAct failed: %v", err)
	}
	if len(proj.Acts) != 1 {
		t.Errorf("Expected 1 act after deletion, got %d", len(proj.Acts))
	}
}


