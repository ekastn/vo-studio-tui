package main

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func createTestModel() *AppModel {
	proj := &Project{
		Name:                "Test Production",
		DefaultSlotDuration: 10.0,
		Acts: []*Act{
			{
				ID:    1,
				Title: "Introduction",
				Frames: []*Frame{
					{
						ID:           1,
						Name:         "Hook",
						Text:         "Welcome to the production.",
						SlotDuration: 10.0,
						Status:       StatusRendered,
					},
					{
						ID:           2,
						Name:         "Problem",
						Text:         "Here is the dilemma.",
						SlotDuration: 10.0,
						Status:       StatusMissing,
					},
				},
			},
			{
				ID:    2,
				Title: "Resolution",
				Frames: []*Frame{
					{
						ID:           3,
						Name:         "Ending",
						Text:         "And so it concluded.",
						SlotDuration: 10.0,
						Status:       StatusMissing,
					},
				},
			},
		},
	}

	m := NewAppModel(proj)
	m.TermWidth = 100
	m.TermHeight = 30
	return m
}

func TestView_RendersTabsAndPanes(t *testing.T) {
	m := createTestModel()
	view := m.View()

	if !strings.Contains(view, "Introduction") {
		t.Errorf("View should contain active act title 'Introduction'")
	}
	if !strings.Contains(view, "Hook") && !strings.Contains(view, "Welcome") {
		t.Errorf("View should contain frame content or name")
	}
	if !strings.Contains(view, "NORMAL") {
		t.Errorf("View should contain NORMAL mode badge")
	}
}

func TestView_HubTwoPaneLayout(t *testing.T) {
	reg := &Registry{
		Projects: []ProjectEntry{
			{Name: "Alpha", Path: "/tmp/alpha"},
			{Name: "Beta", Path: "/tmp/beta"},
		},
	}
	m := NewHubModel(reg, "")
	m.TermWidth = 100
	m.TermHeight = 30

	view := m.View()
	if !strings.Contains(view, "VOICESTUDIO PROJECT HUB") {
		t.Errorf("Hub view should contain header title")
	}
	if !strings.Contains(view, "PROJECTS") {
		t.Errorf("Hub view should contain PROJECTS left pane header")
	}
	if !strings.Contains(view, "PROJECT DETAILS") {
		t.Errorf("Hub view should contain PROJECT DETAILS right pane header")
	}
	if !strings.Contains(view, "Alpha") {
		t.Errorf("Hub view should contain project name 'Alpha'")
	}
	if !strings.Contains(view, "HUB") {
		t.Errorf("Hub view should contain HUB status badge")
	}
}

func TestUpdate_Navigation(t *testing.T) {
	m := createTestModel()

	// Initial position: Act 0, Frame 0
	if m.ActiveActIdx != 0 || m.CursorIdx != 0 {
		t.Fatalf("Expected initial position Act 0, Frame 0, got Act %d, Frame %d", m.ActiveActIdx, m.CursorIdx)
	}

	// Move cursor down: 'j'
	newM, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = newM.(*AppModel)
	if m.CursorIdx != 1 {
		t.Errorf("Expected CursorIdx 1 after 'j', got %d", m.CursorIdx)
	}

	// Move cursor up: 'k'
	newM, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = newM.(*AppModel)
	if m.CursorIdx != 0 {
		t.Errorf("Expected CursorIdx 0 after 'k', got %d", m.CursorIdx)
	}

	// Switch act next: 'l'
	newM, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	m = newM.(*AppModel)
	if m.ActiveActIdx != 1 {
		t.Errorf("Expected ActiveActIdx 1 after 'l', got %d", m.ActiveActIdx)
	}

	// Switch act prev: 'h'
	newM, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	m = newM.(*AppModel)
	if m.ActiveActIdx != 0 {
		t.Errorf("Expected ActiveActIdx 0 after 'h', got %d", m.ActiveActIdx)
	}
}
