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

func TestOpenDirectory_HubAndNormal(t *testing.T) {
	t.Setenv("FILE_MANAGER", "true")

	tempDir := t.TempDir()
	p1, err := CreateNewProject(tempDir, "Project Alpha", 10.0)
	if err != nil {
		t.Fatalf("Failed to create p1: %v", err)
	}

	reg := &Registry{}
	reg.Register(p1.Name, p1.RootPath)
	regPath := filepath.Join(tempDir, "registry.json")

	m := NewHubModel(reg, regPath)

	// In Hub mode, press 'o' to open highlighted project directory
	newM, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})
	m = newM.(*AppModel)
	if cmd == nil {
		t.Fatalf("Expected non-nil cmd for 'o' key in hub")
	}
	msg := cmd()
	openedMsg, ok := msg.(DirectoryOpenedMsg)
	if !ok || openedMsg.Err != nil || openedMsg.Path != p1.RootPath {
		t.Errorf("Expected DirectoryOpenedMsg for p1, got %+v", msg)
	}

	// Update with DirectoryOpenedMsg
	newM, _ = m.Update(openedMsg)
	m = newM.(*AppModel)
	expectedStatus := "Opened directory: " + filepath.Base(p1.RootPath)
	if m.StatusMsg != expectedStatus {
		t.Errorf("Expected status message '%s', got '%s'", expectedStatus, m.StatusMsg)
	}

	// Now enter project
	newM, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newM.(*AppModel)
	if m.Mode != ModeNormal {
		t.Fatalf("Expected ModeNormal, got %v", m.Mode)
	}

	// In Normal mode, press 'O' to open active project directory
	newM, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'O'}})
	m = newM.(*AppModel)
	if cmd == nil {
		t.Fatalf("Expected non-nil cmd for 'O' key in normal mode")
	}
	msg = cmd()
	openedMsg, ok = msg.(DirectoryOpenedMsg)
	if !ok || openedMsg.Err != nil || openedMsg.Path != p1.RootPath {
		t.Errorf("Expected DirectoryOpenedMsg for p1 root path, got %+v", msg)
	}

	// Test :dir command
	m.CmdInput.SetValue("dir")
	newM, cmd = m.executeCommand("dir")
	m = newM.(*AppModel)
	if cmd == nil {
		t.Fatalf("Expected non-nil cmd for :dir command")
	}
	msg = cmd()
	openedMsg, ok = msg.(DirectoryOpenedMsg)
	if !ok || openedMsg.Err != nil || openedMsg.Path != p1.RootPath {
		t.Errorf("Expected DirectoryOpenedMsg for p1 root path from :dir, got %+v", msg)
	}
}
