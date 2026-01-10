package docgen

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewGenerator(t *testing.T) {
	g := NewGenerator("/plugins", "/output")

	if g.PluginDir != "/plugins" {
		t.Errorf("PluginDir = %q, want %q", g.PluginDir, "/plugins")
	}
	if g.OutputDir != "/output" {
		t.Errorf("OutputDir = %q, want %q", g.OutputDir, "/output")
	}
}

func TestLoadPlugins(t *testing.T) {
	// Use the actual plugins directory
	wd, _ := os.Getwd()
	pluginDir := filepath.Join(wd, "..", "..", "plugins")

	g := NewGenerator(pluginDir, t.TempDir())
	plugins, err := g.LoadPlugins()
	if err != nil {
		t.Fatalf("LoadPlugins failed: %v", err)
	}

	// Should find at least some plugins
	if len(plugins) == 0 {
		t.Error("LoadPlugins returned no plugins")
	}

	// Should have both format and tool plugins
	hasFormat := false
	hasTool := false
	for _, p := range plugins {
		if p.Kind == "format" {
			hasFormat = true
		}
		if p.Kind == "tool" {
			hasTool = true
		}
	}

	if !hasFormat {
		t.Error("No format plugins found")
	}
	if !hasTool {
		t.Error("No tool plugins found")
	}
}

func TestWritePluginsDoc(t *testing.T) {
	g := NewGenerator("", "")

	plugins := []PluginManifest{
		{
			PluginID:    "format-osis",
			Version:     "1.0.0",
			Kind:        "format",
			Description: "OSIS XML format",
			LossClass:   "L0",
			Extensions:  []string{".osis", ".xml"},
			IRSupport: &IRSupport{
				CanExtract: true,
				CanEmit:    true,
				LossClass:  "L0",
			},
		},
		{
			PluginID:    "tools.libsword",
			Version:     "1.1.0",
			Kind:        "tool",
			Description: "SWORD module operations",
			Requires:    []string{"diatheke", "mod2osis"},
			Profiles: []Profile{
				{ID: "list-modules", Description: "List installed modules"},
				{ID: "render-verse", Description: "Render a specific verse"},
			},
		},
	}

	var buf bytes.Buffer
	err := g.writePluginsDoc(&buf, plugins)
	if err != nil {
		t.Fatalf("writePluginsDoc failed: %v", err)
	}

	output := buf.String()

	// Check for expected content
	checks := []string{
		"# Plugin Catalog",
		"## Format Plugins",
		"format-osis",
		"L0",
		"## Tool Plugins",
		"tools.libsword",
		"list-modules",
	}

	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Errorf("Output missing %q", check)
		}
	}
}

func TestWriteFormatsDoc(t *testing.T) {
	g := NewGenerator("", "")

	plugins := []PluginManifest{
		{
			PluginID:    "format-osis",
			Kind:        "format",
			LossClass:   "L0",
			Description: "OSIS XML",
			IRSupport:   &IRSupport{CanExtract: true, CanEmit: true},
		},
		{
			PluginID:    "format-txt",
			Kind:        "format",
			LossClass:   "L3",
			Description: "Plain text",
			IRSupport:   &IRSupport{CanExtract: true, CanEmit: true},
		},
	}

	var buf bytes.Buffer
	err := g.writeFormatsDoc(&buf, plugins)
	if err != nil {
		t.Fatalf("writeFormatsDoc failed: %v", err)
	}

	output := buf.String()

	// Check for expected content
	checks := []string{
		"# Format Support Matrix",
		"L0",
		"L3",
		"OSIS XML",
		"Plain text",
		"Format Conversion",
	}

	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Errorf("Output missing %q", check)
		}
	}
}

func TestWriteCLIDoc(t *testing.T) {
	g := NewGenerator("", "")

	var buf bytes.Buffer
	err := g.writeCLIDoc(&buf)
	if err != nil {
		t.Fatalf("writeCLIDoc failed: %v", err)
	}

	output := buf.String()

	// Check for expected sections
	checks := []string{
		"# CLI Reference",
		"## Core Commands",
		"### ingest",
		"### export",
		"## Plugin Commands",
		"## IR Commands",
		"### extract-ir",
		"### convert",
		"## Tool Commands",
		"### run",
		"## Behavioral Testing Commands",
		"## Documentation Commands",
		"### docgen",
	}

	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Errorf("Output missing %q", check)
		}
	}
}

func TestGenerateAll(t *testing.T) {
	// Use the actual plugins directory
	wd, _ := os.Getwd()
	pluginDir := filepath.Join(wd, "..", "..", "plugins")
	outputDir := t.TempDir()

	g := NewGenerator(pluginDir, outputDir)
	err := g.GenerateAll()
	if err != nil {
		t.Fatalf("GenerateAll failed: %v", err)
	}

	// Check that files were created
	expectedFiles := []string{
		"PLUGINS.md",
		"FORMATS.md",
		"CLI_REFERENCE.md",
	}

	for _, file := range expectedFiles {
		path := filepath.Join(outputDir, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("Expected file %q not created", file)
		}
	}
}

func TestGenerateAllOutputDirError(t *testing.T) {
	// Try to create output in a non-writable location
	g := NewGenerator("/tmp", "/dev/null/cannot/create")
	err := g.GenerateAll()
	if err == nil {
		t.Error("Expected error for invalid output dir")
	}
}

func TestLoadManifestNotFound(t *testing.T) {
	_, err := loadManifest("/nonexistent/plugin.json")
	if err == nil {
		t.Error("Expected error for missing manifest")
	}
}

func TestLoadManifestInvalidJSON(t *testing.T) {
	tempDir := t.TempDir()
	manifestPath := filepath.Join(tempDir, "plugin.json")
	os.WriteFile(manifestPath, []byte("{invalid json}"), 0644)

	_, err := loadManifest(manifestPath)
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestLoadPluginsEmptyDir(t *testing.T) {
	tempDir := t.TempDir()
	g := NewGenerator(tempDir, t.TempDir())

	plugins, err := g.LoadPlugins()
	if err != nil {
		t.Fatalf("LoadPlugins failed: %v", err)
	}

	// Should return empty list for empty plugin dir
	if len(plugins) != 0 {
		t.Errorf("Expected 0 plugins, got %d", len(plugins))
	}
}

func TestGeneratePluginsError(t *testing.T) {
	// Test with a directory that exists but has no plugins
	tempDir := t.TempDir()
	outputDir := t.TempDir()

	g := NewGenerator(tempDir, outputDir)
	err := g.GeneratePlugins()
	// Should succeed even with no plugins
	if err != nil {
		t.Errorf("GeneratePlugins failed: %v", err)
	}
}

func TestGenerateFormatsError(t *testing.T) {
	tempDir := t.TempDir()
	outputDir := t.TempDir()

	g := NewGenerator(tempDir, outputDir)
	err := g.GenerateFormats()
	// Should succeed even with no plugins
	if err != nil {
		t.Errorf("GenerateFormats failed: %v", err)
	}
}

func TestGenerateCLIError(t *testing.T) {
	tempDir := t.TempDir()
	outputDir := t.TempDir()

	g := NewGenerator(tempDir, outputDir)
	err := g.GenerateCLI()
	if err != nil {
		t.Errorf("GenerateCLI failed: %v", err)
	}

	// Verify file was created
	cliPath := filepath.Join(outputDir, "CLI_REFERENCE.md")
	if _, err := os.Stat(cliPath); os.IsNotExist(err) {
		t.Error("CLI_REFERENCE.md not created")
	}
}

func TestWritePluginsDocEmptyList(t *testing.T) {
	g := NewGenerator("", "")
	var buf bytes.Buffer
	err := g.writePluginsDoc(&buf, []PluginManifest{})
	if err != nil {
		t.Fatalf("writePluginsDoc failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "# Plugin Catalog") {
		t.Error("Missing header in empty plugin doc")
	}
}

func TestWriteFormatsDocEmptyList(t *testing.T) {
	g := NewGenerator("", "")
	var buf bytes.Buffer
	err := g.writeFormatsDoc(&buf, []PluginManifest{})
	if err != nil {
		t.Fatalf("writeFormatsDoc failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "# Format Support Matrix") {
		t.Error("Missing header in empty formats doc")
	}
}

func TestPluginManifestParsing(t *testing.T) {
	// Create a temporary plugin.json
	tempDir := t.TempDir()
	manifestPath := filepath.Join(tempDir, "plugin.json")

	manifestJSON := `{
  "plugin_id": "test-plugin",
  "version": "1.0.0",
  "kind": "format",
  "entrypoint": "test-plugin",
  "description": "Test plugin",
  "extensions": [".test"],
  "loss_class": "L0",
  "ir_support": {
    "can_extract": true,
    "can_emit": true,
    "loss_class": "L0"
  }
}`

	err := os.WriteFile(manifestPath, []byte(manifestJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to write manifest: %v", err)
	}

	manifest, err := loadManifest(manifestPath)
	if err != nil {
		t.Fatalf("loadManifest failed: %v", err)
	}

	if manifest.PluginID != "test-plugin" {
		t.Errorf("PluginID = %q, want %q", manifest.PluginID, "test-plugin")
	}
	if manifest.Version != "1.0.0" {
		t.Errorf("Version = %q, want %q", manifest.Version, "1.0.0")
	}
	if manifest.LossClass != "L0" {
		t.Errorf("LossClass = %q, want %q", manifest.LossClass, "L0")
	}
	if manifest.IRSupport == nil {
		t.Error("IRSupport is nil")
	} else {
		if !manifest.IRSupport.CanExtract {
			t.Error("IRSupport.CanExtract should be true")
		}
		if !manifest.IRSupport.CanEmit {
			t.Error("IRSupport.CanEmit should be true")
		}
	}
}
