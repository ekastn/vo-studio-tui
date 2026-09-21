package main

import (
	"os"
	"testing"
)


func TestSaveFrameScriptAndReloadScripts(t *testing.T) {
	tempDir := t.TempDir()
	proj, err := CreateNewProject(tempDir, "Script Test", 10.0)
	if err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	act := proj.Acts[0]
	f := act.Frames[0]

	// 1. Save frame script
	newContent := "Updated narration script for frame one."
	if err := SaveFrameScript(f, newContent); err != nil {
		t.Fatalf("SaveFrameScript failed: %v", err)
	}
	if f.Text != newContent {
		t.Errorf("Expected frame text '%s', got '%s'", newContent, f.Text)
	}

	// Verify file on disk
	diskBytes, err := os.ReadFile(f.ScriptPath)
	if err != nil {
		t.Fatalf("Failed to read script file: %v", err)
	}
	if string(diskBytes) != newContent+"\n" {
		t.Errorf("Disk file mismatch: expected '%s\n', got '%s'", newContent, string(diskBytes))
	}

	// 2. Modify file externally on disk
	externalContent := "External edit via another tool."
	if err := os.WriteFile(f.ScriptPath, []byte(externalContent+"\n"), 0644); err != nil {
		t.Fatalf("Failed to write external content: %v", err)
	}

	// 3. Reload scripts
	if err := proj.ReloadScripts(); err != nil {
		t.Fatalf("ReloadScripts failed: %v", err)
	}
	if f.Text != externalContent {
		t.Errorf("Expected reloaded frame text '%s', got '%s'", externalContent, f.Text)
	}
}

func TestScriptStats_WordsAndChars(t *testing.T) {
	text := "This is a five word sentence."
	words, chars := computeScriptStats(text)
	if words != 6 {
		t.Errorf("Expected 6 words, got %d", words)
	}
	if chars != len(text) {
		t.Errorf("Expected %d chars, got %d", len(text), chars)
	}
}
