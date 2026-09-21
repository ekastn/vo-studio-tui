// Package main implements Hub launcher navigation and modal dialog handling for vo-studio-tui.
package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// handleHubKey processes keyboard actions in the Project Hub launcher.
func (m *AppModel) handleHubKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	projCount := 0
	if m.Registry != nil {
		projCount = len(m.Registry.Projects)
	}

	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "j", "down":
		if m.HubCursorIdx < projCount-1 {
			m.HubCursorIdx++
		}

	case "k", "up":
		if m.HubCursorIdx > 0 {
			m.HubCursorIdx--
		}

	case "enter":
		if projCount > 0 && m.HubCursorIdx >= 0 && m.HubCursorIdx < projCount {
			entry := m.Registry.Projects[m.HubCursorIdx]
			proj, err := LoadProject(entry.Path)
			if err != nil {
				m.StatusMsg = "Failed to load project: " + err.Error()
				return m, nil
			}
			proj.ApplyGlobalConfig(m.GlobalConfig)
			m.Project = proj
			m.Mode = ModeNormal
			m.ActiveActIdx = 0
			m.CursorIdx = 0
			m.StatusMsg = "Opened " + proj.Name
			m.Registry.Projects[m.HubCursorIdx].LastOpened = time.Now()
			if m.RegistryPath != "" {
				_ = SaveRegistry(m.RegistryPath, m.Registry)
			}
		}

	case "n":
		m.Mode = ModeDialog
		m.DialogAction = "new_project"
		m.DialogPrompt = "New Project Name:"
		m.DialogInput.SetValue("")
		m.DialogInput.Focus()
		return m, nil

	case "a":
		m.Mode = ModeDialog
		m.DialogAction = "register_path"
		m.DialogPrompt = "Register Project Path:"
		m.DialogInput.SetValue("")
		m.DialogInput.Focus()
		return m, nil

	case "d":
		if projCount > 0 && m.HubCursorIdx >= 0 && m.HubCursorIdx < projCount {
			removed := m.Registry.Projects[m.HubCursorIdx].Name
			m.Registry.Unregister(m.Registry.Projects[m.HubCursorIdx].Path)
			if m.HubCursorIdx >= len(m.Registry.Projects) && m.HubCursorIdx > 0 {
				m.HubCursorIdx--
			}
			if m.RegistryPath != "" {
				_ = SaveRegistry(m.RegistryPath, m.Registry)
			}
			m.StatusMsg = "Unregistered " + removed
		}
	}

	return m, nil
}

// handleDialogKey processes input in modal dialogs.
func (m *AppModel) handleDialogKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.DialogInput.Blur()
		if m.Project != nil {
			m.Mode = ModeNormal
		} else {
			m.Mode = ModeHub
		}
		return m, nil

	case "enter":
		val := strings.TrimSpace(m.DialogInput.Value())
		m.DialogInput.Blur()
		if val == "" {
			if m.Project != nil {
				m.Mode = ModeNormal
			} else {
				m.Mode = ModeHub
			}
			return m, nil
		}

		switch m.DialogAction {
		case "new_project":
			parentDir := filepath.Join(DefaultDataDir(), "projects")
			proj, err := CreateNewProject(parentDir, val, 10.0)
			if err != nil {
				m.StatusMsg = "Failed to create project: " + err.Error()
				m.Mode = ModeHub
				return m, nil
			}
			proj.ApplyGlobalConfig(m.GlobalConfig)
			_ = SaveProject(proj)
			if m.Registry == nil {
				m.Registry = &Registry{}
			}
			m.Registry.Register(proj.Name, proj.RootPath)
			if m.RegistryPath != "" {
				_ = SaveRegistry(m.RegistryPath, m.Registry)
			}
			m.Project = proj
			m.Mode = ModeNormal
			m.ActiveActIdx = 0
			m.CursorIdx = 0
			m.StatusMsg = "Created and opened " + proj.Name
			return m, nil

		case "register_path":
			absPath, err := filepath.Abs(val)
			if err != nil {
				m.StatusMsg = "Invalid path: " + err.Error()
				m.Mode = ModeHub
				return m, nil
			}
			proj, err := LoadProject(absPath)
			if err != nil {
				m.StatusMsg = "Failed to load project at path: " + err.Error()
				m.Mode = ModeHub
				return m, nil
			}
			proj.ApplyGlobalConfig(m.GlobalConfig)
			if m.Registry == nil {
				m.Registry = &Registry{}
			}
			m.Registry.Register(proj.Name, proj.RootPath)
			if m.RegistryPath != "" {
				_ = SaveRegistry(m.RegistryPath, m.Registry)
			}
			m.Project = proj
			m.Mode = ModeNormal
			m.ActiveActIdx = 0
			m.CursorIdx = 0
			m.StatusMsg = "Registered and opened " + proj.Name
			return m, nil

		case "add_frame":
			act := m.CurrentAct()
			if act != nil && m.Project != nil {
				f, err := act.AddFrame(m.Project.RootPath, val, m.Project.DefaultSlotDuration)
				if err != nil {
					m.StatusMsg = "Failed to add frame: " + err.Error()
				} else {
					_ = SaveProject(m.Project)
					m.CursorIdx = len(act.Frames) - 1
					m.StatusMsg = fmt.Sprintf("Added frame %02d: %s", f.ID, f.Name)
				}
			}
			m.Mode = ModeNormal
			return m, nil

		case "confirm_delete_frame":
			if strings.ToLower(val) == "yes" || strings.ToLower(val) == "y" {
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
			}
			m.Mode = ModeNormal
			return m, nil

		case "add_act":
			if m.Project != nil {
				act, err := m.Project.AddAct(val)
				if err != nil {
					m.StatusMsg = "Failed to add act: " + err.Error()
				} else {
					m.ActiveActIdx = len(m.Project.Acts) - 1
					m.CursorIdx = 0
					m.StatusMsg = fmt.Sprintf("Added act %d: %s", act.ID, act.Title)
				}
			}
			m.Mode = ModeNormal
			return m, nil

		case "confirm_delete_act":
			if strings.ToLower(val) == "yes" || strings.ToLower(val) == "y" {
				act := m.CurrentAct()
				if act != nil && m.Project != nil {
					_ = m.Project.DeleteAct(act.ID)
					if m.ActiveActIdx >= len(m.Project.Acts) && m.ActiveActIdx > 0 {
						m.ActiveActIdx--
					}
					m.CursorIdx = 0
					m.StatusMsg = fmt.Sprintf("Deleted act %d", act.ID)
				}
			}
			m.Mode = ModeNormal
			return m, nil

		case "edit_speed":
			if m.Project != nil {
				m.Project.Voice.Speed = val
				_ = SaveProject(m.Project)
				m.StatusMsg = fmt.Sprintf("Voice speed set to %s", val)
			}
			m.Mode = ModeNormal
			return m, nil
		}
	}


	var cmd tea.Cmd
	m.DialogInput, cmd = m.DialogInput.Update(msg)
	return m, cmd
}
