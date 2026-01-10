// Package ipc provides shared IPC protocol types and utilities for plugins.

package ipc

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Transcript provides a JSONL transcript writer for tool plugins.
// It writes events to a transcript.jsonl file in the specified output directory.
// The file creation error is intentionally ignored to allow graceful degradation
// when transcript writing fails (e.g., permission issues) - the tool should still
// complete its primary operation.
type Transcript struct {
	file *os.File
	enc  *json.Encoder
}

// NewTranscript creates a new transcript writer in the given output directory.
// If file creation fails, the returned Transcript will silently skip writes.
// This is intentional - transcript failure should not abort tool operations.
func NewTranscript(outDir string) *Transcript {
	path := filepath.Join(outDir, "transcript.jsonl")
	f, _ := os.Create(path) // Intentionally ignore error - graceful degradation
	var enc *json.Encoder
	if f != nil {
		enc = json.NewEncoder(f)
	}
	return &Transcript{file: f, enc: enc}
}

// WriteEvent writes an event to the transcript.
// If the transcript was not successfully created, this is a no-op.
func (t *Transcript) WriteEvent(event interface{}) {
	if t.enc != nil {
		t.enc.Encode(event)
	}
}

// Close closes the transcript file.
// Safe to call even if the transcript was not successfully created.
func (t *Transcript) Close() {
	if t.file != nil {
		t.file.Close()
	}
}
