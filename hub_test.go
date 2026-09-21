package main

import (
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestHub_NavigationAndSelection(t *testing.T) {
	tempDir := t.TempDir()

	// Scaffold two test projects
	p1, err := CreateNewProject(tempDir, "Project Alpha", 10.0)
	if err != nil {
		t.Fatalf("Failed to create p1: %v", err)
	}
	p2, err := CreateNewProject(tempDir, "Project Beta", 10.0)
	if err != nil {
		t.Fatalf("Failed to create p2: %v", err)
	}

	reg := &Registry{}
	reg.Register(p1.Name, p1.RootPath)
	reg.Register(p2.Name, p2.RootPath)
	regPath := filepath.Join(tempDir, "registry.json")

	m := NewHubModel(reg, regPath)
	m.GlobalConfig = &GlobalConfig{
		APIURL:   "http://test-server:9031/generate",
		APIToken: "tok_hub_test",
	}
	m.TermWidth = 100
	m.TermHeight = 30

	if m.Mode != ModeHub {
		t.Fatalf("Expected initial mode ModeHub, got %v", m.Mode)
	}
	if len(m.Registry.Projects) != 2 {
		t.Fatalf("Expected 2 projects in hub, got %d", len(m.Registry.Projects))
	}

	// Move cursor down
	newM, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = newM.(*AppModel)
	if m.HubCursorIdx != 1 {
		t.Errorf("Expected HubCursorIdx 1, got %d", m.HubCursorIdx)
	}

	// Press enter to open selected project
	newM, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newM.(*AppModel)
	if m.Mode != ModeNormal {
		t.Errorf("Expected ModeNormal after Enter, got %v", m.Mode)
	}
	if m.Project == nil || m.Project.Name != "Project Beta" {
		t.Errorf("Expected Project Beta opened, got %v", m.Project)
	}
	if m.Project.Voice.APIURL != "http://test-server:9031/generate" {
		t.Errorf("Expected cascaded APIURL, got '%s'", m.Project.Voice.APIURL)
	}
	if m.Project.Voice.APIToken != "tok_hub_test" {
		t.Errorf("Expected cascaded APIToken, got '%s'", m.Project.Voice.APIToken)
	}

	// Press Ctrl+P to return to hub
	newM, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlP})
	m = newM.(*AppModel)
	if m.Mode != ModeHub {
		t.Errorf("Expected ModeHub after Ctrl+P, got %v", m.Mode)
	}
}
