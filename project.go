// Package main implements project data modeling, loading, and persistence for vo-studio-tui.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FrameStatus represents the current audio generation state of a voiceover frame.
type FrameStatus int

const (
	StatusMissing FrameStatus = iota
	StatusRendering
	StatusRendered
	StatusPadded
	StatusError
)

// String returns a human-readable representation of FrameStatus.
func (s FrameStatus) String() string {
	switch s {
	case StatusMissing:
		return "Missing"
	case StatusRendering:
		return "Rendering"
	case StatusRendered:
		return "Rendered"
	case StatusPadded:
		return "Padded"
	case StatusError:
		return "Error"
	default:
		return "Unknown"
	}
}

// VoiceConfig defines default speech synthesis parameters for a project.
type VoiceConfig struct {
	APIURL    string `json:"api_url,omitempty"`
	APIToken  string `json:"token,omitempty"`
	ProfileID string `json:"profile_id,omitempty"`
	Language  string `json:"language,omitempty"`
	Instruct  string `json:"instruct,omitempty"`
	Speed     string `json:"speed,omitempty"`
}

// VoiceOverride defines optional frame-level voice synthesis overrides.
type VoiceOverride struct {
	ProfileID string `json:"profile_id,omitempty"`
	Instruct  string `json:"instruct,omitempty"`
	Speed     string `json:"speed,omitempty"`
}

// Frame represents an individual timeline voiceover segment.
type Frame struct {
	ID            int           `json:"id"`
	Name          string        `json:"name"`
	ScriptFile    string        `json:"script_file"`
	SlotDuration  float64       `json:"slot_duration"`
	IsSilent      bool          `json:"is_silent"`
	VoiceOverride VoiceOverride `json:"voice_override,omitempty"`

	// In-memory runtime state
	ActID       int         `json:"-"`
	Text        string      `json:"-"`
	ScriptPath  string      `json:"-"`
	RawPath     string      `json:"-"`
	PaddedPath  string      `json:"-"`
	RawDuration float64     `json:"-"`
	Waveform    []float64   `json:"-"`
	Status      FrameStatus `json:"-"`
	ErrorMsg    string      `json:"-"`
}

// Act represents a top-level narrative section containing sequential frames.
type Act struct {
	ID         int      `json:"id"`
	Title      string   `json:"title"`
	Frames     []*Frame `json:"frames"`
	MasterPath string   `json:"-"`
}

// Project represents a complete voiceover production.
type Project struct {
	Name                string      `json:"name"`
	Version             int         `json:"version"`
	DefaultSlotDuration float64     `json:"default_slot_duration"`
	Voice               VoiceConfig `json:"voice"`
	ExternalAudioEditor string      `json:"external_audio_editor,omitempty"`
	Acts                []*Act      `json:"acts"`

	RootPath string `json:"-"`
}

// LoadProject reads vo-project.json from projectDir and initializes runtime paths and frame state.
func LoadProject(projectDir string) (*Project, error) {
	manifestPath := filepath.Join(projectDir, "vo-project.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read project manifest: %w", err)
	}

	var p Project
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("failed to parse project manifest: %w", err)
	}
	p.RootPath = projectDir

	if p.ExternalAudioEditor == "" {
		p.ExternalAudioEditor = "audacity"
	}

	for _, act := range p.Acts {
		actDirName := fmt.Sprintf("act_%02d", act.ID)
		actAudioDir := filepath.Join(projectDir, "audio", actDirName)
		act.MasterPath = filepath.Join(actAudioDir, fmt.Sprintf("act_%02d_master.wav", act.ID))

		for _, frame := range act.Frames {
			frame.ActID = act.ID

			if frame.ScriptFile == "" {
				frame.ScriptFile = filepath.Join("scripts", actDirName, fmt.Sprintf("frame_%02d.txt", frame.ID))
			}
			frame.ScriptPath = filepath.Join(projectDir, frame.ScriptFile)

			frame.RawPath = filepath.Join(actAudioDir, "raw", fmt.Sprintf("frame_%02d.wav", frame.ID))
			frame.PaddedPath = filepath.Join(actAudioDir, "padded", fmt.Sprintf("frame_%02d_padded.wav", frame.ID))

			// Load script text from disk if exists
			if scriptBytes, err := os.ReadFile(frame.ScriptPath); err == nil {
				frame.Text = strings.TrimSpace(string(scriptBytes))
			}

			// Check audio status
			frame.Status = StatusMissing
			if frame.IsSilent {
				if _, err := os.Stat(frame.PaddedPath); err == nil {
					frame.Status = StatusPadded
					frame.RawDuration = frame.SlotDuration
				}
			} else {
				if fi, err := os.Stat(frame.RawPath); err == nil && fi.Size() > 1000 {
					frame.Status = StatusRendered
				}
				if fi, err := os.Stat(frame.PaddedPath); err == nil && fi.Size() > 1000 {
					frame.Status = StatusPadded
				}
			}
		}
	}

	return &p, nil
}

// SaveProject serializes the project manifest to vo-project.json in p.RootPath.
func SaveProject(p *Project) error {
	if p.RootPath == "" {
		return fmt.Errorf("project RootPath is empty")
	}

	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode project manifest: %w", err)
	}

	manifestPath := filepath.Join(p.RootPath, "vo-project.json")
	return os.WriteFile(manifestPath, data, 0644)
}

// AddAct creates and appends a new Act to the project.
func (p *Project) AddAct(title string) (*Act, error) {
	maxID := 0
	for _, a := range p.Acts {
		if a.ID > maxID {
			maxID = a.ID
		}
	}
	newID := maxID + 1

	actDirName := fmt.Sprintf("act_%02d", newID)
	if err := os.MkdirAll(filepath.Join(p.RootPath, "scripts", actDirName), 0755); err != nil {
		return nil, fmt.Errorf("failed to create script directory for act: %w", err)
	}
	if err := os.MkdirAll(filepath.Join(p.RootPath, "audio", actDirName, "raw"), 0755); err != nil {
		return nil, fmt.Errorf("failed to create audio raw directory for act: %w", err)
	}
	if err := os.MkdirAll(filepath.Join(p.RootPath, "audio", actDirName, "padded"), 0755); err != nil {
		return nil, fmt.Errorf("failed to create audio padded directory for act: %w", err)
	}

	act := &Act{
		ID:         newID,
		Title:      title,
		Frames:     []*Frame{},
		MasterPath: filepath.Join(p.RootPath, "audio", actDirName, fmt.Sprintf("act_%02d_master.wav", newID)),
	}
	p.Acts = append(p.Acts, act)
	return act, SaveProject(p)
}

// RenameAct updates the title of an act and saves the project manifest.
func (p *Project) RenameAct(actID int, newTitle string) error {
	for _, a := range p.Acts {
		if a.ID == actID {
			a.Title = newTitle
			return SaveProject(p)
		}
	}
	return fmt.Errorf("act %d not found", actID)
}

// DeleteAct removes an act from the project and updates the manifest.
func (p *Project) DeleteAct(actID int) error {
	var updated []*Act
	found := false
	for _, a := range p.Acts {
		if a.ID == actID {
			found = true
			continue
		}
		updated = append(updated, a)
	}
	if !found {
		return fmt.Errorf("act %d not found", actID)
	}
	p.Acts = updated
	return SaveProject(p)
}

// AddFrame creates a new frame in the act, creates its script file on disk, and returns the frame.
func (act *Act) AddFrame(projectRoot string, name string, slotDur float64) (*Frame, error) {
	maxID := 0
	for _, f := range act.Frames {
		if f.ID > maxID {
			maxID = f.ID
		}
	}
	newID := maxID + 1

	actDirName := fmt.Sprintf("act_%02d", act.ID)
	scriptRel := filepath.Join("scripts", actDirName, fmt.Sprintf("frame_%02d.txt", newID))
	scriptAbs := filepath.Join(projectRoot, scriptRel)

	if err := os.MkdirAll(filepath.Dir(scriptAbs), 0755); err != nil {
		return nil, fmt.Errorf("failed to create script directory: %w", err)
	}
	if _, err := os.Stat(scriptAbs); os.IsNotExist(err) {
		if err := os.WriteFile(scriptAbs, []byte(""), 0644); err != nil {
			return nil, fmt.Errorf("failed to create script file: %w", err)
		}
	}

	rawAudio := filepath.Join(projectRoot, "audio", actDirName, "raw", fmt.Sprintf("frame_%02d.wav", newID))
	paddedAudio := filepath.Join(projectRoot, "audio", actDirName, "padded", fmt.Sprintf("frame_%02d_padded.wav", newID))

	f := &Frame{
		ID:           newID,
		ActID:        act.ID,
		Name:         name,
		ScriptFile:   scriptRel,
		ScriptPath:   scriptAbs,
		RawPath:      rawAudio,
		PaddedPath:   paddedAudio,
		SlotDuration: slotDur,
		Status:       StatusMissing,
	}

	act.Frames = append(act.Frames, f)
	return f, nil
}

// DeleteFrame removes a frame by ID from the act.
func (act *Act) DeleteFrame(frameID int) error {
	var updated []*Frame
	found := false
	for _, f := range act.Frames {
		if f.ID == frameID {
			found = true
			continue
		}
		updated = append(updated, f)
	}
	if !found {
		return fmt.Errorf("frame %d not found in act %d", frameID, act.ID)
	}
	act.Frames = updated
	return nil
}

// MoveFrame swaps the frame at frameIdx with frameIdx+delta.
func (act *Act) MoveFrame(frameIdx int, delta int) error {
	targetIdx := frameIdx + delta
	if frameIdx < 0 || frameIdx >= len(act.Frames) || targetIdx < 0 || targetIdx >= len(act.Frames) {
		return fmt.Errorf("move target out of bounds")
	}
	act.Frames[frameIdx], act.Frames[targetIdx] = act.Frames[targetIdx], act.Frames[frameIdx]
	return nil
}

// SaveFrameScript persists new narration text to the frame's script file and updates in-memory text.
func SaveFrameScript(f *Frame, newText string) error {
	dir := filepath.Dir(f.ScriptPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create script dir: %w", err)
	}
	f.Text = strings.TrimSpace(newText)
	return os.WriteFile(f.ScriptPath, []byte(f.Text+"\n"), 0644)
}

// ReloadScripts re-reads all script files for all acts and frames from disk.
func (p *Project) ReloadScripts() error {
	for _, act := range p.Acts {
		for _, f := range act.Frames {
			if f.ScriptPath != "" {
				if data, err := os.ReadFile(f.ScriptPath); err == nil {
					f.Text = strings.TrimSpace(string(data))
				}
			}
		}
	}
	return nil
}

// computeScriptStats calculates word count and character count for narration text.
func computeScriptStats(text string) (words, chars int) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return 0, 0
	}
	words = len(strings.Fields(trimmed))
	chars = len(trimmed)
	return words, chars
}



