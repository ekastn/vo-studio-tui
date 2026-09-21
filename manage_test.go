package main

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestInTUI_ActAndFrameManagement(t *testing.T) {
	tempDir := t.TempDir()
	proj, err := CreateNewProject(tempDir, "TUI Manage Test", 10.0)
	if err != nil {
		t.Fatalf("CreateNewProject failed: %v", err)
	}

	m := NewAppModel(proj)
	m.TermWidth = 100
	m.TermHeight = 30

	// 1. Press 'a' to add a frame
	newM, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m = newM.(*AppModel)
	if m.Mode != ModeDialog || m.DialogAction != "add_frame" {
		t.Fatalf("Expected ModeDialog with action add_frame, got %v / %s", m.Mode, m.DialogAction)
	}

	// Type frame title and press Enter
	m.DialogInput.SetValue("Scene Two")
	newM, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newM.(*AppModel)
	if m.Mode != ModeNormal {
		t.Fatalf("Expected return to ModeNormal after dialog, got %v", m.Mode)
	}
	act := m.CurrentAct()
	if len(act.Frames) != 2 {
		t.Fatalf("Expected 2 frames in Act, got %d", len(act.Frames))
	}
	if act.Frames[1].Name != "Scene Two" {
		t.Errorf("Expected frame name 'Scene Two', got '%s'", act.Frames[1].Name)
	}

	// 2. Press 'J' to move frame 0 down
	m.CursorIdx = 0
	newM, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'J'}})
	m = newM.(*AppModel)
	if act.Frames[0].Name != "Scene Two" {
		t.Errorf("Expected 'Scene Two' to be first after J, got '%s'", act.Frames[0].Name)
	}

	// 3. Press 'd' to delete current frame
	newM, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	m = newM.(*AppModel)
	if m.Mode != ModeDialog || m.DialogAction != "confirm_delete_frame" {
		t.Fatalf("Expected confirm_delete_frame dialog, got %v / %s", m.Mode, m.DialogAction)
	}
	m.DialogInput.SetValue("yes")
	newM, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newM.(*AppModel)
	if len(act.Frames) != 1 {
		t.Errorf("Expected 1 frame after deletion, got %d", len(act.Frames))
	}

	// 4. Press 'A' to add an Act
	newM, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'A'}})
	m = newM.(*AppModel)
	if m.Mode != ModeDialog || m.DialogAction != "add_act" {
		t.Fatalf("Expected add_act dialog, got %v / %s", m.Mode, m.DialogAction)
	}
	m.DialogInput.SetValue("Act Two")
	newM, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newM.(*AppModel)
	if len(m.Project.Acts) != 2 {
		t.Errorf("Expected 2 acts after adding act, got %d", len(m.Project.Acts))
	}
}
