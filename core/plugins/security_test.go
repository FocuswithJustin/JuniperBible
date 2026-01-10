package plugins

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidatePluginPath(t *testing.T) {
	// Create temporary directory for test plugins
	tmpDir, err := os.MkdirTemp("", "plugin-security-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a valid plugin file
	validPluginPath := filepath.Join(tmpDir, "test-plugin")
	if err := os.WriteFile(validPluginPath, []byte("#!/bin/sh\necho test"), 0755); err != nil {
		t.Fatalf("failed to create test plugin: %v", err)
	}

	tests := []struct {
		name      string
		path      string
		wantError bool
	}{
		{
			name:      "valid plugin path",
			path:      validPluginPath,
			wantError: false,
		},
		{
			name:      "empty path",
			path:      "",
			wantError: true,
		},
		{
			name:      "path traversal attempt",
			path:      "../etc/passwd",
			wantError: true,
		},
		{
			name:      "non-existent file",
			path:      filepath.Join(tmpDir, "nonexistent"),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePluginPath(tt.path)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidatePluginPath() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestValidatePluginPathWithRestrictions(t *testing.T) {
	// Create temporary directories
	allowedDir, err := os.MkdirTemp("", "allowed-plugins")
	if err != nil {
		t.Fatalf("failed to create allowed dir: %v", err)
	}
	defer os.RemoveAll(allowedDir)

	disallowedDir, err := os.MkdirTemp("", "disallowed-plugins")
	if err != nil {
		t.Fatalf("failed to create disallowed dir: %v", err)
	}
	defer os.RemoveAll(disallowedDir)

	// Create plugin files
	allowedPlugin := filepath.Join(allowedDir, "allowed-plugin")
	if err := os.WriteFile(allowedPlugin, []byte("#!/bin/sh\necho test"), 0755); err != nil {
		t.Fatalf("failed to create allowed plugin: %v", err)
	}

	disallowedPlugin := filepath.Join(disallowedDir, "disallowed-plugin")
	if err := os.WriteFile(disallowedPlugin, []byte("#!/bin/sh\necho test"), 0755); err != nil {
		t.Fatalf("failed to create disallowed plugin: %v", err)
	}

	// Configure security restrictions
	SetSecurityConfig(SecurityConfig{
		AllowedPluginDirs:    []string{allowedDir},
		RequireManifest:      false,
		RestrictToKnownKinds: false,
	})
	defer SetSecurityConfig(SecurityConfig{}) // Reset after test

	tests := []struct {
		name      string
		path      string
		wantError bool
	}{
		{
			name:      "allowed directory",
			path:      allowedPlugin,
			wantError: false,
		},
		{
			name:      "disallowed directory",
			path:      disallowedPlugin,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePluginPath(tt.path)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidatePluginPath() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestValidatePluginManifestSecurity(t *testing.T) {
	// Ensure default config with restrictions
	SetSecurityConfig(SecurityConfig{
		RequireManifest:      true,
		RestrictToKnownKinds: true,
	})
	defer SetSecurityConfig(SecurityConfig{}) // Reset after test

	tests := []struct {
		name      string
		manifest  *PluginManifest
		wantError bool
	}{
		{
			name: "valid manifest",
			manifest: &PluginManifest{
				PluginID:   "test.plugin",
				Version:    "1.0.0",
				Kind:       "format",
				Entrypoint: "format-test",
			},
			wantError: false,
		},
		{
			name:      "nil manifest",
			manifest:  nil,
			wantError: true,
		},
		{
			name: "path traversal in entrypoint",
			manifest: &PluginManifest{
				PluginID:   "test.plugin",
				Version:    "1.0.0",
				Kind:       "format",
				Entrypoint: "../../../etc/passwd",
			},
			wantError: true,
		},
		{
			name: "unknown plugin kind",
			manifest: &PluginManifest{
				PluginID:   "test.plugin",
				Version:    "1.0.0",
				Kind:       "malicious",
				Entrypoint: "test",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePluginManifestSecurity(tt.manifest)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidatePluginManifestSecurity() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestSecureEntrypointPath(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "plugin-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create plugin executable
	pluginPath := filepath.Join(tmpDir, "format-test")
	if err := os.WriteFile(pluginPath, []byte("#!/bin/sh\necho test"), 0755); err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	tests := []struct {
		name      string
		plugin    *Plugin
		wantError bool
	}{
		{
			name: "valid plugin",
			plugin: &Plugin{
				Path: tmpDir,
				Manifest: &PluginManifest{
					PluginID:   "test.plugin",
					Version:    "1.0.0",
					Kind:       "format",
					Entrypoint: "format-test",
				},
			},
			wantError: false,
		},
		{
			name: "no manifest",
			plugin: &Plugin{
				Path:     tmpDir,
				Manifest: nil,
			},
			wantError: true,
		},
		{
			name: "invalid entrypoint",
			plugin: &Plugin{
				Path: tmpDir,
				Manifest: &PluginManifest{
					PluginID:   "test.plugin",
					Version:    "1.0.0",
					Kind:       "format",
					Entrypoint: "../../../etc/passwd",
				},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.plugin.SecureEntrypointPath()
			if (err != nil) != tt.wantError {
				t.Errorf("SecureEntrypointPath() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestSecurityConfigGetSet(t *testing.T) {
	// Save original config
	original := GetSecurityConfig()
	defer SetSecurityConfig(original)

	// Test set and get
	newConfig := SecurityConfig{
		AllowedPluginDirs:    []string{"/tmp/plugins"},
		RequireManifest:      true,
		RestrictToKnownKinds: true,
	}

	SetSecurityConfig(newConfig)
	retrieved := GetSecurityConfig()

	if len(retrieved.AllowedPluginDirs) != len(newConfig.AllowedPluginDirs) {
		t.Errorf("AllowedPluginDirs mismatch: got %v, want %v", retrieved.AllowedPluginDirs, newConfig.AllowedPluginDirs)
	}
	if retrieved.RequireManifest != newConfig.RequireManifest {
		t.Errorf("RequireManifest mismatch: got %v, want %v", retrieved.RequireManifest, newConfig.RequireManifest)
	}
	if retrieved.RestrictToKnownKinds != newConfig.RestrictToKnownKinds {
		t.Errorf("RestrictToKnownKinds mismatch: got %v, want %v", retrieved.RestrictToKnownKinds, newConfig.RestrictToKnownKinds)
	}
}
