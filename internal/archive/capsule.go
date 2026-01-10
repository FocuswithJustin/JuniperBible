package archive

import (
	"encoding/json"
	"strings"
)

// CapsuleManifest represents the manifest.json structure in a capsule.
type CapsuleManifest struct {
	Version      string            `json:"version"`
	ModuleType   string            `json:"module_type,omitempty"`
	Title        string            `json:"title,omitempty"`
	Language     string            `json:"language,omitempty"`
	Rights       string            `json:"rights,omitempty"`
	SourceFormat string            `json:"source_format,omitempty"`
	CreatedAt    string            `json:"created_at,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// ExtractCapsuleID extracts the capsule ID from a filename by removing known extensions.
func ExtractCapsuleID(filename string) string {
	// Handle compound extensions first (most specific)
	id := filename
	compoundExts := []string{
		".capsule.tar.xz",
		".capsule.tar.gz",
	}
	for _, ext := range compoundExts {
		if strings.HasSuffix(id, ext) {
			return strings.TrimSuffix(id, ext)
		}
	}

	// Then single extensions
	singleExts := []string{".tar.xz", ".tar.gz", ".tar"}
	for _, ext := range singleExts {
		if strings.HasSuffix(id, ext) {
			return strings.TrimSuffix(id, ext)
		}
	}

	return id
}

// ExtractIRName generates an IR filename from a capsule filename.
func ExtractIRName(filename string) string {
	return ExtractCapsuleID(filename) + ".ir.json"
}

// IsCASCapsule checks if a capsule uses Content-Addressed Storage (has blobs/ directory).
func IsCASCapsule(path string) bool {
	found, _ := ContainsPath(path, func(name string) bool {
		return strings.Contains(name, "blobs/")
	})
	return found
}

// HasIR checks if a capsule contains an IR file (.ir.json).
func HasIR(path string) bool {
	found, _ := ContainsPath(path, func(name string) bool {
		return strings.HasSuffix(name, ".ir.json")
	})
	return found
}

// ReadIR reads the first IR file from a capsule.
func ReadIR(path string) (map[string]interface{}, error) {
	content, _, err := FindFile(path, func(name string) bool {
		return strings.HasSuffix(name, ".ir.json")
	})
	if err != nil {
		return nil, err
	}

	var ir map[string]interface{}
	if err := json.Unmarshal(content, &ir); err != nil {
		return nil, err
	}
	return ir, nil
}

// DetectFormat detects the archive format from the file extension.
func DetectFormat(path string) string {
	switch {
	case strings.HasSuffix(path, ".tar.xz"):
		return "tar.xz"
	case strings.HasSuffix(path, ".tar.gz"):
		return "tar.gz"
	case strings.HasSuffix(path, ".tar"):
		return "tar"
	default:
		return "unknown"
	}
}

// IsSupportedFormat returns true if the file has a supported archive extension.
func IsSupportedFormat(path string) bool {
	return strings.HasSuffix(path, ".tar.xz") ||
		strings.HasSuffix(path, ".tar.gz") ||
		strings.HasSuffix(path, ".tar")
}
