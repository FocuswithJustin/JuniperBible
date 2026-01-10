package ipc

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

type testEvent struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

func TestNewTranscript(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "transcript-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	transcript := NewTranscript(tempDir)
	if transcript == nil {
		t.Fatal("NewTranscript returned nil")
	}
	defer transcript.Close()

	// Verify file was created
	transcriptPath := filepath.Join(tempDir, "transcript.jsonl")
	if _, err := os.Stat(transcriptPath); os.IsNotExist(err) {
		t.Error("transcript file was not created")
	}
}

func TestTranscriptWriteEvent(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "transcript-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	transcript := NewTranscript(tempDir)
	defer transcript.Close()

	// Write test events
	transcript.WriteEvent(testEvent{Type: "start", Message: "Starting test"})
	transcript.WriteEvent(testEvent{Type: "end", Message: "Test complete"})
	transcript.Close()

	// Read and verify
	transcriptPath := filepath.Join(tempDir, "transcript.jsonl")
	data, err := os.ReadFile(transcriptPath)
	if err != nil {
		t.Fatalf("failed to read transcript: %v", err)
	}

	lines := splitLines(data)
	if len(lines) != 2 {
		t.Errorf("expected 2 lines, got %d", len(lines))
	}

	var event1 testEvent
	if err := json.Unmarshal([]byte(lines[0]), &event1); err != nil {
		t.Fatalf("failed to parse first event: %v", err)
	}
	if event1.Type != "start" {
		t.Errorf("expected type 'start', got %q", event1.Type)
	}
}

func TestTranscriptGracefulDegradation(t *testing.T) {
	// Test with non-existent directory - should not panic
	transcript := NewTranscript("/nonexistent/directory/that/does/not/exist")
	defer transcript.Close()

	// WriteEvent should be a no-op, not panic
	transcript.WriteEvent(testEvent{Type: "test", Message: "Should not panic"})
}

func TestTranscriptNilSafe(t *testing.T) {
	// Test that Close is safe to call multiple times
	tempDir, err := os.MkdirTemp("", "transcript-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	transcript := NewTranscript(tempDir)
	transcript.Close()
	transcript.Close() // Should not panic
}

func splitLines(data []byte) []string {
	var lines []string
	start := 0
	for i, b := range data {
		if b == '\n' {
			if i > start {
				lines = append(lines, string(data[start:i]))
			}
			start = i + 1
		}
	}
	if start < len(data) {
		lines = append(lines, string(data[start:]))
	}
	return lines
}
