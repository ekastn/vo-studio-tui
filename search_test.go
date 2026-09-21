package main

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestSearchMode_FiltersFrames(t *testing.T) {
	proj := &Project{
		Acts: []*Act{
			{
				ID: 1,
				Frames: []*Frame{
					{ID: 1, Name: "Intro", Text: "Welcome everyone to the demo"},
					{ID: 2, Name: "Dilemma", Text: "Here is the big challenge"},
					{ID: 3, Name: "Demo", Text: "Let us show the live product"},
				},
			},
		},
	}

	m := NewAppModel(proj)
	m.TermWidth = 100
	m.TermHeight = 30

	// 1. Press '/' to enter search mode
	newM, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m = newM.(*AppModel)
	if m.Mode != ModeSearch {
		t.Fatalf("Expected ModeSearch after '/', got %v", m.Mode)
	}

	// 2. Type "challenge"
	for _, r := range "challenge" {
		newM, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = newM.(*AppModel)
	}

	if len(m.FilteredIdxs) != 1 || m.FilteredIdxs[0] != 2 {
		t.Fatalf("Expected filtered index [2], got %v", m.FilteredIdxs)
	}

	// 3. Press Esc to reset search
	newM, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = newM.(*AppModel)
	if m.Mode != ModeNormal {
		t.Errorf("Expected return to ModeNormal on Esc, got %v", m.Mode)
	}
	if len(m.FilteredIdxs) != 0 {
		t.Errorf("Expected FilteredIdxs to be cleared, got %v", m.FilteredIdxs)
	}
}
