package validation

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"
)

func TestSanitizePath(t *testing.T) {
	baseDir := "/tmp/test"

	tests := []struct {
		name      string
		baseDir   string
		userPath  string
		want      string
		wantError error
	}{
		{
			name:      "simple valid path",
			baseDir:   baseDir,
			userPath:  "file.txt",
			want:      "file.txt",
			wantError: nil,
		},
		{
			name:      "nested valid path",
			baseDir:   baseDir,
			userPath:  "subdir/file.txt",
			want:      filepath.Join("subdir", "file.txt"),
			wantError: nil,
		},
		{
			name:      "path with redundant separators",
			baseDir:   baseDir,
			userPath:  "subdir//file.txt",
			want:      filepath.Join("subdir", "file.txt"),
			wantError: nil,
		},
		{
			name:      "path with dot component",
			baseDir:   baseDir,
			userPath:  "./file.txt",
			want:      "file.txt",
			wantError: nil,
		},
		{
			name:      "path traversal with dotdot",
			baseDir:   baseDir,
			userPath:  "../etc/passwd",
			want:      "",
			wantError: ErrPathTraversal,
		},
		{
			name:      "path traversal in middle",
			baseDir:   baseDir,
			userPath:  "subdir/../../etc/passwd",
			want:      "",
			wantError: ErrPathTraversal,
		},
		{
			name:      "absolute path",
			baseDir:   baseDir,
			userPath:  "/etc/passwd",
			want:      "",
			wantError: ErrPathTraversal,
		},
		{
			name:      "empty path",
			baseDir:   baseDir,
			userPath:  "",
			want:      "",
			wantError: ErrEmptyPath,
		},
		{
			name:      "very long path",
			baseDir:   baseDir,
			userPath:  strings.Repeat("a/", 2048) + "file.txt",
			want:      "",
			wantError: ErrPathTooLong,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SanitizePath(tt.baseDir, tt.userPath)

			if tt.wantError != nil {
				if err == nil {
					t.Errorf("SanitizePath() expected error %v, got nil", tt.wantError)
					return
				}
				if !errors.Is(err, tt.wantError) && !strings.Contains(err.Error(), tt.wantError.Error()) {
					t.Errorf("SanitizePath() error = %v, want %v", err, tt.wantError)
				}
				return
			}

			if err != nil {
				t.Errorf("SanitizePath() unexpected error: %v", err)
				return
			}

			if got != tt.want {
				t.Errorf("SanitizePath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateFilename(t *testing.T) {
	tests := []struct {
		name      string
		filename  string
		wantError error
	}{
		{
			name:      "valid simple filename",
			filename:  "file.txt",
			wantError: nil,
		},
		{
			name:      "valid filename with spaces",
			filename:  "my file.txt",
			wantError: nil,
		},
		{
			name:      "valid filename with special chars",
			filename:  "file_name-2024.tar.gz",
			wantError: nil,
		},
		{
			name:      "empty filename",
			filename:  "",
			wantError: ErrInvalidFilename,
		},
		{
			name:      "dot filename",
			filename:  ".",
			wantError: ErrInvalidFilename,
		},
		{
			name:      "dotdot filename",
			filename:  "..",
			wantError: ErrInvalidFilename,
		},
		{
			name:      "filename with slash",
			filename:  "dir/file.txt",
			wantError: ErrInvalidFilename,
		},
		{
			name:      "filename with backslash",
			filename:  "dir\\file.txt",
			wantError: ErrInvalidFilename,
		},
		{
			name:      "filename with null byte",
			filename:  "file\x00.txt",
			wantError: ErrInvalidFilename,
		},
		{
			name:      "filename with control character",
			filename:  "file\n.txt",
			wantError: ErrInvalidFilename,
		},
		{
			name:      "filename starting with hyphen",
			filename:  "-file.txt",
			wantError: ErrInvalidFilename,
		},
		{
			name:      "too long filename",
			filename:  strings.Repeat("a", 256),
			wantError: ErrFilenameTooLong,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFilename(tt.filename)

			if tt.wantError != nil {
				if err == nil {
					t.Errorf("ValidateFilename() expected error %v, got nil", tt.wantError)
					return
				}
				if !errors.Is(err, tt.wantError) && !strings.Contains(err.Error(), tt.wantError.Error()) {
					t.Errorf("ValidateFilename() error = %v, want %v", err, tt.wantError)
				}
				return
			}

			if err != nil {
				t.Errorf("ValidateFilename() unexpected error: %v", err)
			}
		})
	}
}

func TestIsPathSafe(t *testing.T) {
	baseDir := "/tmp/test"

	tests := []struct {
		name     string
		baseDir  string
		userPath string
		want     bool
	}{
		{
			name:     "safe path",
			baseDir:  baseDir,
			userPath: "file.txt",
			want:     true,
		},
		{
			name:     "safe nested path",
			baseDir:  baseDir,
			userPath: "subdir/file.txt",
			want:     true,
		},
		{
			name:     "unsafe path traversal",
			baseDir:  baseDir,
			userPath: "../etc/passwd",
			want:     false,
		},
		{
			name:     "unsafe absolute path",
			baseDir:  baseDir,
			userPath: "/etc/passwd",
			want:     false,
		},
		{
			name:     "empty path",
			baseDir:  baseDir,
			userPath: "",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsPathSafe(tt.baseDir, tt.userPath)
			if got != tt.want {
				t.Errorf("IsPathSafe() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidatePath(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		wantError error
	}{
		{
			name:      "valid relative path",
			path:      "file.txt",
			wantError: nil,
		},
		{
			name:      "valid absolute path",
			path:      "/tmp/file.txt",
			wantError: nil,
		},
		{
			name:      "valid nested path",
			path:      "dir/subdir/file.txt",
			wantError: nil,
		},
		{
			name:      "empty path",
			path:      "",
			wantError: ErrEmptyPath,
		},
		{
			name:      "path with null byte",
			path:      "file\x00.txt",
			wantError: ErrInvalidCharacter,
		},
		{
			name:      "path with control character",
			path:      "dir/file\n.txt",
			wantError: ErrInvalidCharacter,
		},
		{
			name:      "very long path",
			path:      strings.Repeat("a/", 2048) + "file.txt",
			wantError: ErrPathTooLong,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePath(tt.path)

			if tt.wantError != nil {
				if err == nil {
					t.Errorf("ValidatePath() expected error %v, got nil", tt.wantError)
					return
				}
				if !errors.Is(err, tt.wantError) && !strings.Contains(err.Error(), tt.wantError.Error()) {
					t.Errorf("ValidatePath() error = %v, want %v", err, tt.wantError)
				}
				return
			}

			if err != nil {
				t.Errorf("ValidatePath() unexpected error: %v", err)
			}
		})
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name      string
		filename  string
		want      string
		wantError error
	}{
		{
			name:      "valid filename unchanged",
			filename:  "file.txt",
			want:      "file.txt",
			wantError: nil,
		},
		{
			name:      "filename with leading/trailing spaces",
			filename:  "  file.txt  ",
			want:      "file.txt",
			wantError: nil,
		},
		{
			name:      "filename with slashes replaced",
			filename:  "dir/file.txt",
			want:      "dir_file.txt",
			wantError: nil,
		},
		{
			name:      "filename with backslashes replaced",
			filename:  "dir\\file.txt",
			want:      "dir_file.txt",
			wantError: nil,
		},
		{
			name:      "filename with null byte removed",
			filename:  "file\x00name.txt",
			want:      "filename.txt",
			wantError: nil,
		},
		{
			name:      "filename with control characters removed",
			filename:  "file\nname\r.txt",
			want:      "filename.txt",
			wantError: nil,
		},
		{
			name:      "filename with leading hyphen removed",
			filename:  "-file.txt",
			want:      "file.txt",
			wantError: nil,
		},
		{
			name:      "empty filename",
			filename:  "",
			want:      "",
			wantError: ErrInvalidFilename,
		},
		{
			name:      "filename that becomes empty after sanitization",
			filename:  "---",
			want:      "",
			wantError: ErrInvalidFilename,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SanitizeFilename(tt.filename)

			if tt.wantError != nil {
				if err == nil {
					t.Errorf("SanitizeFilename() expected error %v, got nil", tt.wantError)
					return
				}
				if !errors.Is(err, tt.wantError) && !strings.Contains(err.Error(), tt.wantError.Error()) {
					t.Errorf("SanitizeFilename() error = %v, want %v", err, tt.wantError)
				}
				return
			}

			if err != nil {
				t.Errorf("SanitizeFilename() unexpected error: %v", err)
				return
			}

			if got != tt.want {
				t.Errorf("SanitizeFilename() = %q, want %q", got, tt.want)
			}
		})
	}
}

// Benchmark tests
func BenchmarkSanitizePath(b *testing.B) {
	baseDir := "/tmp/test"
	userPath := "subdir/file.txt"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SanitizePath(baseDir, userPath)
	}
}

func BenchmarkValidateFilename(b *testing.B) {
	filename := "valid_filename.txt"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidateFilename(filename)
	}
}

func BenchmarkSanitizeFilename(b *testing.B) {
	filename := "file-with-special_chars.txt"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SanitizeFilename(filename)
	}
}
