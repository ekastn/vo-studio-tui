// Package main implements project registry tracking and project scaffolding for vo-studio-tui.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ProjectEntry represents an individual project recorded in the global registry.
type ProjectEntry struct {
	Name       string    `json:"name"`
	Path       string    `json:"path"`
	LastOpened time.Time `json:"last_opened"`
}

// Registry encapsulates the list of known voiceover projects on the system.
type Registry struct {
	Projects []ProjectEntry `json:"projects"`
}

// DefaultDataDir returns the XDG data directory path for vo-studio.
func DefaultDataDir() string {
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return filepath.Join(xdg, "vo-studio")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ".vo-studio"
	}
	return filepath.Join(home, ".local", "share", "vo-studio")
}

// DefaultConfigDir returns the XDG config directory path for vo-studio.
func DefaultConfigDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "vo-studio")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ".vo-studio"
	}
	return filepath.Join(home, ".config", "vo-studio")
}

// LoadRegistry reads the registry file from disk, creating an empty registry if missing.
func LoadRegistry(path string) (*Registry, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &Registry{Projects: []ProjectEntry{}}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read registry: %w", err)
	}

	var reg Registry
	if err := json.Unmarshal(data, &reg); err != nil {
		return nil, fmt.Errorf("failed to parse registry: %w", err)
	}
	return &reg, nil
}

// SaveRegistry writes the registry to disk, creating parent directories if needed.
func SaveRegistry(path string, reg *Registry) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create registry directory: %w", err)
	}

	data, err := json.MarshalIndent(reg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode registry: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

// Register adds or updates a project entry in the registry.
func (r *Registry) Register(name, projectPath string) {
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		absPath = projectPath
	}

	for i, p := range r.Projects {
		if p.Path == absPath {
			r.Projects[i].Name = name
			r.Projects[i].LastOpened = time.Now()
			return
		}
	}

	r.Projects = append(r.Projects, ProjectEntry{
		Name:       name,
		Path:       absPath,
		LastOpened: time.Now(),
	})
}

// Unregister removes a project by path from the registry.
func (r *Registry) Unregister(projectPath string) {
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		absPath = projectPath
	}

	var updated []ProjectEntry
	for _, p := range r.Projects {
		if p.Path != absPath {
			updated = append(updated, p)
		}
	}
	r.Projects = updated
}

// CreateNewProject scaffolds a new voiceover project on disk.
func CreateNewProject(parentDir, projectName string, defaultSlot float64) (*Project, error) {
	slug := strings.ToLower(strings.ReplaceAll(projectName, " ", "-"))
	projectDir := filepath.Join(parentDir, slug)

	if err := os.MkdirAll(filepath.Join(projectDir, "scripts", "act_01"), 0755); err != nil {
		return nil, fmt.Errorf("failed to create script dir: %w", err)
	}
	if err := os.MkdirAll(filepath.Join(projectDir, "audio", "act_01", "raw"), 0755); err != nil {
		return nil, fmt.Errorf("failed to create audio raw dir: %w", err)
	}
	if err := os.MkdirAll(filepath.Join(projectDir, "audio", "act_01", "padded"), 0755); err != nil {
		return nil, fmt.Errorf("failed to create audio padded dir: %w", err)
	}

	// Create initial frame script
	initialScript := filepath.Join(projectDir, "scripts", "act_01", "frame_01.txt")
	if err := os.WriteFile(initialScript, []byte("Enter narration text here.\n"), 0644); err != nil {
		return nil, fmt.Errorf("failed to write initial script: %w", err)
	}

	if defaultSlot <= 0 {
		defaultSlot = 10.0
	}

	proj := &Project{
		RootPath:            projectDir,
		Name:                projectName,
		Version:             1,
		DefaultSlotDuration: defaultSlot,
		Voice: VoiceConfig{
			ProfileID: "default",
			Language:  "English",
			Instruct:  "",
			Speed:     "1.0",
		},
		ExternalAudioEditor: "audacity",
		Acts: []*Act{
			{
				ID:    1,
				Title: "Act 1",
				Frames: []*Frame{
					{
						ID:           1,
						Name:         "Frame 1",
						ScriptFile:   "scripts/act_01/frame_01.txt",
						SlotDuration: defaultSlot,
					},
				},
			},
		},
	}

	if err := SaveProject(proj); err != nil {
		return nil, fmt.Errorf("failed to save new project manifest: %w", err)
	}

	return LoadProject(projectDir)
}
