// Package main implements the fullscreen two-pane Project Hub layout for vo-studio-tui.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderProjectHub renders the fullscreen two-pane Project Hub launcher and manager.
func (m *AppModel) renderProjectHub() string {
	width := m.TermWidth
	height := m.TermHeight
	if width < 80 {
		width = 80
	}
	if height < 20 {
		height = 20
	}

	header := m.renderHubHeader(width)
	statusBar := m.renderHubStatusBar(width)

	headerHeight := lipgloss.Height(header)
	statusHeight := lipgloss.Height(statusBar)
	contentHeight := height - headerHeight - statusHeight
	if contentHeight < 10 {
		contentHeight = 10
	}

	leftWidth := int(float64(width) * 0.35)
	if leftWidth < 30 {
		leftWidth = 30
	}
	if leftWidth > 45 {
		leftWidth = 45
	}
	rightWidth := width - leftWidth - 4
	if rightWidth < 30 {
		rightWidth = 30
	}

	leftPane := m.renderHubProjectList(leftWidth, contentHeight)
	rightPane := m.renderHubProjectDetails(rightWidth, contentHeight)

	panes := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)
	baseView := lipgloss.JoinVertical(lipgloss.Left, header, panes, statusBar)

	if m.Mode == ModeDialog {
		return m.renderInputDialog(baseView)
	}

	return baseView
}

// renderHubHeader builds the top navigation header for the Project Hub.
func (m *AppModel) renderHubHeader(width int) string {
	title := lipgloss.NewStyle().Bold(true).Foreground(primaryColor).Render(" VOICESTUDIO PROJECT HUB ")
	subtitle := lipgloss.NewStyle().Foreground(inactiveColor).Render("Select a project or initialize a new voiceover workspace")

	projCount := 0
	if m.Registry != nil {
		projCount = len(m.Registry.Projects)
	}
	counter := lipgloss.NewStyle().Foreground(fgLightColor).Render(fmt.Sprintf("[%d Registered Projects] ", projCount))

	gap := width - lipgloss.Width(title) - lipgloss.Width(subtitle) - lipgloss.Width(counter) - 2
	if gap < 1 {
		gap = 1
	}

	row := lipgloss.JoinHorizontal(lipgloss.Center, title, subtitle, strings.Repeat(" ", gap), counter)
	return tabBarStyle.Width(width - 2).Render(row)
}

// renderHubProjectList renders the left pane containing selectable registered projects.
func (m *AppModel) renderHubProjectList(width, height int) string {
	title := lipgloss.NewStyle().Bold(true).Foreground(primaryColor).Render("PROJECTS")
	divider := lipgloss.NewStyle().Foreground(borderDimColor).Render(strings.Repeat("─", width-4))

	var lines []string
	lines = append(lines, title, divider)

	if m.Registry == nil || len(m.Registry.Projects) == 0 {
		lines = append(lines, "",
			lipgloss.NewStyle().Foreground(inactiveColor).Render("No registered projects."),
			"",
			lipgloss.NewStyle().Foreground(inactiveColor).Render("Press 'n' to create one,"),
			lipgloss.NewStyle().Foreground(inactiveColor).Render("or 'a' to register path."),
		)
	} else {
		for i, p := range m.Registry.Projects {
			isCur := (i == m.HubCursorIdx)
			timeStr := p.LastOpened.Format("2006-01-02 15:04")

			var nameLine string
			var metaLine string

			if isCur {
				nameLine = selectedItemStyle.Width(width - 4).Render(fmt.Sprintf("> %d. %s", i+1, p.Name))
				metaLine = selectedItemStyle.Width(width - 4).Render(fmt.Sprintf("     %s", timeStr))
			} else {
				nameLine = normalItemStyle.Width(width - 4).Render(fmt.Sprintf("  %d. %s", i+1, p.Name))
				metaLine = lipgloss.NewStyle().Foreground(inactiveColor).Width(width - 4).Render(fmt.Sprintf("     %s", timeStr))
			}

			lines = append(lines, nameLine, metaLine)
		}
	}

	return leftPaneStyle.Width(width).Height(height).Render(strings.Join(lines, "\n"))
}

// renderHubProjectDetails renders the right pane displaying rich metadata for the selected project.
func (m *AppModel) renderHubProjectDetails(width, height int) string {
	title := lipgloss.NewStyle().Bold(true).Foreground(primaryColor).Render("PROJECT DETAILS")
	divider := lipgloss.NewStyle().Foreground(borderDimColor).Render(strings.Repeat("─", width-4))

	var lines []string
	lines = append(lines, title, divider)

	if m.Registry == nil || len(m.Registry.Projects) == 0 || m.HubCursorIdx < 0 || m.HubCursorIdx >= len(m.Registry.Projects) {
		lines = append(lines, "",
			lipgloss.NewStyle().Foreground(inactiveColor).Render("No project selected."),
			"",
			lipgloss.NewStyle().Foreground(inactiveColor).Render("Create a new project with 'n', or register an existing one with 'a'."),
		)
		return rightPaneStyle.Width(width).Height(height).Render(strings.Join(lines, "\n"))
	}

	entry := m.Registry.Projects[m.HubCursorIdx]

	stat, err := os.Stat(entry.Path)
	if err != nil || !stat.IsDir() {
		lines = append(lines,
			lipgloss.NewStyle().Bold(true).Foreground(fgBrightColor).Render(entry.Name),
			"",
			fmt.Sprintf("%s %s", lipgloss.NewStyle().Foreground(inactiveColor).Render("Location:"), entry.Path),
			fmt.Sprintf("%s %s", lipgloss.NewStyle().Foreground(inactiveColor).Render("Status:  "), lipgloss.NewStyle().Bold(true).Foreground(errorColor).Render("[DIRECTORY NOT FOUND]")),
			"",
			lipgloss.NewStyle().Foreground(warningColor).Render("The project directory no longer exists on disk."),
			lipgloss.NewStyle().Foreground(inactiveColor).Render("Press 'd' to remove this entry from the registry."),
		)
		return rightPaneStyle.Width(width).Height(height).Render(strings.Join(lines, "\n"))
	}

	manifestPath := filepath.Join(entry.Path, "vo-project.json")
	if _, err := os.Stat(manifestPath); err != nil {
		lines = append(lines,
			lipgloss.NewStyle().Bold(true).Foreground(fgBrightColor).Render(entry.Name),
			"",
			fmt.Sprintf("%s %s", lipgloss.NewStyle().Foreground(inactiveColor).Render("Location:"), entry.Path),
			fmt.Sprintf("%s %s", lipgloss.NewStyle().Foreground(inactiveColor).Render("Status:  "), lipgloss.NewStyle().Bold(true).Foreground(warningColor).Render("[MISSING MANIFEST]")),
			"",
			lipgloss.NewStyle().Foreground(warningColor).Render("No vo-project.json manifest found in this directory."),
		)
		return rightPaneStyle.Width(width).Height(height).Render(strings.Join(lines, "\n"))
	}

	proj, err := LoadProject(entry.Path)
	if err != nil {
		lines = append(lines,
			lipgloss.NewStyle().Bold(true).Foreground(fgBrightColor).Render(entry.Name),
			"",
			lipgloss.NewStyle().Foreground(errorColor).Render("Error loading manifest: "+err.Error()),
		)
		return rightPaneStyle.Width(width).Height(height).Render(strings.Join(lines, "\n"))
	}

	// Header section
	headerName := lipgloss.NewStyle().Bold(true).Foreground(primaryColor).Render(proj.Name)
	statusBadge := lipgloss.NewStyle().Bold(true).Foreground(successColor).Render("[ON DISK]")
	timeStr := entry.LastOpened.Format("2006-01-02 15:04:05")

	lines = append(lines,
		fmt.Sprintf("%s  %s", headerName, statusBadge),
		fmt.Sprintf("%s %s", lipgloss.NewStyle().Foreground(inactiveColor).Render("Path:       "), entry.Path),
		fmt.Sprintf("%s %s", lipgloss.NewStyle().Foreground(inactiveColor).Render("Last Opened:"), timeStr),
		"",
	)

	// Summary stats
	totalFrames := 0
	renderedFrames := 0
	paddedFrames := 0
	silentFrames := 0
	for _, act := range proj.Acts {
		totalFrames += len(act.Frames)
		for _, f := range act.Frames {
			if f.IsSilent {
				silentFrames++
			}
			if f.Status == StatusRendered || f.Status == StatusPadded {
				renderedFrames++
			}
			if f.Status == StatusPadded {
				paddedFrames++
			}
		}
	}

	fullMasterStatus := lipgloss.NewStyle().Foreground(inactiveColor).Render("Not Assembled")
	fullMasterPath := filepath.Join(proj.RootPath, "audio", "full_production_master.wav")
	if fi, err := os.Stat(fullMasterPath); err == nil && fi.Size() > 1000 {
		fullMasterStatus = lipgloss.NewStyle().Foreground(successColor).Render("Ready (full_production_master.wav)")
	}

	actMastersCount := 0
	for _, act := range proj.Acts {
		if fi, err := os.Stat(act.MasterPath); err == nil && fi.Size() > 1000 {
			actMastersCount++
		}
	}

	sectionMetrics := lipgloss.NewStyle().Bold(true).Foreground(activeTabColor).Render("PRODUCTION METRICS")
	lines = append(lines,
		sectionMetrics,
		fmt.Sprintf("  • Acts:        %d narrative acts", len(proj.Acts)),
		fmt.Sprintf("  • Frames:      %d total frames (slot duration: %.1fs)", totalFrames, proj.DefaultSlotDuration),
		fmt.Sprintf("  • Audio State: %d/%d rendered, %d/%d padded (silent: %d)", renderedFrames, totalFrames, paddedFrames, totalFrames, silentFrames),
		fmt.Sprintf("  • Masters:     %d/%d act masters, Full Master: %s", actMastersCount, len(proj.Acts), fullMasterStatus),
		"",
	)

	// Voice Config
	sectionVoice := lipgloss.NewStyle().Bold(true).Foreground(activeTabColor).Render("VOICE STYLING & TOOLS")
	lines = append(lines,
		sectionVoice,
		fmt.Sprintf("  • Profile ID:   %s    Speed: %s    Language: %s", proj.Voice.ProfileID, proj.Voice.Speed, proj.Voice.Language),
		fmt.Sprintf("  • Voice Style:  %s", proj.Voice.Instruct),
		fmt.Sprintf("  • Audio Editor: %s", proj.ExternalAudioEditor),
		"",
	)

	// Narrative Acts breakdown
	sectionActs := lipgloss.NewStyle().Bold(true).Foreground(activeTabColor).Render("NARRATIVE BREAKDOWN")
	lines = append(lines, sectionActs)
	for _, act := range proj.Acts {
		lines = append(lines, fmt.Sprintf("  Act %02d: %-30s (%d frames)", act.ID, act.Title, len(act.Frames)))
	}

	return rightPaneStyle.Width(width).Height(height).Render(strings.Join(lines, "\n"))
}

// renderHubStatusBar creates the bottom status bar for the Project Hub.
func (m *AppModel) renderHubStatusBar(width int) string {
	badge := statusBadgeNormal.Render("HUB")
	actions := lipgloss.NewStyle().Foreground(fgBrightColor).Render(" [Enter] Open  [n] New Project  [a] Register Path  [d] Unregister  [q] Quit ")

	statusText := ""
	if m.StatusMsg != "" {
		statusText = toastStyle.Render("(" + m.StatusMsg + ")")
	}

	left := lipgloss.JoinHorizontal(lipgloss.Center, badge, actions)
	gap := width - lipgloss.Width(left) - lipgloss.Width(statusText) - 2
	if gap < 0 {
		gap = 0
	}

	content := lipgloss.JoinHorizontal(lipgloss.Center, left, strings.Repeat(" ", gap), statusText)
	return lipgloss.NewStyle().Width(width).Render(content)
}
