// Package main implements the state model for the vo-studio-tui application.
package main

import (
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
)

// AppMode represents the active modal interaction mode of the TUI.
type AppMode int

const (
	ModeNormal AppMode = iota
	ModeCommand
	ModeSearch
	ModeDialog
	ModeHub
)

// AppModel holds the central application state for Bubble Tea.
type AppModel struct {
	Project      *Project
	Mode         AppMode
	ActiveActIdx int
	CursorIdx    int
	FilteredIdxs []int

	// Hub & Registry
	Registry     *Registry
	RegistryPath string
	HubCursorIdx int
	GlobalConfig *GlobalConfig


	// Modal Dialogs
	DialogPrompt string
	DialogInput  textinput.Model
	DialogAction string

	// Sub-components
	Spinner     spinner.Model
	CmdInput    textinput.Model
	SearchInput textinput.Model
	DetailView  viewport.Model

	// Terminal dimensions
	TermWidth  int
	TermHeight int

	// Status and audio playback tracking
	StatusMsg      string
	IsPlaying      bool
	PlayingFile    string
	PlayingFrameID int
	PlayStartTime  time.Time
	PlayDuration   float64
	PlayTick       int
	ShowHelp       bool
	ActiveJobs     map[int]bool
	SearchQuery    string
}

// NewAppModel initializes and returns an AppModel with default sub-components.
func NewAppModel(p *Project) *AppModel {
	m := initBaseModel()
	m.Project = p
	m.Mode = ModeNormal
	return m
}

// NewHubModel initializes an AppModel in ModeHub pointing to a global registry.
func NewHubModel(reg *Registry, regPath string) *AppModel {
	m := initBaseModel()
	m.Registry = reg
	m.RegistryPath = regPath
	m.Mode = ModeHub
	return m
}

func initBaseModel() *AppModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#fabd2f"))

	cmdInput := textinput.New()
	cmdInput.Prompt = ":"
	cmdInput.Placeholder = "command (w, q, assemble, assemble-all, speed <val>)"

	searchInput := textinput.New()
	searchInput.Prompt = "/"
	searchInput.Placeholder = "search frames by text or ID..."

	dialogInput := textinput.New()

	vp := viewport.New(60, 20)

	return &AppModel{
		Spinner:      s,
		CmdInput:     cmdInput,
		SearchInput:  searchInput,
		DialogInput:  dialogInput,
		DetailView:   vp,
		ActiveJobs:   make(map[int]bool),
		StatusMsg:    "",
	}
}


// CurrentAct returns a pointer to the currently selected Act.
func (m *AppModel) CurrentAct() *Act {
	if m.Project == nil || m.ActiveActIdx < 0 || m.ActiveActIdx >= len(m.Project.Acts) {
		return nil
	}
	return m.Project.Acts[m.ActiveActIdx]
}

// CurrentFrame returns a pointer to the currently highlighted Frame.
func (m *AppModel) CurrentFrame() *Frame {
	act := m.CurrentAct()
	if act == nil || len(act.Frames) == 0 {
		return nil
	}
	if len(m.FilteredIdxs) > 0 {
		if m.CursorIdx < 0 || m.CursorIdx >= len(m.FilteredIdxs) {
			return nil
		}
		targetID := m.FilteredIdxs[m.CursorIdx]
		for _, f := range act.Frames {
			if f.ID == targetID {
				return f
			}
		}
	}
	if m.CursorIdx < 0 || m.CursorIdx >= len(act.Frames) {
		return nil
	}
	return act.Frames[m.CursorIdx]
}
