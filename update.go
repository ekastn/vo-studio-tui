// Package main implements the update event loop for vo-studio-tui.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// EditorFinishedMsg is dispatched when the external editor process terminates.
type EditorFinishedMsg struct {
	FrameID int
	Err     error
}

// Init initializes the Bubble Tea application.
func (m *AppModel) Init() tea.Cmd {
	return m.Spinner.Tick
}

// Update processes incoming Bubble Tea messages and updates model state.
func (m *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.TermWidth = msg.Width
		m.TermHeight = msg.Height
		return m, nil

	case EditorFinishedMsg:
		f := m.CurrentFrame()
		if f != nil && f.ID == msg.FrameID {
			if data, err := os.ReadFile(f.ScriptPath); err == nil {
				f.Text = strings.TrimSpace(string(data))
				m.StatusMsg = fmt.Sprintf("Frame %02d script updated", f.ID)
			}
		}
		return m, nil

	case RenderCompleteMsg:
		delete(m.ActiveJobs, msg.FrameID)
		if m.Project != nil {
			for _, act := range m.Project.Acts {
				for _, f := range act.Frames {
					if f.ID == msg.FrameID {
						if msg.Err != nil {
							f.Status = StatusError
							f.ErrorMsg = msg.Err.Error()
							m.StatusMsg = fmt.Sprintf("Frame %02d failed: %v", msg.FrameID, msg.Err)
						} else {
							if f.SlotDuration > 0 {
								f.Status = StatusPadded
							} else {
								f.Status = StatusRendered
							}
							f.RawDuration = msg.Duration
							f.Waveform = nil
							f.ErrorMsg = ""
							m.StatusMsg = fmt.Sprintf("Frame %02d rendered (%.2fs)", msg.FrameID, msg.Duration)
						}
						break
					}
				}
			}
		}
		return m, nil

	case PlaybackTickMsg:
		if m.IsPlaying {
			m.PlayTick++
			return m, playbackTickCmd()
		}
		return m, nil

	case PlaybackFinishedMsg:
		m.IsPlaying = false
		m.PlayingFile = ""
		m.PlayingFrameID = 0
		m.PlayTick = 0
		if msg.Err != nil {
			m.StatusMsg = fmt.Sprintf("Playback stopped: %v", msg.Err)
		} else {
			m.StatusMsg = "Playback finished"
		}
		return m, nil

	case AudacityLaunchedMsg:
		if msg.Err != nil {
			m.StatusMsg = fmt.Sprintf("Audio editor error: %v", msg.Err)
		} else {
			m.StatusMsg = "Audio editor opened successfully"
		}
		return m, nil

	case AssembleCompleteMsg:
		if msg.Err != nil {
			m.StatusMsg = fmt.Sprintf("Assemble error: %v", msg.Err)
		} else {
			m.StatusMsg = fmt.Sprintf("Master ready: %s (%.2fs)", filepath.Base(msg.Path), msg.Duration)
		}
		return m, nil

	case tea.KeyMsg:



		switch m.Mode {
		case ModeHub:
			return m.handleHubKey(msg)
		case ModeDialog:
			return m.handleDialogKey(msg)
		case ModeCommand:
			return m.handleCommandKey(msg)
		case ModeSearch:
			return m.handleSearchKey(msg)
		case ModeNormal:
			return m.handleNormalKey(msg)
		}
	}




	return m, nil
}

// handleNormalKey processes standard Vim-style navigation keys in normal mode.
func (m *AppModel) handleNormalKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.ShowHelp {
		switch msg.String() {
		case "esc", "q", "?":
			m.ShowHelp = false
			return m, nil
		}
	}

	act := m.CurrentAct()
	frameCount := 0
	if act != nil {
		frameCount = len(act.Frames)
		if len(m.FilteredIdxs) > 0 {
			frameCount = len(m.FilteredIdxs)
		}
	}

	switch msg.String() {
	case "?":
		m.ShowHelp = !m.ShowHelp
		return m, nil

	case "q", "ctrl+c":
		return m, tea.Quit

	case "j", "down":
		if m.CursorIdx < frameCount-1 {
			m.CursorIdx++
		}

	case "k", "up":
		if m.CursorIdx > 0 {
			m.CursorIdx--
		}

	case "g":
		m.CursorIdx = 0

	case "G":
		if frameCount > 0 {
			m.CursorIdx = frameCount - 1
		}

	case "l", "right", "tab":
		if m.Project != nil && len(m.Project.Acts) > 0 {
			m.ActiveActIdx = (m.ActiveActIdx + 1) % len(m.Project.Acts)
			m.CursorIdx = 0
			m.FilteredIdxs = nil
		}

	case "h", "left", "shift+tab":
		if m.Project != nil && len(m.Project.Acts) > 0 {
			m.ActiveActIdx = (m.ActiveActIdx - 1 + len(m.Project.Acts)) % len(m.Project.Acts)
			m.CursorIdx = 0
			m.FilteredIdxs = nil
		}

	case "ctrl+p":
		if m.Registry != nil {
			m.Mode = ModeHub
			m.StatusMsg = ""
			return m, nil
		}

	case "a":
		if m.CurrentAct() != nil {
			m.Mode = ModeDialog
			m.DialogAction = "add_frame"
			m.DialogPrompt = "New Frame Title:"
			m.DialogInput.SetValue("")
			m.DialogInput.Focus()
			return m, nil
		}

	case "d":
		f := m.CurrentFrame()
		if f != nil {
			m.Mode = ModeDialog
			m.DialogAction = "confirm_delete_frame"
			m.DialogPrompt = fmt.Sprintf("Delete Frame %02d? (type 'yes' to confirm):", f.ID)
			m.DialogInput.SetValue("")
			m.DialogInput.Focus()
			return m, nil
		}

	case "A":
		m.Mode = ModeDialog
		m.DialogAction = "add_act"
		m.DialogPrompt = "New Act Title:"
		m.DialogInput.SetValue("")
		m.DialogInput.Focus()
		return m, nil

	case "D":
		act := m.CurrentAct()
		if act != nil {
			m.Mode = ModeDialog
			m.DialogAction = "confirm_delete_act"
			m.DialogPrompt = fmt.Sprintf("Delete Act %d (%s)? (type 'yes' to confirm):", act.ID, act.Title)
			m.DialogInput.SetValue("")
			m.DialogInput.Focus()
			return m, nil
		}

	case "J":
		act := m.CurrentAct()
		if act != nil && m.CursorIdx < len(act.Frames)-1 {
			_ = act.MoveFrame(m.CursorIdx, 1)
			m.CursorIdx++
			_ = SaveProject(m.Project)
		}

	case "K":
		act := m.CurrentAct()
		if act != nil && m.CursorIdx > 0 {
			_ = act.MoveFrame(m.CursorIdx, -1)
			m.CursorIdx--
			_ = SaveProject(m.Project)
		}

	case "e":
		f := m.CurrentFrame()
		if f != nil {
			editor := os.Getenv("EDITOR")
			if editor == "" {
				editor = "nvim"
			}
			cmd := exec.Command(editor, f.ScriptPath)
			return m, tea.ExecProcess(cmd, func(err error) tea.Msg {
				return EditorFinishedMsg{FrameID: f.ID, Err: err}
			})
		}

	case "r":
		f := m.CurrentFrame()
		if f != nil && m.Project != nil {
			m.ActiveJobs[f.ID] = true
			m.StatusMsg = fmt.Sprintf("Rendering Frame %02d...", f.ID)
			return m, tea.Batch(m.Spinner.Tick, renderFrameCmd(f, m.Project.Voice))
		}

	case "R":
		act := m.CurrentAct()
		if act != nil && m.Project != nil {
			var batch []tea.Cmd
			batch = append(batch, m.Spinner.Tick)
			for _, f := range act.Frames {
				if f.Status != StatusPadded && f.Status != StatusRendered {
					m.ActiveJobs[f.ID] = true
					batch = append(batch, renderFrameCmd(f, m.Project.Voice))
				}
			}
			m.StatusMsg = fmt.Sprintf("Batch rendering Act %d...", act.ID)
			return m, tea.Batch(batch...)
		}

	case "s":
		if m.IsPlaying {
			m.StatusMsg = "Stopping playback..."
			return m, stopAudioCmd()
		}

	case "p", " ":
		if m.IsPlaying {
			m.StatusMsg = "Stopping playback..."
			return m, stopAudioCmd()
		}
		f := m.CurrentFrame()
		if f != nil {
			target := f.RawPath
			dur := f.RawDuration
			if _, err := os.Stat(target); err != nil {
				target = f.PaddedPath
				dur = f.SlotDuration
			}
			m.IsPlaying = true
			m.PlayingFrameID = f.ID
			m.PlayTick = 0
			m.PlayStartTime = time.Now()
			m.PlayDuration = dur
			m.PlayingFile = filepath.Base(target)
			m.StatusMsg = fmt.Sprintf("Playing %s...", m.PlayingFile)
			return m, tea.Batch(playAudioCmd(target), playbackTickCmd())
		}

	case "P":
		if m.IsPlaying {
			m.StatusMsg = "Stopping playback..."
			return m, stopAudioCmd()
		}
		f := m.CurrentFrame()
		if f != nil {
			target := f.PaddedPath
			dur := f.SlotDuration
			if _, err := os.Stat(target); err != nil {
				target = f.RawPath
				dur = f.RawDuration
			}
			m.IsPlaying = true
			m.PlayingFrameID = f.ID
			m.PlayTick = 0
			m.PlayStartTime = time.Now()
			m.PlayDuration = dur
			m.PlayingFile = filepath.Base(target)
			m.StatusMsg = fmt.Sprintf("Playing %s...", m.PlayingFile)
			return m, tea.Batch(playAudioCmd(target), playbackTickCmd())
		}

	case "o":
		f := m.CurrentFrame()
		if f != nil && m.Project != nil {
			editor := m.Project.ExternalAudioEditor
			m.StatusMsg = fmt.Sprintf("Opening frame %02d in %s...", f.ID, editor)
			return m, openAudioEditorCmd(editor, []string{f.RawPath, f.PaddedPath})
		}

	case "m":
		act := m.CurrentAct()
		if act != nil && m.Project != nil {
			editor := m.Project.ExternalAudioEditor
			m.StatusMsg = fmt.Sprintf("Opening Act %d master in %s...", act.ID, editor)
			return m, openAudioEditorCmd(editor, []string{act.MasterPath})
		}

	case "M":
		if m.Project != nil {
			editor := m.Project.ExternalAudioEditor
			fullMaster := filepath.Join(m.Project.RootPath, "audio", "full_production_master.wav")
			m.StatusMsg = fmt.Sprintf("Opening full production master in %s...", editor)
			return m, openAudioEditorCmd(editor, []string{fullMaster})
		}

	case "/":
		m.Mode = ModeSearch
		m.SearchInput.SetValue("")
		m.SearchInput.Focus()
		return m, nil

	case "c":
		if m.Project != nil {
			m.Mode = ModeDialog
			m.DialogAction = "edit_speed"
			m.DialogPrompt = fmt.Sprintf("Voice Speed (current: %s):", m.Project.Voice.Speed)
			m.DialogInput.SetValue(m.Project.Voice.Speed)
			m.DialogInput.Focus()
			return m, nil
		}

	case ":":
		m.Mode = ModeCommand
		m.CmdInput.SetValue(":")
		m.CmdInput.Focus()
		return m, nil

	case "1", "2", "3", "4", "5", "6", "7", "8", "9":
		idx, _ := strconv.Atoi(msg.String())
		if m.Project != nil && idx >= 1 && idx <= len(m.Project.Acts) {
			m.ActiveActIdx = idx - 1
			m.CursorIdx = 0
			m.FilteredIdxs = nil
		}
	}

	return m, nil
}



