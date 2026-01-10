package ipc

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStringArg(t *testing.T) {
	tests := []struct {
		name    string
		args    map[string]interface{}
		key     string
		want    string
		wantErr bool
	}{
		{
			name:    "valid string",
			args:    map[string]interface{}{"path": "/test"},
			key:     "path",
			want:    "/test",
			wantErr: false,
		},
		{
			name:    "missing key",
			args:    map[string]interface{}{},
			key:     "path",
			wantErr: true,
		},
		{
			name:    "wrong type",
			args:    map[string]interface{}{"path": 123},
			key:     "path",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := StringArg(tt.args, tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("StringArg() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("StringArg() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStringArgOr(t *testing.T) {
	tests := []struct {
		name       string
		args       map[string]interface{}
		key        string
		defaultVal string
		want       string
	}{
		{
			name:       "present",
			args:       map[string]interface{}{"format": "json"},
			key:        "format",
			defaultVal: "xml",
			want:       "json",
		},
		{
			name:       "missing",
			args:       map[string]interface{}{},
			key:        "format",
			defaultVal: "xml",
			want:       "xml",
		},
		{
			name:       "wrong type",
			args:       map[string]interface{}{"format": 123},
			key:        "format",
			defaultVal: "xml",
			want:       "xml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StringArgOr(tt.args, tt.key, tt.defaultVal)
			if got != tt.want {
				t.Errorf("StringArgOr() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBoolArg(t *testing.T) {
	tests := []struct {
		name       string
		args       map[string]interface{}
		key        string
		defaultVal bool
		want       bool
	}{
		{
			name:       "present true",
			args:       map[string]interface{}{"verbose": true},
			key:        "verbose",
			defaultVal: false,
			want:       true,
		},
		{
			name:       "present false",
			args:       map[string]interface{}{"verbose": false},
			key:        "verbose",
			defaultVal: true,
			want:       false,
		},
		{
			name:       "missing",
			args:       map[string]interface{}{},
			key:        "verbose",
			defaultVal: true,
			want:       true,
		},
		{
			name:       "wrong type",
			args:       map[string]interface{}{"verbose": "yes"},
			key:        "verbose",
			defaultVal: false,
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BoolArg(tt.args, tt.key, tt.defaultVal)
			if got != tt.want {
				t.Errorf("BoolArg() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPathAndOutputDir(t *testing.T) {
	tests := []struct {
		name      string
		args      map[string]interface{}
		wantPath  string
		wantDir   string
		wantErr   bool
		errPrefix string
	}{
		{
			name:     "valid",
			args:     map[string]interface{}{"path": "/input", "output_dir": "/output"},
			wantPath: "/input",
			wantDir:  "/output",
		},
		{
			name:      "missing path",
			args:      map[string]interface{}{"output_dir": "/output"},
			wantErr:   true,
			errPrefix: "path",
		},
		{
			name:      "missing output_dir",
			args:      map[string]interface{}{"path": "/input"},
			wantErr:   true,
			errPrefix: "output_dir",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, dir, err := PathAndOutputDir(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("PathAndOutputDir() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if path != tt.wantPath {
					t.Errorf("path = %v, want %v", path, tt.wantPath)
				}
				if dir != tt.wantDir {
					t.Errorf("outputDir = %v, want %v", dir, tt.wantDir)
				}
			}
		})
	}
}

func TestStoreBlob(t *testing.T) {
	tmpDir := t.TempDir()

	data := []byte("test content")
	hashHex, err := StoreBlob(tmpDir, data)
	if err != nil {
		t.Fatalf("StoreBlob() error = %v", err)
	}

	// Verify hash is correct length (SHA256 = 64 hex chars)
	if len(hashHex) != 64 {
		t.Errorf("hash length = %d, want 64", len(hashHex))
	}

	// Verify file exists
	blobPath := filepath.Join(tmpDir, hashHex[:2], hashHex)
	if _, err := os.Stat(blobPath); err != nil {
		t.Errorf("blob file not found: %v", err)
	}

	// Verify content
	got, err := os.ReadFile(blobPath)
	if err != nil {
		t.Fatalf("failed to read blob: %v", err)
	}
	if string(got) != string(data) {
		t.Errorf("blob content = %v, want %v", string(got), string(data))
	}
}

func TestStoreBlobInvalidDir(t *testing.T) {
	// Try to store in a path that doesn't exist and can't be created
	_, err := StoreBlob("/nonexistent/readonly/path", []byte("test"))
	if err == nil {
		t.Error("expected error for invalid directory")
	}
}

func TestArtifactIDFromPath(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"/path/to/file.zip", "file"},
		{"/path/to/file.tar.gz", "file.tar"},
		{"file", "file"},
		{"/path/to/file", "file"},
		{"test.xml", "test"},
		{".gitignore", ".gitignore"},          // hidden file without extension
		{".env", ".env"},                      // hidden file without extension
		{"/path/to/.gitignore", ".gitignore"}, // hidden file in path
		{".config.yaml", ".config"},           // hidden file with extension
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := ArtifactIDFromPath(tt.path)
			if got != tt.want {
				t.Errorf("ArtifactIDFromPath(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}
