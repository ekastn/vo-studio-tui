// Package main implements modal overlays, help references, and Project Hub rendering for vo-studio-tui.
package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderHelpModal renders a centered reference modal for shortcuts and commands.
func (m *AppModel) renderHelpModal() string {
	helpWidth := 74
	if m.TermWidth > 0 && m.TermWidth < 78 {
		helpWidth = m.TermWidth - 4
	}

	header := lipgloss.NewStyle().Bold(true).Foreground(primaryColor).Render("VoiceStudio TUI Keyboard Reference")
	divider := lipgloss.NewStyle().Foreground(borderDimColor).Render(strings.Repeat("─", helpWidth-4))

	helpContent := `Navigation:
  j / k (or ↓ / ↑)       Move cursor down / up across frames
  g / G                  Jump to first / last frame in active Act
  h / l (or Tab/S-Tab)   Switch to previous / next Act
  1 – 9                  Jump directly to Act 1 through 9

Frame & Act Management:
  a                      Add a new frame to active Act
  d                      Delete highlighted frame (with confirmation)
  A                      Add a new Act
  D                      Delete active Act (with confirmation)
  J / K                  Reorder highlighted frame down / up

Audio & Narration:
  e                      Open frame script in $EDITOR (nvim)
  r / R                  Render highlighted frame / batch-render Act
  p / <Space>            Play raw frame audio via mpv
  P                      Play padded frame audio via mpv
  s                      Stop active audio playback immediately
  o                      Open frame audio in audio editor
  m / M                  Open Act master / Full master in audio editor

General:
  :                      Enter Command mode (:assemble, :speed, :q)
  /                      Search frames by text or ID
  c                      Open project configuration modal
  Ctrl+P                 Switch to Project Hub
  ?                      Toggle this Help overlay
  q / Ctrl+C             Exit application`

	box := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(primaryColor).
		Padding(1, 2).
		Width(helpWidth).
		Render(fmt.Sprintf("%s\n%s\n\n%s", header, divider, helpContent))

	return lipgloss.Place(m.TermWidth, m.TermHeight, lipgloss.Center, lipgloss.Center, box)
}

// renderInputDialog renders an input overlay prompt on top of the underlying view.
func (m *AppModel) renderInputDialog(underlying string) string {
	dialogWidth := 60
	if m.TermWidth > 0 && m.TermWidth < 64 {
		dialogWidth = m.TermWidth - 4
	}

	title := lipgloss.NewStyle().Bold(true).Foreground(warningColor).Render(m.DialogPrompt)
	inputView := m.DialogInput.View()
	hint := lipgloss.NewStyle().Foreground(inactiveColor).Render("[Enter] Confirm    [Esc] Cancel")

	content := fmt.Sprintf("%s\n\n%s\n\n%s", title, inputView, hint)
	dialogBox := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(warningColor).
		Padding(1, 2).
		Width(dialogWidth).
		Render(content)

	return lipgloss.Place(m.TermWidth, m.TermHeight, lipgloss.Center, lipgloss.Center, dialogBox)
}
