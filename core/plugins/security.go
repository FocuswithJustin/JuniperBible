// Package plugins provides plugin loading and management for Juniper Bible.
package plugins

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ErrInvalidPluginPath is returned when a plugin path fails security validation.
var ErrInvalidPluginPath = errors.New("invalid plugin path")

// SecurityConfig holds plugin security settings.
type SecurityConfig struct {
	// AllowedPluginDirs is a list of directories where plugins may be loaded from.
	// Empty list means allow plugins from any directory (not recommended for production).
	AllowedPluginDirs []string

	// RequireManifest enforces that all plugins must have a valid plugin.json manifest.
	// Default: true
	RequireManifest bool

	// RestrictToKnownKinds enforces that plugin kinds must be in PluginKinds list.
	// Default: true
	RestrictToKnownKinds bool
}

var (
	// globalSecurityConfig is the active security configuration.
	// By default, it's permissive to maintain backward compatibility.
	globalSecurityConfig = SecurityConfig{
		AllowedPluginDirs:    nil,
		RequireManifest:      true,
		RestrictToKnownKinds: true,
	}
)

// SetSecurityConfig updates the global plugin security configuration.
// This should be called during server initialization before loading any plugins.
func SetSecurityConfig(cfg SecurityConfig) {
	globalSecurityConfig = cfg
}

// GetSecurityConfig returns the current plugin security configuration.
func GetSecurityConfig() SecurityConfig {
	return globalSecurityConfig
}

// ValidatePluginPath validates that a plugin path is safe to execute.
// It checks for:
// - Path traversal attempts (../)
// - Absolute path restrictions (must be in allowed directories)
// - Symbolic link restrictions
// - Executable file existence
func ValidatePluginPath(pluginPath string) error {
	if pluginPath == "" {
		return fmt.Errorf("%w: empty path", ErrInvalidPluginPath)
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(pluginPath)
	if err != nil {
		return fmt.Errorf("%w: failed to resolve absolute path: %v", ErrInvalidPluginPath, err)
	}

	// Check for path traversal attempts in the original path
	if strings.Contains(pluginPath, "..") {
		return fmt.Errorf("%w: path traversal detected", ErrInvalidPluginPath)
	}

	// Check if file exists
	info, err := os.Lstat(absPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("%w: plugin file not found", ErrInvalidPluginPath)
		}
		return fmt.Errorf("%w: failed to stat plugin file: %v", ErrInvalidPluginPath, err)
	}

	// Prevent following symlinks to unauthorized locations
	if info.Mode()&os.ModeSymlink != 0 {
		// Resolve symlink
		realPath, err := filepath.EvalSymlinks(absPath)
		if err != nil {
			return fmt.Errorf("%w: failed to resolve symlink: %v", ErrInvalidPluginPath, err)
		}
		// Validate the real path
		if err := validatePluginDirectory(realPath); err != nil {
			return fmt.Errorf("%w: symlink target failed validation: %v", ErrInvalidPluginPath, err)
		}
	}

	// Ensure file is regular (not device, socket, etc.)
	if !info.Mode().IsRegular() {
		return fmt.Errorf("%w: not a regular file", ErrInvalidPluginPath)
	}

	// Check if path is within allowed directories
	if err := validatePluginDirectory(absPath); err != nil {
		return err
	}

	return nil
}

// validatePluginDirectory checks if a path is within allowed plugin directories.
func validatePluginDirectory(absPath string) error {
	cfg := globalSecurityConfig

	// If no restrictions configured, allow any path
	if len(cfg.AllowedPluginDirs) == 0 {
		return nil
	}

	// Check if path is within any allowed directory
	for _, allowedDir := range cfg.AllowedPluginDirs {
		// Convert allowed dir to absolute path
		absAllowedDir, err := filepath.Abs(allowedDir)
		if err != nil {
			continue
		}

		// Check if plugin path is within this allowed directory
		relPath, err := filepath.Rel(absAllowedDir, absPath)
		if err != nil {
			continue
		}

		// If relPath doesn't start with "..", it's within the allowed directory
		if !strings.HasPrefix(relPath, "..") {
			return nil
		}
	}

	return fmt.Errorf("%w: path not in allowed plugin directories", ErrInvalidPluginPath)
}

// ValidatePluginManifestSecurity validates plugin manifest for security concerns.
func ValidatePluginManifestSecurity(manifest *PluginManifest) error {
	if manifest == nil {
		return fmt.Errorf("manifest is nil")
	}

	// Check if manifest is required
	if globalSecurityConfig.RequireManifest && manifest.PluginID == "" {
		return fmt.Errorf("plugin manifest required but missing plugin_id")
	}

	// Validate plugin kind is known
	if globalSecurityConfig.RestrictToKnownKinds {
		validKind := false
		for _, kind := range PluginKinds {
			if manifest.Kind == kind {
				validKind = true
				break
			}
		}
		if !validKind {
			return fmt.Errorf("unknown plugin kind: %s (allowed: %v)", manifest.Kind, PluginKinds)
		}
	}

	// Validate entrypoint doesn't contain path traversal
	if strings.Contains(manifest.Entrypoint, "..") {
		return fmt.Errorf("entrypoint contains path traversal")
	}

	return nil
}

// SecureEntrypointPath returns the validated full path to a plugin's entrypoint.
// This should be used instead of Plugin.EntrypointPath() when security is a concern.
func (p *Plugin) SecureEntrypointPath() (string, error) {
	if p.Manifest == nil {
		return "", fmt.Errorf("plugin has no manifest")
	}

	// Validate manifest security
	if err := ValidatePluginManifestSecurity(p.Manifest); err != nil {
		return "", fmt.Errorf("manifest validation failed: %w", err)
	}

	// Construct entrypoint path
	entrypoint := filepath.Join(p.Path, p.Manifest.Entrypoint)

	// Validate the entrypoint path
	if err := ValidatePluginPath(entrypoint); err != nil {
		return "", fmt.Errorf("entrypoint validation failed: %w", err)
	}

	return entrypoint, nil
}
