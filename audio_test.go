package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildConcatList(t *testing.T) {
	tempDir := t.TempDir()
	f1 := filepath.Join(tempDir, "f1.wav")
	f2 := filepath.Join(tempDir, "f2.wav")
	_ = os.WriteFile(f1, []byte("audio1"), 0644)
	_ = os.WriteFile(f2, []byte("audio2"), 0644)

	concatFile := filepath.Join(tempDir, "concat.txt")
	err := buildConcatFile(concatFile, []string{f1, f2})
	if err != nil {
		t.Fatalf("buildConcatFile failed: %v", err)
	}

	content, err := os.ReadFile(concatFile)
	if err != nil {
		t.Fatalf("Failed to read concat file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != 2 {
		t.Fatalf("Expected 2 lines in concat file, got %d", len(lines))
	}
	if !strings.HasPrefix(lines[0], "file '") || !strings.HasSuffix(lines[0], "f1.wav'") {
		t.Errorf("Unexpected concat line format: %s", lines[0])
	}
}

func TestBuildConcatList_MissingFile(t *testing.T) {
	tempDir := t.TempDir()
	concatFile := filepath.Join(tempDir, "concat.txt")
	err := buildConcatFile(concatFile, []string{"/nonexistent/path/file.wav"})
	if err == nil {
		t.Fatalf("Expected error for missing audio file in concat list, got nil")
	}
}
