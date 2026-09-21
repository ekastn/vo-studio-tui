// Package main implements command mode prompt parsing and execution for vo-studio-tui.
package main

import (
	"fmt"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// handleCommandKey processes input when in command mode (:).
func (m *AppModel) handleCommandKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.Mode = ModeNormal
		m.CmdInput.Blur()
		return m, nil

	case "enter":
		cmdText := strings.TrimSpace(m.CmdInput.Value())
		m.Mode = ModeNormal
		m.CmdInput.Blur()
		return m.executeCommand(cmdText)
	}

	var cmd tea.Cmd
	m.CmdInput, cmd = m.CmdInput.Update(msg)
	return m, cmd
}

// executeCommand runs commands entered at the : prompt.
func (m *AppModel) executeCommand(cmdStr string) (tea.Model, tea.Cmd) {
	cmdStr = strings.TrimPrefix(cmdStr, ":")
	parts := strings.Fields(cmdStr)
	if len(parts) == 0 {
		return m, nil
	}

	switch parts[0] {
	case "q", "quit":
		return m, tea.Quit

	case "h", "help":
		m.ShowHelp = !m.ShowHelp
		return m, nil

	case "add-frame":
		name := "New Frame"
		if len(parts) > 1 {
			name = strings.Join(parts[1:], " ")
		}
		act := m.CurrentAct()
		if act != nil && m.Project != nil {
			f, err := act.AddFrame(m.Project.RootPath, name, m.Project.DefaultSlotDuration)
			if err != nil {
				m.StatusMsg = "Failed to add frame: " + err.Error()
			} else {
				_ = SaveProject(m.Project)
				m.CursorIdx = len(act.Frames) - 1
				m.StatusMsg = fmt.Sprintf("Added frame %02d: %s", f.ID, f.Name)
			}
		}

	case "del-frame":
		act := m.CurrentAct()
		f := m.CurrentFrame()
		if act != nil && f != nil && m.Project != nil {
			_ = act.DeleteFrame(f.ID)
			_ = SaveProject(m.Project)
			if m.CursorIdx >= len(act.Frames) && m.CursorIdx > 0 {
				m.CursorIdx--
			}
			m.StatusMsg = fmt.Sprintf("Deleted frame %02d", f.ID)
		}

	case "add-act":
		title := "New Act"
		if len(parts) > 1 {
			title = strings.Join(parts[1:], " ")
		}
		if m.Project != nil {
			act, err := m.Project.AddAct(title)
			if err != nil {
				m.StatusMsg = "Failed to add act: " + err.Error()
			} else {
				m.ActiveActIdx = len(m.Project.Acts) - 1
				m.CursorIdx = 0
				m.StatusMsg = fmt.Sprintf("Added act %d: %s", act.ID, act.Title)
			}
		}

	case "rename-act":
		if len(parts) > 1 && m.Project != nil {
			title := strings.Join(parts[1:], " ")
			act := m.CurrentAct()
			if act != nil {
				_ = m.Project.RenameAct(act.ID, title)
				m.StatusMsg = "Renamed act to: " + title
			}
		}

	case "del-act":
		act := m.CurrentAct()
		if act != nil && m.Project != nil {
			_ = m.Project.DeleteAct(act.ID)
			if m.ActiveActIdx >= len(m.Project.Acts) && m.ActiveActIdx > 0 {
				m.ActiveActIdx--
			}
			m.CursorIdx = 0
			m.StatusMsg = fmt.Sprintf("Deleted act %d", act.ID)
		}

	case "w", "reload":
		if m.Project != nil {
			_ = m.Project.ReloadScripts()
			m.StatusMsg = "Reloaded frame scripts from disk"
		}

	case "assemble":
		act := m.CurrentAct()
		if act != nil {
			m.StatusMsg = fmt.Sprintf("Assembling Act %d master...", act.ID)
			return m, assembleActCmd(act)
		}

	case "assemble-all":
		if m.Project != nil {
			out := filepath.Join(m.Project.RootPath, "audio", "full_production_master.wav")
			m.StatusMsg = "Assembling full production master..."
			return m, assembleGlobalMasterCmd(m.Project.Acts, out)
		}

	case "s", "stop":
		if m.IsPlaying {
			m.StatusMsg = "Stopping playback..."
			return m, stopAudioCmd()
		}

	case "act":
		act := m.CurrentAct()
		if act != nil && m.Project != nil {
			editor := m.Project.ExternalAudioEditor
			m.StatusMsg = fmt.Sprintf("Opening Act %d Master in %s...", act.ID, editor)
			return m, openAudioEditorCmd(editor, []string{act.MasterPath})
		}

	case "full":
		if m.Project != nil {
			fullMaster := filepath.Join(m.Project.RootPath, "audio", "full_production_master.wav")
			editor := m.Project.ExternalAudioEditor
			m.StatusMsg = fmt.Sprintf("Opening Full Master in %s...", editor)
			return m, openAudioEditorCmd(editor, []string{fullMaster})
		}

	case "editor", "audacity", "a":
		f := m.CurrentFrame()
		if f != nil && m.Project != nil {
			editor := m.Project.ExternalAudioEditor
			m.StatusMsg = fmt.Sprintf("Opening frame %02d in %s...", f.ID, editor)
			return m, openAudioEditorCmd(editor, []string{f.RawPath, f.PaddedPath})
		}

	case "dir", "open":
		if m.Project != nil && m.Project.RootPath != "" {
			m.StatusMsg = fmt.Sprintf("Opening project directory: %s...", filepath.Base(m.Project.RootPath))
			return m, openDirectoryCmd(m.Project.RootPath)
		}

	case "speed":
		if len(parts) > 1 && m.Project != nil {
			m.Project.Voice.Speed = parts[1]
			_ = SaveProject(m.Project)
			m.StatusMsg = fmt.Sprintf("Voice speed set to %s", parts[1])
		} else if m.Project != nil {
			m.StatusMsg = fmt.Sprintf("Current voice speed: %s", m.Project.Voice.Speed)
		}

	case "profile":
		if len(parts) > 1 && m.Project != nil {
			m.Project.Voice.ProfileID = parts[1]
			_ = SaveProject(m.Project)
			m.StatusMsg = fmt.Sprintf("Speaker profile set to %s", parts[1])
		} else if m.Project != nil {
			m.StatusMsg = fmt.Sprintf("Current speaker profile: %s", m.Project.Voice.ProfileID)
		}

	case "lang", "language":
		if len(parts) > 1 && m.Project != nil {
			m.Project.Voice.Language = strings.Join(parts[1:], " ")
			_ = SaveProject(m.Project)
			m.StatusMsg = fmt.Sprintf("Voice language set to %s", m.Project.Voice.Language)
		} else if m.Project != nil {
			m.StatusMsg = fmt.Sprintf("Current voice language: %s", m.Project.Voice.Language)
		}

	case "instruct":
		if len(parts) > 1 && m.Project != nil {
			m.Project.Voice.Instruct = strings.Join(parts[1:], " ")
			_ = SaveProject(m.Project)
			m.StatusMsg = fmt.Sprintf("Voice style set to '%s'", m.Project.Voice.Instruct)
		} else if m.Project != nil {
			m.StatusMsg = fmt.Sprintf("Current voice style: '%s'", m.Project.Voice.Instruct)
		}

	default:
		m.StatusMsg = fmt.Sprintf("Unknown command: :%s", cmdStr)
	}

	return m, nil
}
