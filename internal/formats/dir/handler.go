// Package dir provides the embedded handler for directory format plugin.
package dir

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/FocuswithJustin/JuniperBible/core/plugins"
)

// Handler implements the EmbeddedFormatHandler interface for directories.
type Handler struct{}

// Manifest returns the plugin manifest for registration.
func Manifest() *plugins.PluginManifest {
	return &plugins.PluginManifest{
		PluginID:   "format.dir",
		Version:    "1.0.0",
		Kind:       "format",
		Entrypoint: "format-dir",
		Capabilities: plugins.Capabilities{
			Inputs:  []string{"directory"},
			Outputs: []string{"artifact.kind:directory"},
		},
	}
}

// Register registers this plugin with the embedded registry.
func Register() {
	plugins.RegisterEmbeddedPlugin(&plugins.EmbeddedPlugin{
		Manifest: Manifest(),
		Format:   &Handler{},
	})
}

func init() {
	Register()
}

// Detect implements EmbeddedFormatHandler.Detect.
func (h *Handler) Detect(path string) (*plugins.DetectResult, error) {
	info, err := os.Stat(path)
	if err != nil {
		return &plugins.DetectResult{Detected: false, Reason: fmt.Sprintf("cannot stat: %v", err)}, nil
	}

	if !info.IsDir() {
		return &plugins.DetectResult{Detected: false, Reason: "not a directory"}, nil
	}

	return &plugins.DetectResult{
		Detected: true,
		Format:   "directory",
		Reason:   "is a directory",
	}, nil
}

// Ingest implements EmbeddedFormatHandler.Ingest.
func (h *Handler) Ingest(path, outputDir string) (*plugins.IngestResult, error) {
	return nil, fmt.Errorf("directory format does not support direct ingest")
}

// Enumerate implements EmbeddedFormatHandler.Enumerate.
func (h *Handler) Enumerate(path string) (*plugins.EnumerateResult, error) {
	var entries []plugins.EnumerateEntry

	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, _ := filepath.Rel(path, filePath)
		if relPath == "." {
			return nil
		}
		entries = append(entries, plugins.EnumerateEntry{
			Path:      relPath,
			SizeBytes: info.Size(),
			IsDir:     info.IsDir(),
		})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	return &plugins.EnumerateResult{Entries: entries}, nil
}

// ExtractIR implements EmbeddedFormatHandler.ExtractIR.
func (h *Handler) ExtractIR(path, outputDir string) (*plugins.ExtractIRResult, error) {
	return nil, fmt.Errorf("directory format does not support IR extraction")
}

// EmitNative implements EmbeddedFormatHandler.EmitNative.
func (h *Handler) EmitNative(irPath, outputDir string) (*plugins.EmitNativeResult, error) {
	return nil, fmt.Errorf("directory format does not support native emission")
}
