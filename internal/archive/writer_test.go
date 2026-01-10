package archive

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestCreateTarGz(t *testing.T) {
	tempDir := t.TempDir()

	// Create source directory structure
	srcDir := filepath.Join(tempDir, "src")
	if err := os.MkdirAll(filepath.Join(srcDir, "subdir"), 0755); err != nil {
		t.Fatalf("failed to create source dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "file1.txt"), []byte("content1"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "subdir", "file2.txt"), []byte("content2"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	// Create archive
	dstPath := filepath.Join(tempDir, "output", "test.tar.gz")
	if err := CreateTarGz(srcDir, dstPath, "myarchive", true); err != nil {
		t.Fatalf("CreateTarGz failed: %v", err)
	}

	// Verify archive exists
	if _, err := os.Stat(dstPath); os.IsNotExist(err) {
		t.Error("archive file not created")
	}

	// Verify archive content (directories have trailing slashes)
	files := readTarGzFiles(t, dstPath)
	expected := map[string]bool{
		"myarchive/file1.txt":        false,
		"myarchive/subdir/":          false,
		"myarchive/subdir/file2.txt": false,
	}
	for _, f := range files {
		if _, ok := expected[f]; ok {
			expected[f] = true
		}
	}
	for name, found := range expected {
		if !found {
			t.Errorf("missing file in archive: %s (got: %v)", name, files)
		}
	}
}

func TestCreateTarGz_NoParentDir(t *testing.T) {
	tempDir := t.TempDir()

	// Create source directory
	srcDir := filepath.Join(tempDir, "src")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatalf("failed to create source dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "test.txt"), []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	// Create archive without creating parent dir (should fail if parent doesn't exist)
	dstPath := filepath.Join(tempDir, "nonexistent", "test.tar.gz")
	err := CreateTarGz(srcDir, dstPath, "test", false)
	if err == nil {
		t.Error("expected error when parent directory doesn't exist")
	}
}

func TestCreateTarGz_EmptyDir(t *testing.T) {
	tempDir := t.TempDir()

	// Create empty source directory
	srcDir := filepath.Join(tempDir, "empty")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatalf("failed to create source dir: %v", err)
	}

	// Create archive
	dstPath := filepath.Join(tempDir, "empty.tar.gz")
	if err := CreateTarGz(srcDir, dstPath, "empty", false); err != nil {
		t.Fatalf("CreateTarGz failed: %v", err)
	}

	// Verify archive exists
	if _, err := os.Stat(dstPath); os.IsNotExist(err) {
		t.Error("archive file not created")
	}
}

func TestCreateCapsuleTarGz(t *testing.T) {
	tempDir := t.TempDir()

	// Create source directory
	srcDir := filepath.Join(tempDir, "mycapsule")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatalf("failed to create source dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "manifest.json"), []byte("{}"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	// Create capsule archive
	dstPath := filepath.Join(tempDir, "output", "mycapsule.tar.gz")
	if err := CreateCapsuleTarGz(srcDir, dstPath); err != nil {
		t.Fatalf("CreateCapsuleTarGz failed: %v", err)
	}

	// Verify base dir is derived from srcDir
	files := readTarGzFiles(t, dstPath)
	found := false
	for _, f := range files {
		if f == "mycapsule/manifest.json" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected mycapsule/manifest.json in archive")
	}
}

func TestCreateTarGz_NonexistentSource(t *testing.T) {
	tempDir := t.TempDir()

	err := CreateTarGz("/nonexistent/source", filepath.Join(tempDir, "test.tar.gz"), "test", false)
	if err == nil {
		t.Error("expected error for nonexistent source")
	}
}

func TestCreateCapsuleTarGzFromPath(t *testing.T) {
	tempDir := t.TempDir()

	// Create source directory
	srcDir := filepath.Join(tempDir, "srcdata")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatalf("failed to create source dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "manifest.json"), []byte("{}"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	// Create capsule archive - base dir should be derived from dstPath
	dstPath := filepath.Join(tempDir, "mytest.capsule.tar.gz")
	if err := CreateCapsuleTarGzFromPath(srcDir, dstPath); err != nil {
		t.Fatalf("CreateCapsuleTarGzFromPath failed: %v", err)
	}

	// Verify base dir is "mytest" (derived from dstPath, not srcDir)
	files := readTarGzFiles(t, dstPath)
	found := false
	for _, f := range files {
		if f == "mytest/manifest.json" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected mytest/manifest.json in archive, got: %v", files)
	}
}

// readTarGzFiles is a helper to read file names from a tar.gz archive.
func readTarGzFiles(t *testing.T, path string) []string {
	t.Helper()

	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("failed to open archive: %v", err)
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		t.Fatalf("failed to create gzip reader: %v", err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	var files []string
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("failed to read tar header: %v", err)
		}
		files = append(files, header.Name)
	}

	return files
}
