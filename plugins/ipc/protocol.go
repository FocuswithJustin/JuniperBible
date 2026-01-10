// Package ipc provides shared IPC protocol types and utilities for plugins.
// This eliminates duplicate definitions across 45+ format plugins.
package ipc

import (
	"encoding/json"
	"fmt"
	"os"
)

// Request is the incoming JSON request from the host.
type Request struct {
	Command string                 `json:"command"`
	Args    map[string]interface{} `json:"args,omitempty"`
}

// Response is the outgoing JSON response to the host.
type Response struct {
	Status string      `json:"status"`
	Result interface{} `json:"result,omitempty"`
	Error  string      `json:"error,omitempty"`
}

// DetectResult is the result of a detect command.
type DetectResult struct {
	Detected bool   `json:"detected"`
	Format   string `json:"format,omitempty"`
	Reason   string `json:"reason,omitempty"`
}

// IngestResult is the result of an ingest command.
type IngestResult struct {
	ArtifactID string            `json:"artifact_id"`
	BlobSHA256 string            `json:"blob_sha256"`
	SizeBytes  int64             `json:"size_bytes"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

// EnumerateResult is the result of an enumerate command.
type EnumerateResult struct {
	Entries []EnumerateEntry `json:"entries"`
}

// EnumerateEntry represents a file entry in enumeration.
type EnumerateEntry struct {
	Path      string            `json:"path"`
	SizeBytes int64             `json:"size_bytes"`
	IsDir     bool              `json:"is_dir"`
	ModTime   string            `json:"mod_time,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// Note: ExtractIRResult, EmitNativeResult, and LossReport moved to results.go
// Note: IR types (Corpus, Document, etc.) moved to ir.go

// ReadRequest reads and decodes an IPC request from stdin.
func ReadRequest() (*Request, error) {
	var req Request
	if err := json.NewDecoder(os.Stdin).Decode(&req); err != nil {
		return nil, fmt.Errorf("failed to decode request: %w", err)
	}
	return &req, nil
}

// Respond writes a success response with the given result to stdout.
// Returns an error if encoding fails.
func Respond(result interface{}) error {
	resp := Response{
		Status: "ok",
		Result: result,
	}
	if err := json.NewEncoder(os.Stdout).Encode(resp); err != nil {
		return fmt.Errorf("failed to encode response: %w", err)
	}
	return nil
}

// RespondError writes an error response to stdout.
// Returns an error if encoding fails. Does NOT exit - caller decides.
// For plugins that need to exit, use RespondErrorAndExit instead.
func RespondError(msg string) error {
	resp := Response{
		Status: "error",
		Error:  msg,
	}
	if err := json.NewEncoder(os.Stdout).Encode(resp); err != nil {
		return fmt.Errorf("failed to encode error response: %w", err)
	}
	return nil
}

// RespondErrorAndExit writes an error response and exits with status 1.
// Use this in main plugin code where exiting is desired.
func RespondErrorAndExit(msg string) {
	resp := Response{
		Status: "error",
		Error:  msg,
	}
	// Best effort - if this fails, we still need to exit
	json.NewEncoder(os.Stdout).Encode(resp)
	os.Exit(1)
}

// RespondErrorf writes a formatted error response to stdout.
// Returns an error if encoding fails. Does NOT exit - caller decides.
func RespondErrorf(format string, args ...interface{}) error {
	return RespondError(fmt.Sprintf(format, args...))
}

// RespondErrorfAndExit writes a formatted error response and exits with status 1.
// Use this in main plugin code where exiting is desired.
func RespondErrorfAndExit(format string, args ...interface{}) {
	RespondErrorAndExit(fmt.Sprintf(format, args...))
}

// MustRespond writes a success response and exits on failure.
func MustRespond(result interface{}) {
	if err := Respond(result); err != nil {
		RespondErrorfAndExit("failed to encode response: %v", err)
	}
}
