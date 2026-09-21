// Package main implements real-time search filtering for vo-studio-tui.
package main

import (
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// handleSearchKey processes keystrokes in search filter mode (/).
func (m *AppModel) handleSearchKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.Mode = ModeNormal
		m.SearchInput.Blur()
		m.FilteredIdxs = nil
		m.SearchQuery = ""
		return m, nil

	case "enter":
		m.Mode = ModeNormal
		m.SearchInput.Blur()
		return m, nil
	}

	var cmd tea.Cmd
	m.SearchInput, cmd = m.SearchInput.Update(msg)
	query := strings.ToLower(strings.TrimSpace(m.SearchInput.Value()))
	m.SearchQuery = query

	act := m.CurrentAct()
	if act != nil && query != "" {
		m.FilteredIdxs = nil
		for _, f := range act.Frames {
			if strings.Contains(strings.ToLower(f.Text), query) ||
				strings.Contains(strings.ToLower(f.Name), query) ||
				strings.Contains(strconv.Itoa(f.ID), query) {
				m.FilteredIdxs = append(m.FilteredIdxs, f.ID)
			}
		}
		m.CursorIdx = 0
	} else {
		m.FilteredIdxs = nil
	}

	return m, cmd
}
