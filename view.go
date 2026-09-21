// Package main implements terminal user interface rendering for vo-studio-tui.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)


var (
	// Gruvbox dark color palette (preserves transparent host background)
	primaryColor   = lipgloss.Color("#fabd2f") // Warm yellow
	secondaryColor = lipgloss.Color("#504945") // Dark accent
	activeTabColor = lipgloss.Color("#8ec07c") // Aqua
	inactiveColor  = lipgloss.Color("#928374") // Gray
	successColor   = lipgloss.Color("#b8bb26") // Green
	warningColor   = lipgloss.Color("#fe8019") // Orange
	errorColor     = lipgloss.Color("#fb4934") // Red
	borderDimColor = lipgloss.Color("#504945") // Border
	fgLightColor   = lipgloss.Color("#ebdbb2") // Foreground
	fgBrightColor  = lipgloss.Color("#fbf1c7") // Bright foreground
	bgSelectColor  = lipgloss.Color("#3c3836") // Selection background

	tabActiveStyle    = lipgloss.NewStyle().Bold(true).Foreground(fgBrightColor).Background(secondaryColor).Padding(0, 1)
	tabInactiveStyle  = lipgloss.NewStyle().Foreground(inactiveColor).Padding(0, 1)
	tabBarStyle       = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, false, true, false).BorderForeground(borderDimColor).MarginBottom(1)
	leftPaneStyle     = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(borderDimColor).Padding(0, 1)
	rightPaneStyle    = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(borderDimColor).Padding(0, 1)
	selectedItemStyle = lipgloss.NewStyle().Bold(true).Foreground(fgBrightColor).Background(bgSelectColor)
	normalItemStyle   = lipgloss.NewStyle().Foreground(fgLightColor)

	statusBadgeNormal  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#282828")).Background(activeTabColor).Padding(0, 1)
	statusBadgeCommand = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#282828")).Background(warningColor).Padding(0, 1)
	statusBadgeSearch  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#282828")).Background(primaryColor).Padding(0, 1)
	statusBadgeDialog  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#282828")).Background(warningColor).Padding(0, 1)
	toastStyle         = lipgloss.NewStyle().Foreground(fgLightColor).Padding(0, 1)
)

// View renders the root interface based on the active mode and model state.
func (m *AppModel) View() string {
	if m.TermWidth == 0 || m.TermHeight == 0 {
		return "Initializing vo-studio-tui..."
	}

	if m.ShowHelp {
		return m.renderHelpModal()
	}

	if m.Mode == ModeHub {
		return m.renderProjectHub()
	}

	tabs := m.renderTabBar()
	panes := m.renderPanes()
	statusBar := m.renderStatusBar()

	baseView := lipgloss.JoinVertical(lipgloss.Left, tabs, panes, statusBar)

	if m.Mode == ModeDialog {
		return m.renderInputDialog(baseView)
	}

	return baseView
}


// renderTabBar creates the horizontal Act tab strip across the top.
func (m *AppModel) renderTabBar() string {
	if m.Project == nil || len(m.Project.Acts) == 0 {
		return tabBarStyle.Width(m.TermWidth - 2).Render(" No Acts (Press 'A' to create an Act)")
	}

	var tabs []string
	isNarrow := m.TermWidth > 0 && m.TermWidth < 95

	for i, act := range m.Project.Acts {
		var label string
		if isNarrow {
			if i == m.ActiveActIdx {
				label = fmt.Sprintf("[%d: %s]", act.ID, act.Title)
			} else {
				label = fmt.Sprintf(" %d ", act.ID)
			}
		} else {
			if i == m.ActiveActIdx {
				label = fmt.Sprintf("%d. %s (%d frames)", act.ID, act.Title, len(act.Frames))
			} else {
				label = fmt.Sprintf("%d. %s", act.ID, act.Title)
			}
		}

		if i == m.ActiveActIdx {
			tabs = append(tabs, tabActiveStyle.Render(label))
		} else {
			tabs = append(tabs, tabInactiveStyle.Render(label))
		}
	}

	content := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
	return tabBarStyle.Width(m.TermWidth - 2).Render(content)
}

// renderPanes renders the split-screen layout (left frame list + right detail pane).
func (m *AppModel) renderPanes() string {
	leftWidth := 42
	if m.TermWidth > 120 {
		leftWidth = 46
	}
	rightWidth := m.TermWidth - leftWidth - 4
	if rightWidth < 30 {
		rightWidth = 30
	}

	contentHeight := m.TermHeight - 6
	if contentHeight < 10 {
		contentHeight = 10
	}

	left := m.renderFrameList(leftWidth, contentHeight)
	right := m.renderDetailPane(rightWidth, contentHeight)

	return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
}

// renderFrameList renders the list of frames for the active act.
func (m *AppModel) renderFrameList(width, height int) string {
	act := m.CurrentAct()
	if act == nil || len(act.Frames) == 0 {
		return leftPaneStyle.Width(width).Height(height).Render("No frames in this Act.\nPress 'a' to add a frame.")
	}

	var lines []string
	frames := act.Frames
	if len(m.FilteredIdxs) > 0 {
		frames = make([]*Frame, 0, len(m.FilteredIdxs))
		for _, id := range m.FilteredIdxs {
			for _, f := range act.Frames {
				if f.ID == id {
					frames = append(frames, f)
					break
				}
			}
		}
	}

	maxVisible := height - 2
	startIdx := 0
	if m.CursorIdx >= maxVisible {
		startIdx = m.CursorIdx - maxVisible + 1
	}
	endIdx := startIdx + maxVisible
	if endIdx > len(frames) {
		endIdx = len(frames)
	}

	for i := startIdx; i < endIdx; i++ {
		f := frames[i]
		isCur := (i == m.CursorIdx)
		isPlaying := m.IsPlaying && m.PlayingFrameID == f.ID
		isJob := m.ActiveJobs[f.ID]

		badge := formatFrameBadge(f, isJob, m.Spinner.View(), isPlaying)
		preview := f.Text
		if preview == "" && f.Name != "" {
			preview = f.Name
		}
		if preview == "" {
			preview = "(empty script)"
		}

		snipLen := width - 20
		if snipLen < 8 {
			snipLen = 8
		}
		if len(preview) > snipLen {
			preview = preview[:snipLen-3] + "..."
		}

		line := fmt.Sprintf(" %02d %-14s %s", f.ID, preview, badge)
		if isCur {
			line = selectedItemStyle.Width(width - 2).Render("> " + strings.TrimSpace(line))
		} else {
			line = normalItemStyle.Width(width - 2).Render("  " + strings.TrimSpace(line))
		}
		lines = append(lines, line)
	}

	for len(lines) < maxVisible {
		lines = append(lines, "")
	}

	title := fmt.Sprintf("Act %d Frames (%d)", act.ID, len(frames))
	content := fmt.Sprintf("%s\n%s", lipgloss.NewStyle().Bold(true).Foreground(activeTabColor).Render(title), strings.Join(lines, "\n"))
	return leftPaneStyle.Width(width).Height(height).Render(content)
}

// formatFrameBadge produces a status indicator string for a frame.
func formatFrameBadge(f *Frame, isJob bool, spinner string, isPlaying bool) string {
	if isJob {
		return lipgloss.NewStyle().Foreground(warningColor).Render(spinner + " [Rendering]")
	}
	if isPlaying {
		return lipgloss.NewStyle().Foreground(primaryColor).Bold(true).Render("▶ [Playing]")
	}
	switch f.Status {
	case StatusPadded:
		if f.RawDuration > 0 {
			return lipgloss.NewStyle().Foreground(successColor).Render(fmt.Sprintf("[%.2fs]", f.RawDuration))
		}
		return lipgloss.NewStyle().Foreground(successColor).Render("[Padded]")
	case StatusRendered:
		if f.RawDuration > 0 {
			return lipgloss.NewStyle().Foreground(activeTabColor).Render(fmt.Sprintf("[%.2fs]", f.RawDuration))
		}
		return lipgloss.NewStyle().Foreground(activeTabColor).Render("[Rendered]")
	case StatusRendering:
		return lipgloss.NewStyle().Foreground(warningColor).Render("[Rendering]")
	case StatusError:
		return lipgloss.NewStyle().Foreground(errorColor).Render("[Error]")
	default:
		return lipgloss.NewStyle().Foreground(inactiveColor).Render("[Missing]")
	}
}

// renderDetailPane renders the inspector panel for the highlighted frame.
func (m *AppModel) renderDetailPane(width, height int) string {
	f := m.CurrentFrame()
	if f == nil {
		return rightPaneStyle.Width(width).Height(height).Render("No frame selected.\nPress 'a' to add a frame.")
	}

	act := m.CurrentAct()
	title := fmt.Sprintf("Frame %02d Detail – %s", f.ID, act.Title)
	header := lipgloss.NewStyle().Bold(true).Foreground(primaryColor).Render(title)

	// Timeline slot timing
	startSec := 0.0
	for _, prevF := range act.Frames {
		if prevF.ID == f.ID {
			break
		}
		startSec += prevF.SlotDuration
	}
	endSec := startSec + f.SlotDuration

	var timeline string
	if f.SlotDuration > 0 {
		timeline = fmt.Sprintf("Timeline Slot : %02d:%02d – %02d:%02d (Slot: %.1fs)", int(startSec)/60, int(startSec)%60, int(endSec)/60, int(endSec)%60, f.SlotDuration)
	} else {
		timeline = fmt.Sprintf("Timeline Slot : %02d:%02d (Natural / unpadded)", int(startSec)/60, int(startSec)%60)
	}

	// Audio duration & headroom
	var durStatus string
	if f.RawDuration > 0 {
		if f.SlotDuration > 0 {
			headroom := f.SlotDuration - f.RawDuration
			if headroom < 0 {
				durStatus = fmt.Sprintf("Audio Duration: %.2fs (%s – exceeds slot)", f.RawDuration, lipgloss.NewStyle().Foreground(errorColor).Render("OVER"))
			} else {
				durStatus = fmt.Sprintf("Audio Duration: %.2fs (Headroom: %.2fs)", f.RawDuration, headroom)
			}
		} else {
			durStatus = fmt.Sprintf("Audio Duration: %.2fs", f.RawDuration)
		}
	} else {
		durStatus = "Audio Duration: None (Not Rendered)"
	}

	// Status line
	statusColor := inactiveColor
	statusText := f.Status.String()
	if m.ActiveJobs[f.ID] {
		statusColor = warningColor
		statusText = m.Spinner.View() + " Synthesizing Voiceover..."
	} else if m.IsPlaying && m.PlayingFrameID == f.ID {
		statusColor = primaryColor
		statusText = fmt.Sprintf("▶ Playing (%s)", m.PlayingFile)
	} else {
		switch f.Status {
		case StatusPadded:
			statusColor = successColor
		case StatusRendered:
			statusColor = activeTabColor
		case StatusError:
			statusColor = errorColor
		case StatusRendering:
			statusColor = warningColor
		}
	}
	statusLine := fmt.Sprintf("Status        : %s", lipgloss.NewStyle().Foreground(statusColor).Bold(true).Render(statusText))

	infoBox := fmt.Sprintf("%s\n%s\n%s\n%s", header, timeline, durStatus, statusLine)

	// Synthesis diagnostics card if rendering or errored
	var diagBox string
	if m.ActiveJobs[f.ID] && m.Project != nil {
		diagHdr := lipgloss.NewStyle().Bold(true).Foreground(warningColor).Render(m.Spinner.View() + " VoiceStudio Synthesis in Progress:")
		voiceCfg := m.Project.Voice
		if f.VoiceOverride.ProfileID != "" {
			voiceCfg.ProfileID = f.VoiceOverride.ProfileID
		}
		if f.VoiceOverride.Speed != "" {
			voiceCfg.Speed = f.VoiceOverride.Speed
		}
		if f.VoiceOverride.Instruct != "" {
			voiceCfg.Instruct = f.VoiceOverride.Instruct
		}
		stageText := fmt.Sprintf("Sending request -> Receiving raw WAV -> Padding to %.1fs", f.SlotDuration)
		if f.SlotDuration <= 0 {
			stageText = "Sending request -> Receiving raw WAV (natural duration)"
		}
		diagContent := fmt.Sprintf("Engine : VoiceStudio API (%s)\nProfile: %s (%s, %s)\nSpeed  : %s\nStage  : %s",
			voiceCfg.APIURL, voiceCfg.ProfileID, voiceCfg.Language, voiceCfg.Instruct, voiceCfg.Speed, stageText)
		diagBox = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(warningColor).
			Padding(0, 1).
			Width(width - 4).
			Render(fmt.Sprintf("%s\n%s", diagHdr, diagContent))
	} else if f.Status == StatusError && f.ErrorMsg != "" {
		diagHdr := lipgloss.NewStyle().Bold(true).Foreground(errorColor).Render("Synthesis Error:")
		diagBox = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(errorColor).
			Padding(0, 1).
			Width(width - 4).
			Render(fmt.Sprintf("%s\n%s", diagHdr, f.ErrorMsg))
	}

	// Waveform extraction and rendering
	bins := width - 8
	if bins < 16 {
		bins = 16
	}
	if bins > 50 {
		bins = 50
	}

	targetAudio := f.RawPath
	if _, err := os.Stat(targetAudio); err != nil {
		targetAudio = f.PaddedPath
	}

	if len(f.Waveform) != bins {
		if _, err := os.Stat(targetAudio); err == nil {
			if wf, err := ExtractWaveform(targetAudio, bins); err == nil {
				f.Waveform = wf
			}
		}
	}

	progress := -1.0
	if m.IsPlaying && m.PlayDuration > 0 && (m.PlayingFile == filepath.Base(f.RawPath) || m.PlayingFile == filepath.Base(f.PaddedPath)) {
		progress = time.Since(m.PlayStartTime).Seconds() / m.PlayDuration
		if progress > 1.0 {
			progress = 1.0
		}
	}

	waveformBox := RenderWaveform(f.Waveform, progress, width)

	// Script preview with statistics
	words, chars := computeScriptStats(f.Text)
	titleText := "Narration Script:"
	if f.Text != "" {
		titleText = fmt.Sprintf("Narration Script (%d words, %d chars):", words, chars)
	}
	scriptTitle := lipgloss.NewStyle().Bold(true).Foreground(fgBrightColor).Render(titleText)
	scriptBody := f.Text

	if scriptBody == "" {
		scriptBody = "(No narration text – press 'e' to edit in $EDITOR)"
	}
	wrappedScript := lipgloss.NewStyle().Width(width - 4).Foreground(fgLightColor).Padding(0, 1).Render(scriptBody)

	// File path info
	fileInfo := fmt.Sprintf("Script : %s\nAudio  : %s", f.ScriptFile, f.PaddedPath)
	fileBox := lipgloss.NewStyle().Foreground(inactiveColor).Render(fileInfo)

	var items []string
	items = append(items, infoBox)
	if diagBox != "" {
		items = append(items, "", diagBox)
	}
	if waveformBox != "" {
		items = append(items, "", waveformBox)
	}
	items = append(items, "", scriptTitle, wrappedScript, "", fileBox)
	content := lipgloss.JoinVertical(lipgloss.Left, items...)
	return rightPaneStyle.Width(width).Height(height).Render(content)
}


// renderStatusBar renders the modal prompt and system status toast.
func (m *AppModel) renderStatusBar() string {
	var badge string
	switch m.Mode {
	case ModeNormal:
		badge = statusBadgeNormal.Render(" NORMAL ")
	case ModeCommand:
		badge = statusBadgeCommand.Render(" COMMAND ")
	case ModeSearch:
		badge = statusBadgeSearch.Render(" SEARCH ")
	case ModeDialog:
		badge = statusBadgeDialog.Render(" INPUT ")
	default:
		badge = statusBadgeNormal.Render(" NORMAL ")
	}

	var leftText string
	switch m.Mode {
	case ModeCommand:
		leftText = lipgloss.JoinHorizontal(lipgloss.Center, badge, " ", m.CmdInput.View())
	case ModeSearch:
		leftText = lipgloss.JoinHorizontal(lipgloss.Center, badge, " ", m.SearchInput.View())
	default:
		fileName := "no file"
		f := m.CurrentFrame()
		if f != nil && f.ScriptFile != "" {
			fileName = filepath.Base(f.ScriptFile)
		}
		fileStyled := lipgloss.NewStyle().Foreground(fgLightColor).Render(fileName)
		helpStyled := lipgloss.NewStyle().Foreground(inactiveColor).Render("? for help")

		parts := []string{badge, " ", fileStyled}
		if m.StatusMsg != "" {
			parts = append(parts, "  ", toastStyle.Render("("+m.StatusMsg+")"))
		}
		parts = append(parts, "  ", helpStyled)
		leftText = lipgloss.JoinHorizontal(lipgloss.Center, parts...)
	}

	var rightText string
	if len(m.ActiveJobs) > 0 {
		rightText = fmt.Sprintf("%s Rendering %d frame(s)... ", m.Spinner.View(), len(m.ActiveJobs))
	}

	spacerWidth := m.TermWidth - lipgloss.Width(leftText) - lipgloss.Width(rightText) - 2
	if spacerWidth < 1 {
		spacerWidth = 1
	}
	spacer := strings.Repeat(" ", spacerWidth)

	return lipgloss.JoinHorizontal(lipgloss.Center, leftText, spacer, rightText)
}


