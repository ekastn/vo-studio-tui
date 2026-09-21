// Package main implements audio playback, ffmpeg slot padding, silence generation, and master concatenation.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// PlaybackTickMsg is dispatched periodically during audio playback to sweep the playhead.
type PlaybackTickMsg time.Time

// playbackTickCmd fires an animation tick every 80 milliseconds.
func playbackTickCmd() tea.Cmd {
	return tea.Tick(80*time.Millisecond, func(t time.Time) tea.Msg {
		return PlaybackTickMsg(t)
	})
}

// PlaybackFinishedMsg is dispatched when mpv playback completes.
type PlaybackFinishedMsg struct {
	Path string
	Err  error
}

// AudacityLaunchedMsg is dispatched when the external audio editor completes or fails to launch.
type AudacityLaunchedMsg struct {
	Paths []string
	Err   error
}

// AssembleCompleteMsg is dispatched when an Act or Global master is assembled.
type AssembleCompleteMsg struct {
	ActID    int
	Path     string
	Duration float64
	Err      error
}

// playAudioCmd launches mpv in the background without blocking the TUI.
func playAudioCmd(path string) tea.Cmd {
	return func() tea.Msg {
		if _, err := os.Stat(path); err != nil {
			return PlaybackFinishedMsg{Path: path, Err: err}
		}
		cmd := exec.Command("mpv", "--no-video", "--really-quiet", path)
		err := cmd.Run()
		return PlaybackFinishedMsg{Path: path, Err: err}
	}
}

// stopAudioCmd terminates active mpv audio playback.
func stopAudioCmd() tea.Cmd {
	return func() tea.Msg {
		_ = exec.Command("pkill", "-TERM", "mpv").Run()
		return nil
	}
}

// openAudioEditorCmd launches the configured external audio editor detached from the TUI.
func openAudioEditorCmd(editorCmd string, paths []string) tea.Cmd {
	return func() tea.Msg {
		if editorCmd == "" {
			editorCmd = "audacity"
		}
		var valid []string
		for _, p := range paths {
			if _, err := os.Stat(p); err == nil {
				valid = append(valid, p)
			}
		}
		if len(valid) == 0 {
			return AudacityLaunchedMsg{Paths: paths, Err: fmt.Errorf("no valid audio files found to open")}
		}

		cmd := exec.Command(editorCmd, valid...)
		cmd.Stdout = nil
		cmd.Stderr = nil
		cmd.Stdin = nil
		err := cmd.Start()
		return AudacityLaunchedMsg{Paths: valid, Err: err}
	}
}

// DirectoryOpenedMsg is dispatched when opening a folder in the file manager.
type DirectoryOpenedMsg struct {
	Path string
	Err  error
}

// openDirectoryCmd opens the specified directory in the system file manager using xdg-open.
func openDirectoryCmd(dirPath string) tea.Cmd {
	return func() tea.Msg {
		cmdName := "xdg-open"
		if fm := os.Getenv("FILE_MANAGER"); fm != "" {
			cmdName = fm
		}
		cmd := exec.Command(cmdName, dirPath)
		cmd.Stdout = nil
		cmd.Stderr = nil
		cmd.Stdin = nil
		err := cmd.Start()
		return DirectoryOpenedMsg{Path: dirPath, Err: err}
	}
}

// padFrameCmd pads or trims a raw audio file to slotDur seconds using ffmpeg.
func padFrameCmd(rawPath, paddedPath string, slotDur float64) error {
	dir := filepath.Dir(paddedPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create padded dir: %w", err)
	}
	filter := fmt.Sprintf("apad=whole_dur=%.2f,atrim=0:%.2f", slotDur, slotDur)
	cmd := exec.Command("ffmpeg", "-y", "-i", rawPath, "-af", filter, paddedPath)
	return cmd.Run()
}

// generateSilenceCmd creates a silence audio track for slotDur seconds.
func generateSilenceCmd(paddedPath string, slotDur float64) error {
	dir := filepath.Dir(paddedPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create padded dir: %w", err)
	}
	durStr := fmt.Sprintf("%.2f", slotDur)
	cmd := exec.Command("ffmpeg", "-y", "-f", "lavfi", "-i", "anullsrc=r=24000:cl=mono", "-t", durStr, "-c:a", "pcm_s16le", paddedPath)
	return cmd.Run()
}

// buildConcatFile creates an ffmpeg concat list file from a slice of audio file paths.
func buildConcatFile(concatListFile string, paths []string) error {
	dir := filepath.Dir(concatListFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	f, err := os.Create(concatListFile)
	if err != nil {
		return fmt.Errorf("failed to create concat file: %w", err)
	}
	defer f.Close()

	for _, p := range paths {
		if _, err := os.Stat(p); err != nil {
			return fmt.Errorf("audio file missing: %s", p)
		}
		absPath, err := filepath.Abs(p)
		if err != nil {
			absPath = p
		}
		if _, err := fmt.Fprintf(f, "file '%s'\n", absPath); err != nil {
			return err
		}
	}
	return nil
}

// assembleActCmd concatenates all frames in an Act into an Act master WAV.
func assembleActCmd(act *Act) tea.Cmd {
	return func() tea.Msg {
		if act == nil || len(act.Frames) == 0 {
			return AssembleCompleteMsg{ActID: 0, Err: fmt.Errorf("act has no frames")}
		}

		var filePaths []string
		for _, f := range act.Frames {
			target := f.PaddedPath
			if _, err := os.Stat(target); err != nil {
				target = f.RawPath
			}
			filePaths = append(filePaths, target)
		}

		concatListFile := filepath.Join(filepath.Dir(act.MasterPath), "padded", "concat_list.txt")
		if err := buildConcatFile(concatListFile, filePaths); err != nil {
			return AssembleCompleteMsg{ActID: act.ID, Err: err}
		}

		cmd := exec.Command("ffmpeg", "-y", "-f", "concat", "-safe", "0", "-i", concatListFile, "-c", "copy", act.MasterPath)
		if err := cmd.Run(); err != nil {
			return AssembleCompleteMsg{ActID: act.ID, Err: fmt.Errorf("ffmpeg concat failed: %w", err)}
		}

		dur := GetAudioDuration(act.MasterPath)
		return AssembleCompleteMsg{ActID: act.ID, Path: act.MasterPath, Duration: dur, Err: nil}
	}
}

// assembleGlobalMasterCmd stitches all Act masters into a complete production master.
func assembleGlobalMasterCmd(acts []*Act, outputPath string) tea.Cmd {
	return func() tea.Msg {
		var actMasterPaths []string
		for _, act := range acts {
			if _, err := os.Stat(act.MasterPath); err != nil {
				return AssembleCompleteMsg{ActID: 0, Err: fmt.Errorf("act %d master missing: %s", act.ID, act.MasterPath)}
			}
			actMasterPaths = append(actMasterPaths, act.MasterPath)
		}

		concatListFile := filepath.Join(filepath.Dir(outputPath), "global_concat_list.txt")
		if err := buildConcatFile(concatListFile, actMasterPaths); err != nil {
			return AssembleCompleteMsg{ActID: 0, Err: err}
		}

		cmd := exec.Command("ffmpeg", "-y", "-f", "concat", "-safe", "0", "-i", concatListFile, "-c", "copy", outputPath)
		if err := cmd.Run(); err != nil {
			return AssembleCompleteMsg{ActID: 0, Err: fmt.Errorf("global ffmpeg concat failed: %w", err)}
		}

		dur := GetAudioDuration(outputPath)
		return AssembleCompleteMsg{ActID: 0, Path: outputPath, Duration: dur, Err: nil}
	}
}
