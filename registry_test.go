package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRegistry_AddAndRemove(t *testing.T) {
	tempDir := t.TempDir()
	regPath := filepath.Join(tempDir, "registry.json")

	reg, err := LoadRegistry(regPath)
	if err != nil {
		t.Fatalf("LoadRegistry failed: %v", err)
	}

	if len(reg.Projects) != 0 {
		t.Errorf("Expected 0 projects initially, got %d", len(reg.Projects))
	}

	// Register a project
	projPath := filepath.Join(tempDir, "proj1")
	reg.Register("Project One", projPath)

	if len(reg.Projects) != 1 {
		t.Fatalf("Expected 1 project after register, got %d", len(reg.Projects))
	}
	if reg.Projects[0].Name != "Project One" {
		t.Errorf("Expected project name 'Project One', got '%s'", reg.Projects[0].Name)
	}

	// Save and reload
	if err := SaveRegistry(regPath, reg); err != nil {
		t.Fatalf("SaveRegistry failed: %v", err)
	}

	reloaded, err := LoadRegistry(regPath)
	if err != nil {
		t.Fatalf("Failed to reload registry: %v", err)
	}
	if len(reloaded.Projects) != 1 {
		t.Fatalf("Expected 1 project in reloaded registry, got %d", len(reloaded.Projects))
	}

	// Unregister
	reloaded.Unregister(projPath)
	if len(reloaded.Projects) != 0 {
		t.Errorf("Expected 0 projects after unregister, got %d", len(reloaded.Projects))
	}
}

func TestCreateNewProject_Scaffolding(t *testing.T) {
	tempDir := t.TempDir()
	proj, err := CreateNewProject(tempDir, "Test Scaffold", 10.0)
	if err != nil {
		t.Fatalf("CreateNewProject failed: %v", err)
	}

	if proj.Name != "Test Scaffold" {
		t.Errorf("Expected project name 'Test Scaffold', got '%s'", proj.Name)
	}

	manifestPath := filepath.Join(proj.RootPath, "vo-project.json")
	if _, err := os.Stat(manifestPath); err != nil {
		t.Errorf("Expected vo-project.json to exist at %s", manifestPath)
	}

	scriptPath := filepath.Join(proj.RootPath, "scripts", "act_01", "frame_01.txt")
	if _, err := os.Stat(scriptPath); err != nil {
		t.Errorf("Expected initial frame script to exist at %s", scriptPath)
	}
}
