package ipc

import (
	"testing"
)

func TestComputeHash(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
		{"hello", "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"},
		{"The quick brown fox", "5cac4f980fedc3d3f1f99b4be3472c9b30d56523e632d151237ec9309048bda9"},
	}

	for _, tt := range tests {
		result := ComputeHash([]byte(tt.input))
		if result != tt.expected {
			t.Errorf("ComputeHash(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestComputeSourceHash(t *testing.T) {
	input := []byte("hello world")
	raw, hexStr := ComputeSourceHash(input)

	// Verify hex string matches raw bytes
	expectedHex := "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"
	if hexStr != expectedHex {
		t.Errorf("ComputeSourceHash hex = %q, want %q", hexStr, expectedHex)
	}

	// Verify we can use raw bytes
	if len(raw) != 32 {
		t.Errorf("ComputeSourceHash raw length = %d, want 32", len(raw))
	}
}

// Note: TestArtifactIDFromPath, TestStoreBlob, and TestPathAndOutputDir
// are already in args_test.go
