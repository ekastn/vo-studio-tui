package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)


func TestRenderFrameAPICall_Success(t *testing.T) {
	tempDir := t.TempDir()
	outWav := filepath.Join(tempDir, "output.wav")

	var receivedToken string
	var receivedProfile string
	var receivedText string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedToken = r.Header.Get("Authorization")
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		receivedProfile = r.FormValue("profile_id")
		receivedText = r.FormValue("text")

		// Return synthetic audio payload > 1000 bytes
		payload := bytes.Repeat([]byte{0x00, 0x01}, 600)
		w.WriteHeader(http.StatusOK)
		w.Write(payload)
	}))
	defer server.Close()

	cfg := VoiceConfig{
		APIURL:    server.URL,
		APIToken:  "secret_token_123",
		ProfileID: "voice_prof_abc",
		Language:  "Indonesian",
		Instruct:  "calm voice",
		Speed:     "0.85",
	}

	err := renderFrameAPICall(cfg, "Testing voice generation", outWav)
	if err != nil {
		t.Fatalf("renderFrameAPICall failed: %v", err)
	}

	if receivedToken != "Bearer secret_token_123" {
		t.Errorf("Expected token 'Bearer secret_token_123', got '%s'", receivedToken)
	}
	if receivedProfile != "voice_prof_abc" {
		t.Errorf("Expected profile 'voice_prof_abc', got '%s'", receivedProfile)
	}
	if receivedText != "Testing voice generation" {
		t.Errorf("Expected text 'Testing voice generation', got '%s'", receivedText)
	}
}

func TestRenderFrameAPICall_ErrorResponse(t *testing.T) {
	tempDir := t.TempDir()
	outWav := filepath.Join(tempDir, "output.wav")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Invalid profile ID", http.StatusBadRequest)
	}))
	defer server.Close()

	cfg := VoiceConfig{
		APIURL:    server.URL,
		APIToken:  "token",
		ProfileID: "invalid",
	}

	err := renderFrameAPICall(cfg, "Text", outWav)
	if err == nil {
		t.Fatalf("Expected error for 400 response, got nil")
	}
}
