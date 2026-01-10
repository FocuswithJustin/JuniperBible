# Example: Refactoring a Handler to Use Base Package

This document shows how to refactor an existing handler to use the base package.

## Before: Original Handler (Zefania)

```go
// Detect implements EmbeddedFormatHandler.Detect.
func (h *Handler) Detect(path string) (*plugins.DetectResult, error) {
	info, err := os.Stat(path)
	if err != nil {
		return &plugins.DetectResult{Detected: false, Reason: fmt.Sprintf("cannot stat: %v", err)}, nil
	}

	if info.IsDir() {
		return &plugins.DetectResult{Detected: false, Reason: "path is a directory"}, nil
	}

	ext := strings.ToLower(filepath.Ext(path))
	if ext != ".xml" {
		return &plugins.DetectResult{Detected: false, Reason: "not a .xml file"}, nil
	}

	return &plugins.DetectResult{
		Detected: true,
		Format:   "zefania",
		Reason:   "Zefania Bible file detected",
	}, nil
}

// Ingest implements EmbeddedFormatHandler.Ingest.
func (h *Handler) Ingest(path, outputDir string) (*plugins.IngestResult, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	hash := sha256.Sum256(data)
	hashHex := hex.EncodeToString(hash[:])

	blobDir := filepath.Join(outputDir, hashHex[:2])
	if err := os.MkdirAll(blobDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create blob dir: %w", err)
	}

	blobPath := filepath.Join(blobDir, hashHex)
	if err := os.WriteFile(blobPath, data, 0644); err != nil {
		return nil, fmt.Errorf("failed to write blob: %w", err)
	}

	artifactID := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	return &plugins.IngestResult{
		ArtifactID: artifactID,
		BlobSHA256: hashHex,
		SizeBytes:  int64(len(data)),
		Metadata:   map[string]string{"format": "zefania"},
	}, nil
}

// Enumerate implements EmbeddedFormatHandler.Enumerate.
func (h *Handler) Enumerate(path string) (*plugins.EnumerateResult, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat: %w", err)
	}

	return &plugins.EnumerateResult{
		Entries: []plugins.EnumerateEntry{
			{Path: filepath.Base(path), SizeBytes: info.Size(), IsDir: false},
		},
	}, nil
}
```

## After: Using Base Package

```go
import (
	"github.com/FocuswithJustin/mimicry/internal/formats/base"
	"github.com/FocuswithJustin/mimicry/core/plugins"
)

// Detect implements EmbeddedFormatHandler.Detect.
func (h *Handler) Detect(path string) (*plugins.DetectResult, error) {
	return base.DetectFile(path, base.DetectConfig{
		Extensions: []string{".xml"},
		FormatName: "zefania",
	})
}

// Ingest implements EmbeddedFormatHandler.Ingest.
func (h *Handler) Ingest(path, outputDir string) (*plugins.IngestResult, error) {
	return base.IngestFile(path, outputDir, base.IngestConfig{
		FormatName: "zefania",
	})
}

// Enumerate implements EmbeddedFormatHandler.Enumerate.
func (h *Handler) Enumerate(path string) (*plugins.EnumerateResult, error) {
	return base.EnumerateFile(path, map[string]string{
		"format": "zefania",
	})
}
```

## Result

- **Lines of code reduced**: ~50 lines → ~15 lines (70% reduction)
- **Complexity reduced**: No more manual error handling for common operations
- **Maintainability improved**: Common bugs fixed in one place
- **Readability improved**: Intent is clearer, less boilerplate

## More Complex Example: USFM Handler

### Before (Detect method)

```go
func (h *Handler) Detect(path string) (*plugins.DetectResult, error) {
	info, err := os.Stat(path)
	if err != nil {
		return &plugins.DetectResult{Detected: false, Reason: fmt.Sprintf("cannot stat: %v", err)}, nil
	}

	if info.IsDir() {
		return &plugins.DetectResult{Detected: false, Reason: "path is a directory, not a file"}, nil
	}

	// Read file and check for USFM markers
	data, err := os.ReadFile(path)
	if err != nil {
		return &plugins.DetectResult{Detected: false, Reason: fmt.Sprintf("cannot read: %v", err)}, nil
	}

	content := string(data)

	// Check for USFM markers
	if strings.Contains(content, "\\id ") || strings.Contains(content, "\\c ") ||
		strings.Contains(content, "\\v ") || strings.Contains(content, "\\p") {
		return &plugins.DetectResult{
			Detected: true,
			Format:   "USFM",
			Reason:   "USFM markers detected",
		}, nil
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(path))
	if ext == ".usfm" || ext == ".sfm" || ext == ".ptx" {
		return &plugins.DetectResult{
			Detected: true,
			Format:   "USFM",
			Reason:   "USFM file extension detected",
		}, nil
	}

	return &plugins.DetectResult{Detected: false, Reason: "not a USFM file"}, nil
}
```

### After

```go
func (h *Handler) Detect(path string) (*plugins.DetectResult, error) {
	return base.DetectFile(path, base.DetectConfig{
		Extensions:     []string{".usfm", ".sfm", ".ptx"},
		ContentMarkers: []string{"\\id ", "\\c ", "\\v "},
		FormatName:     "USFM",
	})
}
```

**Lines: 35 → 7 (80% reduction)**

### Before (Ingest method with custom artifact ID)

```go
func (h *Handler) Ingest(path, outputDir string) (*plugins.IngestResult, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	hash := sha256.Sum256(data)
	hashHex := hex.EncodeToString(hash[:])

	blobDir := filepath.Join(outputDir, hashHex[:2])
	if err := os.MkdirAll(blobDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create blob dir: %w", err)
	}

	blobPath := filepath.Join(blobDir, hashHex)
	if err := os.WriteFile(blobPath, data, 0644); err != nil {
		return nil, fmt.Errorf("failed to write blob: %w", err)
	}

	// Parse book ID from \id marker
	artifactID := filepath.Base(path)
	content := string(data)
	if idx := strings.Index(content, "\\id "); idx >= 0 {
		endIdx := strings.IndexAny(content[idx+4:], " \n\r")
		if endIdx > 0 {
			artifactID = strings.TrimSpace(content[idx+4 : idx+4+endIdx])
		}
	}

	return &plugins.IngestResult{
		ArtifactID: artifactID,
		BlobSHA256: hashHex,
		SizeBytes:  int64(len(data)),
		Metadata: map[string]string{
			"original_name": filepath.Base(path),
			"format":        "USFM",
		},
	}, nil
}
```

### After

```go
func (h *Handler) Ingest(path, outputDir string) (*plugins.IngestResult, error) {
	return base.IngestFile(path, outputDir, base.IngestConfig{
		FormatName: "USFM",
		ArtifactIDExtractor: func(path string, data []byte) string {
			content := string(data)
			if idx := strings.Index(content, "\\id "); idx >= 0 {
				endIdx := strings.IndexAny(content[idx+4:], " \n\r")
				if endIdx > 0 {
					return strings.TrimSpace(content[idx+4 : idx+4+endIdx])
				}
			}
			return filepath.Base(path)
		},
	})
}
```

**Lines: 38 → 15 (60% reduction)**

## Benefits Summary

1. **Code Reduction**: 50-80% fewer lines for common operations
2. **Consistency**: All handlers handle errors the same way
3. **Testing**: Common code is tested once, comprehensively
4. **Maintenance**: Improvements benefit all handlers
5. **Focus**: Handler code focuses on format-specific logic
6. **Documentation**: Pattern is documented and standardized

## Migration Strategy

Handlers can be migrated incrementally:

1. Start with simple handlers (Zefania, SWORD, TXT)
2. Move to handlers with custom logic (USFM, OSIS)
3. Keep format-specific parsing/generation logic untouched
4. Test each migration independently
5. Keep old code as comments during migration for verification
