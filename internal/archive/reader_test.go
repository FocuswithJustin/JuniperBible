package archive

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/ulikunitz/xz"
)

func createTestTarGz(t *testing.T, dir string) string {
	path := filepath.Join(dir, "test.tar.gz")
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create file: %v", err)
	}
	defer f.Close()

	gw := gzip.NewWriter(f)
	tw := tar.NewWriter(gw)

	// Add a file
	content := []byte("hello world")
	if err := tw.WriteHeader(&tar.Header{
		Name: "test/hello.txt",
		Mode: 0644,
		Size: int64(len(content)),
	}); err != nil {
		t.Fatalf("write header: %v", err)
	}
	if _, err := tw.Write(content); err != nil {
		t.Fatalf("write content: %v", err)
	}

	// Add an IR file
	irContent := []byte(`{"test": true}`)
	if err := tw.WriteHeader(&tar.Header{
		Name: "test/bible.ir.json",
		Mode: 0644,
		Size: int64(len(irContent)),
	}); err != nil {
		t.Fatalf("write header: %v", err)
	}
	if _, err := tw.Write(irContent); err != nil {
		t.Fatalf("write content: %v", err)
	}

	tw.Close()
	gw.Close()
	return path
}

func createTestTarXz(t *testing.T, dir string) string {
	path := filepath.Join(dir, "test.tar.xz")
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create file: %v", err)
	}
	defer f.Close()

	xw, err := xz.NewWriter(f)
	if err != nil {
		t.Fatalf("xz writer: %v", err)
	}
	tw := tar.NewWriter(xw)

	// Add blobs directory (CAS indicator)
	if err := tw.WriteHeader(&tar.Header{
		Name:     "capsule/blobs/",
		Mode:     0755,
		Typeflag: tar.TypeDir,
	}); err != nil {
		t.Fatalf("write header: %v", err)
	}

	// Add a blob file
	content := []byte("blob content")
	if err := tw.WriteHeader(&tar.Header{
		Name: "capsule/blobs/sha256/ab/abcd1234",
		Mode: 0644,
		Size: int64(len(content)),
	}); err != nil {
		t.Fatalf("write header: %v", err)
	}
	if _, err := tw.Write(content); err != nil {
		t.Fatalf("write content: %v", err)
	}

	tw.Close()
	xw.Close()
	return path
}

func TestNewReader(t *testing.T) {
	dir := t.TempDir()

	tests := []struct {
		name    string
		setup   func(t *testing.T) string
		wantErr bool
	}{
		{
			name: "tar.gz archive",
			setup: func(t *testing.T) string {
				return createTestTarGz(t, dir)
			},
			wantErr: false,
		},
		{
			name: "tar.xz archive",
			setup: func(t *testing.T) string {
				return createTestTarXz(t, dir)
			},
			wantErr: false,
		},
		{
			name: "unsupported format",
			setup: func(t *testing.T) string {
				path := filepath.Join(dir, "test.zip")
				os.WriteFile(path, []byte("not a tar"), 0644)
				return path
			},
			wantErr: true,
		},
		{
			name: "nonexistent file",
			setup: func(t *testing.T) string {
				return filepath.Join(dir, "nonexistent.tar.gz")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup(t)
			r, err := NewReader(path)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewReader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if r != nil {
				r.Close()
			}
		})
	}
}

func TestReaderIterate(t *testing.T) {
	dir := t.TempDir()
	path := createTestTarGz(t, dir)

	r, err := NewReader(path)
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}
	defer r.Close()

	var files []string
	err = r.Iterate(func(header *tar.Header, _ io.Reader) (bool, error) {
		files = append(files, header.Name)
		return false, nil
	})
	if err != nil {
		t.Errorf("Iterate: %v", err)
	}

	if len(files) != 2 {
		t.Errorf("expected 2 files, got %d: %v", len(files), files)
	}
}

func TestIterateCapsule(t *testing.T) {
	dir := t.TempDir()
	path := createTestTarGz(t, dir)

	var count int
	err := IterateCapsule(path, func(header *tar.Header, _ io.Reader) (bool, error) {
		count++
		return false, nil
	})
	if err != nil {
		t.Errorf("IterateCapsule: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 entries, got %d", count)
	}
}

func TestContainsPath(t *testing.T) {
	dir := t.TempDir()

	tests := []struct {
		name      string
		setup     func(t *testing.T) string
		predicate func(string) bool
		want      bool
	}{
		{
			name: "find IR file",
			setup: func(t *testing.T) string {
				return createTestTarGz(t, dir)
			},
			predicate: func(name string) bool {
				return filepath.Ext(name) == ".json" && filepath.Base(name) != "manifest.json"
			},
			want: true,
		},
		{
			name: "find blobs directory (CAS)",
			setup: func(t *testing.T) string {
				return createTestTarXz(t, dir)
			},
			predicate: func(name string) bool {
				return filepath.Base(name) == "blobs" || filepath.Dir(name) == "blobs"
			},
			want: true,
		},
		{
			name: "file not found",
			setup: func(t *testing.T) string {
				return createTestTarGz(t, dir)
			},
			predicate: func(name string) bool {
				return name == "nonexistent.txt"
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup(t)
			got, err := ContainsPath(path, tt.predicate)
			if err != nil {
				t.Errorf("ContainsPath() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("ContainsPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReadFile(t *testing.T) {
	dir := t.TempDir()
	path := createTestTarGz(t, dir)

	tests := []struct {
		name     string
		filename string
		want     string
		wantErr  bool
	}{
		{
			name:     "read hello.txt",
			filename: "hello.txt",
			want:     "hello world",
			wantErr:  false,
		},
		{
			name:     "read IR file",
			filename: "bible.ir.json",
			want:     `{"test": true}`,
			wantErr:  false,
		},
		{
			name:     "file not found",
			filename: "nonexistent.txt",
			want:     "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReadFile(path, tt.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if string(got) != tt.want {
				t.Errorf("ReadFile() = %q, want %q", string(got), tt.want)
			}
		})
	}
}

func TestFindFile(t *testing.T) {
	dir := t.TempDir()
	path := createTestTarGz(t, dir)

	tests := []struct {
		name      string
		predicate func(string) bool
		wantData  string
		wantErr   bool
	}{
		{
			name: "find by extension",
			predicate: func(name string) bool {
				return filepath.Ext(name) == ".txt"
			},
			wantData: "hello world",
			wantErr:  false,
		},
		{
			name: "find JSON",
			predicate: func(name string) bool {
				return filepath.Ext(name) == ".json"
			},
			wantData: `{"test": true}`,
			wantErr:  false,
		},
		{
			name: "no match",
			predicate: func(name string) bool {
				return filepath.Ext(name) == ".xml"
			},
			wantData: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _, err := FindFile(path, tt.predicate)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if string(got) != tt.wantData {
				t.Errorf("FindFile() = %q, want %q", string(got), tt.wantData)
			}
		})
	}
}

func TestReaderClose(t *testing.T) {
	dir := t.TempDir()
	path := createTestTarGz(t, dir)

	r, err := NewReader(path)
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}

	// Close should not error
	if err := r.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}
}
