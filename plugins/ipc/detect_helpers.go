package ipc

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CheckExtension checks if a file has any of the specified extensions.
// Extensions should be provided with dots (e.g., ".xml", ".html").
// Returns true if the file matches, false otherwise.
func CheckExtension(path string, extensions ...string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	for _, validExt := range extensions {
		if ext == strings.ToLower(validExt) {
			return true
		}
	}
	return false
}

// CheckContentContains reads a file and checks if it contains all specified substrings.
// Returns true if all substrings are found, false otherwise.
// If the file cannot be read or is too large (>10MB), returns false.
func CheckContentContains(path string, substrings ...string) bool {
	const maxSize = 10 * 1024 * 1024 // 10MB limit

	info, err := os.Stat(path)
	if err != nil || info.Size() > maxSize {
		return false
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}

	content := string(data)
	for _, substr := range substrings {
		if !strings.Contains(content, substr) {
			return false
		}
	}
	return true
}

// CheckContentContainsAny reads a file and checks if it contains any of the specified substrings.
// Returns true if at least one substring is found, false otherwise.
// If the file cannot be read or is too large (>10MB), returns false.
func CheckContentContainsAny(path string, substrings ...string) bool {
	const maxSize = 10 * 1024 * 1024 // 10MB limit

	info, err := os.Stat(path)
	if err != nil || info.Size() > maxSize {
		return false
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}

	content := string(data)
	for _, substr := range substrings {
		if strings.Contains(content, substr) {
			return true
		}
	}
	return false
}

// CheckMagicBytes reads the first N bytes of a file and compares them to expected magic bytes.
// Returns true if the magic bytes match, false otherwise.
func CheckMagicBytes(path string, magic []byte) bool {
	if len(magic) == 0 {
		return false
	}

	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	header := make([]byte, len(magic))
	n, err := f.Read(header)
	if err != nil || n < len(magic) {
		return false
	}

	return bytes.Equal(header, magic)
}

// DetectByExtension performs standard extension-based detection.
// Returns a DetectResult with appropriate format and reason.
// Extensions should be provided with dots (e.g., ".xml", ".html").
func DetectByExtension(path, formatName string, extensions ...string) *DetectResult {
	if CheckExtension(path, extensions...) {
		extList := strings.Join(extensions, ", ")
		return &DetectResult{
			Detected: true,
			Format:   formatName,
			Reason:   fmt.Sprintf("%s file extension detected (%s)", formatName, extList),
		}
	}
	return &DetectResult{
		Detected: false,
		Reason:   fmt.Sprintf("not a %s file (extension mismatch)", formatName),
	}
}

// DetectByContent performs content-based detection using substring matching.
// Checks for all specified substrings in the file content.
// Returns a DetectResult with appropriate format and reason.
func DetectByContent(path, formatName string, substrings ...string) *DetectResult {
	if CheckContentContains(path, substrings...) {
		return &DetectResult{
			Detected: true,
			Format:   formatName,
			Reason:   fmt.Sprintf("%s format detected (content match)", formatName),
		}
	}
	return &DetectResult{
		Detected: false,
		Reason:   fmt.Sprintf("no %s structure found", formatName),
	}
}

// DetectByContentAny performs content-based detection using substring matching.
// Checks for any of the specified substrings in the file content.
// Returns a DetectResult with appropriate format and reason.
func DetectByContentAny(path, formatName string, substrings ...string) *DetectResult {
	if CheckContentContainsAny(path, substrings...) {
		return &DetectResult{
			Detected: true,
			Format:   formatName,
			Reason:   fmt.Sprintf("%s format detected (content match)", formatName),
		}
	}
	return &DetectResult{
		Detected: false,
		Reason:   fmt.Sprintf("no %s structure found", formatName),
	}
}

// DetectByMagicBytes performs magic byte detection.
// Returns a DetectResult with appropriate format and reason.
func DetectByMagicBytes(path, formatName string, magic []byte) *DetectResult {
	if CheckMagicBytes(path, magic) {
		return &DetectResult{
			Detected: true,
			Format:   formatName,
			Reason:   fmt.Sprintf("%s magic bytes detected", formatName),
		}
	}
	return &DetectResult{
		Detected: false,
		Reason:   fmt.Sprintf("not a %s file (magic bytes mismatch)", formatName),
	}
}

// StandardDetect performs a two-stage detection: extension check followed by content check.
// This is the most common pattern across format plugins.
// First checks the file extension, then validates with content patterns.
// Returns a DetectResult with appropriate format and reason.
func StandardDetect(path, formatName string, extensions []string, contentPatterns []string) *DetectResult {
	// Stage 1: Extension check
	if !CheckExtension(path, extensions...) {
		extList := strings.Join(extensions, ", ")
		return &DetectResult{
			Detected: false,
			Reason:   fmt.Sprintf("not a %s file (expected %s)", formatName, extList),
		}
	}

	// Stage 2: Content validation
	if len(contentPatterns) > 0 {
		if CheckContentContainsAny(path, contentPatterns...) {
			return &DetectResult{
				Detected: true,
				Format:   formatName,
				Reason:   fmt.Sprintf("%s format detected", formatName),
			}
		}
		return &DetectResult{
			Detected: false,
			Reason:   fmt.Sprintf("no %s structure found", formatName),
		}
	}

	// No content validation needed, extension match is sufficient
	return &DetectResult{
		Detected: true,
		Format:   formatName,
		Reason:   fmt.Sprintf("%s file extension detected", formatName),
	}
}

// DetectResult convenience constructors

// DetectSuccess returns a successful detection result.
func DetectSuccess(formatName, reason string) *DetectResult {
	return &DetectResult{
		Detected: true,
		Format:   formatName,
		Reason:   reason,
	}
}

// DetectFailure returns a failed detection result.
func DetectFailure(reason string) *DetectResult {
	return &DetectResult{
		Detected: false,
		Reason:   reason,
	}
}
