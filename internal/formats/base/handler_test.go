package base

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDetectFile_ExtensionOnly(t *testing.T) {
	// Create a temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := DetectFile(testFile, DetectConfig{
		Extensions: []string{".txt"},
		FormatName: "TXT",
	})
	if err != nil {
		t.Fatal(err)
	}

	if !result.Detected {
		t.Errorf("Expected detection to succeed, got: %s", result.Reason)
	}
	if result.Format != "TXT" {
		t.Errorf("Expected format TXT, got %s", result.Format)
	}
}

func TestDetectFile_WrongExtension(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.xml")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := DetectFile(testFile, DetectConfig{
		Extensions: []string{".txt"},
		FormatName: "TXT",
	})
	if err != nil {
		t.Fatal(err)
	}

	if result.Detected {
		t.Error("Expected detection to fail for wrong extension")
	}
}

func TestDetectFile_Directory(t *testing.T) {
	tmpDir := t.TempDir()

	result, err := DetectFile(tmpDir, DetectConfig{
		Extensions: []string{".txt"},
		FormatName: "TXT",
	})
	if err != nil {
		t.Fatal(err)
	}

	if result.Detected {
		t.Error("Expected detection to fail for directory")
	}
	if !strings.Contains(result.Reason, "directory") {
		t.Errorf("Expected reason to mention directory, got: %s", result.Reason)
	}
}

func TestDetectFile_ContentMarkers(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.usfm")
	content := "\\id GEN\n\\c 1\n\\v 1 In the beginning..."
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := DetectFile(testFile, DetectConfig{
		Extensions:     []string{".usfm"},
		FormatName:     "USFM",
		ContentMarkers: []string{"\\id ", "\\c ", "\\v "},
	})
	if err != nil {
		t.Fatal(err)
	}

	if !result.Detected {
		t.Errorf("Expected detection with content markers, got: %s", result.Reason)
	}
}

func TestDetectFile_CustomValidator(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.json")
	if err := os.WriteFile(testFile, []byte(`{"valid": true}`), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := DetectFile(testFile, DetectConfig{
		Extensions:   []string{".json"},
		FormatName:   "JSON",
		CheckContent: true,
		CustomValidator: func(path string, data []byte) (bool, string, error) {
			if strings.Contains(string(data), `"valid"`) {
				return true, "Valid JSON detected", nil
			}
			return false, "", nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if !result.Detected {
		t.Errorf("Expected custom validator to detect, got: %s", result.Reason)
	}
}

func TestIngestFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := []byte("test content")
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatal(err)
	}

	outputDir := filepath.Join(tmpDir, "output")
	result, err := IngestFile(testFile, outputDir, IngestConfig{
		FormatName: "TXT",
	})
	if err != nil {
		t.Fatal(err)
	}

	if result.ArtifactID != "test" {
		t.Errorf("Expected artifact ID 'test', got %s", result.ArtifactID)
	}
	if result.SizeBytes != int64(len(content)) {
		t.Errorf("Expected size %d, got %d", len(content), result.SizeBytes)
	}
	if result.Metadata["format"] != "TXT" {
		t.Errorf("Expected format TXT, got %s", result.Metadata["format"])
	}

	// Verify blob was written
	blobPath := filepath.Join(outputDir, result.BlobSHA256[:2], result.BlobSHA256)
	if _, err := os.Stat(blobPath); os.IsNotExist(err) {
		t.Error("Expected blob file to exist")
	}
}

func TestIngestFile_CustomArtifactID(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := []byte("\\id CUSTOM")
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatal(err)
	}

	outputDir := filepath.Join(tmpDir, "output")
	result, err := IngestFile(testFile, outputDir, IngestConfig{
		FormatName: "TXT",
		ArtifactIDExtractor: func(path string, data []byte) string {
			if idx := strings.Index(string(data), "\\id "); idx >= 0 {
				return strings.TrimSpace(string(data)[idx+4:])
			}
			return filepath.Base(path)
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if result.ArtifactID != "CUSTOM" {
		t.Errorf("Expected artifact ID 'CUSTOM', got %s", result.ArtifactID)
	}
}

func TestEnumerateFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := []byte("test content")
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatal(err)
	}

	result, err := EnumerateFile(testFile, map[string]string{
		"format": "TXT",
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(result.Entries))
	}

	entry := result.Entries[0]
	if entry.Path != "test.txt" {
		t.Errorf("Expected path 'test.txt', got %s", entry.Path)
	}
	if entry.SizeBytes != int64(len(content)) {
		t.Errorf("Expected size %d, got %d", len(content), entry.SizeBytes)
	}
	if entry.IsDir {
		t.Error("Expected IsDir to be false")
	}
	if entry.Metadata["format"] != "TXT" {
		t.Errorf("Expected format TXT, got %s", entry.Metadata["format"])
	}
}

func TestReadFileInfo(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := []byte("test content")
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatal(err)
	}

	info, err := ReadFileInfo(testFile)
	if err != nil {
		t.Fatal(err)
	}

	if info.Path != testFile {
		t.Errorf("Expected path %s, got %s", testFile, info.Path)
	}
	if string(info.Data) != string(content) {
		t.Errorf("Expected data %s, got %s", content, info.Data)
	}
	if info.Size != int64(len(content)) {
		t.Errorf("Expected size %d, got %d", len(content), info.Size)
	}
	if info.Extension != ".txt" {
		t.Errorf("Expected extension .txt, got %s", info.Extension)
	}
	if info.Hash == "" {
		t.Error("Expected hash to be computed")
	}
}

func TestWriteOutput(t *testing.T) {
	tmpDir := t.TempDir()
	content := []byte("output content")

	outputPath, err := WriteOutput(tmpDir, "output.json", content)
	if err != nil {
		t.Fatal(err)
	}

	expectedPath := filepath.Join(tmpDir, "output.json")
	if outputPath != expectedPath {
		t.Errorf("Expected path %s, got %s", expectedPath, outputPath)
	}

	// Verify file was written
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string(content) {
		t.Errorf("Expected content %s, got %s", content, data)
	}
}

func TestUnsupportedOperationError(t *testing.T) {
	err := UnsupportedOperationError("IR extraction", "SWORD")
	expected := "SWORD format does not support IR extraction"
	if err.Error() != expected {
		t.Errorf("Expected error %s, got %s", expected, err.Error())
	}
}
