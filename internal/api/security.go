// Package api provides HTTP API handlers with security hardening.
package api

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/FocuswithJustin/JuniperBible/internal/validation"
)

var (
	// ErrPathTraversal is returned when path traversal is detected
	ErrPathTraversal = errors.New("path traversal detected")
	// ErrInvalidPath is returned when the path is invalid
	ErrInvalidPath = errors.New("invalid path")
	// ErrPathOutsideBase is returned when path escapes base directory
	ErrPathOutsideBase = errors.New("path outside allowed directory")
)

// ValidatePath validates a user-supplied path to prevent path traversal attacks.
// This function provides defense in depth by:
// 1. Rejecting paths containing ".." (path traversal attempts)
// 2. Cleaning paths using filepath.Clean to normalize separators and remove redundant elements
// 3. Ensuring paths are within the allowed base directory
// 4. Delegating to validation.SanitizePath for comprehensive validation
//
// Parameters:
//   - baseDir: The base directory that the path must be within
//   - userPath: The user-supplied path to validate
//
// Returns:
//   - The cleaned, safe path relative to baseDir
//   - An error if the path is invalid or attempts traversal
//
// Security considerations:
//   - CWE-22: Improper Limitation of a Pathname to a Restricted Directory ('Path Traversal')
//   - OWASP A01:2021: Broken Access Control
func ValidatePath(baseDir, userPath string) (string, error) {
	// Early rejection of empty paths
	if userPath == "" {
		return "", fmt.Errorf("%w: path cannot be empty", ErrInvalidPath)
	}

	// First pass: explicit check for ".." to catch obvious traversal attempts
	// This provides fast rejection before expensive operations
	if strings.Contains(userPath, "..") {
		return "", fmt.Errorf("%w: path contains '..'", ErrPathTraversal)
	}

	// Clean the path to normalize it (remove redundant separators, resolve . and ..)
	// This is critical for preventing obfuscated traversal attempts like:
	// - "foo/./../../bar"
	// - "foo///../bar"
	// - "foo/bar/../../etc/passwd"
	cleanPath := filepath.Clean(userPath)

	// Second check after cleaning - attackers might try encoded or obfuscated paths
	if strings.Contains(cleanPath, "..") {
		return "", fmt.Errorf("%w: path contains '..' after cleaning", ErrPathTraversal)
	}

	// Reject absolute paths - all paths should be relative to baseDir
	if filepath.IsAbs(cleanPath) {
		return "", fmt.Errorf("%w: absolute paths not allowed", ErrInvalidPath)
	}

	// Use the validation package's comprehensive sanitization
	// This checks:
	// - Path length limits
	// - Null bytes and control characters
	// - Verifies path resolves within baseDir after joining
	safePath, err := validation.SanitizePath(baseDir, cleanPath)
	if err != nil {
		// Translate validation errors to our API-specific errors
		if errors.Is(err, validation.ErrPathTraversal) {
			return "", fmt.Errorf("%w: %v", ErrPathTraversal, err)
		}
		return "", fmt.Errorf("%w: %v", ErrInvalidPath, err)
	}

	// Additional verification: build full path and ensure it's within baseDir
	fullPath := filepath.Join(baseDir, safePath)

	// Get absolute paths for comparison
	absBase, err := filepath.Abs(baseDir)
	if err != nil {
		return "", fmt.Errorf("failed to resolve base directory: %w", err)
	}

	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve path: %w", err)
	}

	// Verify the resolved path is within baseDir using path prefix check
	// We use filepath.Rel to check containment rather than string prefix
	// to avoid issues with similar directory names
	relPath, err := filepath.Rel(absBase, absPath)
	if err != nil {
		return "", fmt.Errorf("%w: path resolution failed", ErrPathOutsideBase)
	}

	// If the relative path starts with "..", it's trying to escape
	if strings.HasPrefix(relPath, "..") {
		return "", fmt.Errorf("%w: path escapes base directory", ErrPathOutsideBase)
	}

	return safePath, nil
}

// ValidateID validates a capsule/resource ID to ensure it's safe to use as a filename.
// IDs are used in URL paths and translated to filenames, so they must be carefully validated.
// This is a convenience wrapper around ValidatePath for single-component identifiers.
//
// Security considerations:
//   - Prevents directory traversal via ID parameter
//   - Ensures IDs cannot be ".", "..", or contain path separators
//   - Validates against the same rules as filenames
func ValidateID(id string) error {
	if id == "" {
		return fmt.Errorf("%w: ID cannot be empty", ErrInvalidPath)
	}

	// IDs should never contain path separators
	if strings.ContainsAny(id, "/\\") {
		return fmt.Errorf("%w: ID cannot contain path separators", ErrInvalidPath)
	}

	// Use the validation package's filename validation
	if err := validation.ValidateFilename(id); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidPath, err)
	}

	// Additional check: cleaned version should equal original
	// This catches edge cases where filepath.Clean might normalize something unexpected
	cleaned := filepath.Base(filepath.Clean(id))
	if cleaned != id {
		return fmt.Errorf("%w: ID normalization changed value", ErrInvalidPath)
	}

	return nil
}
