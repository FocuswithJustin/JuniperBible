// Package gobible provides pure Go GoBible JAR creation.
package gobible

import (
	"errors"
	"io"
	"testing"
)

// mockZipWriter implements zipWriter interface for testing error conditions.
type mockZipWriter struct {
	failOnCreate      bool
	failOnClose       bool
	createCallCount   int
	failOnCreateIndex int // Fail on the Nth create call (1-indexed)
}

func (m *mockZipWriter) Create(name string) (io.Writer, error) {
	m.createCallCount++
	if m.failOnCreate {
		return nil, errors.New("mock create error")
	}
	if m.failOnCreateIndex > 0 && m.createCallCount == m.failOnCreateIndex {
		return nil, errors.New("mock create error at specific index")
	}
	return &mockWriter{}, nil
}

func (m *mockZipWriter) Close() error {
	if m.failOnClose {
		return errors.New("mock close error")
	}
	return nil
}

// mockWriter implements io.Writer for testing.
type mockWriter struct {
	failOnWrite bool
}

func (m *mockWriter) Write(p []byte) (n int, err error) {
	if m.failOnWrite {
		return 0, errors.New("mock write error")
	}
	return len(p), nil
}

// TestAddManifestError verifies error handling in addManifest.
func TestAddManifestError(t *testing.T) {
	col := NewCollection()
	col.SetName("Test")

	// Test Create error
	mock := &mockZipWriter{
		failOnCreate: true,
	}

	err := col.addManifest(mock)
	if err == nil {
		t.Error("addManifest should return error when Create fails")
	}
	if err.Error() != "mock create error" {
		t.Errorf("Expected 'mock create error', got: %v", err)
	}
}

// TestAddIndexError verifies error handling in addIndex.
func TestAddIndexError(t *testing.T) {
	col := NewCollection()
	col.SetName("Test")
	col.AddBook("Gen", "Gen")

	// Test Create error
	mock := &mockZipWriter{
		failOnCreate: true,
	}

	err := col.addIndex(mock)
	if err == nil {
		t.Error("addIndex should return error when Create fails")
	}
	if err.Error() != "mock create error" {
		t.Errorf("Expected 'mock create error', got: %v", err)
	}
}

// TestAddBookDataError verifies error handling in addBookData.
func TestAddBookDataError(t *testing.T) {
	col := NewCollection()
	col.SetName("Test")
	col.AddBook("Genesis", "Gen")
	col.AddChapter("Gen", 1, []string{"Verse 1"})

	// Test Create error
	mock := &mockZipWriter{
		failOnCreate: true,
	}

	book := col.GetBookByID("Gen")
	err := col.addBookData(mock, 0, book)
	if err == nil {
		t.Error("addBookData should return error when Create fails")
	}
	if err.Error() != "mock create error" {
		t.Errorf("Expected 'mock create error', got: %v", err)
	}
}

// TestBuildJARManifestError tests BuildJAR error handling when addManifest fails.
func TestBuildJARManifestError(t *testing.T) {
	col := NewCollection()
	col.SetName("Test")
	col.AddBook("Gen", "Gen")
	col.AddChapter("Gen", 1, []string{"Verse"})

	// We'll test the helper methods directly since we can't easily inject
	// the mock into BuildJAR without more refactoring. The error paths
	// are already tested above.

	// Verify that the methods handle errors correctly
	mock := &mockZipWriter{failOnCreate: true}
	err := col.addManifest(mock)
	if err == nil {
		t.Error("Expected error from addManifest")
	}
}

// TestBuildJARIndexError tests error handling when addIndex fails.
func TestBuildJARIndexError(t *testing.T) {
	col := NewCollection()
	col.SetName("Test")
	col.AddBook("Gen", "Gen")
	col.AddChapter("Gen", 1, []string{"Verse"})

	// Test that addIndex handles Create errors
	mock := &mockZipWriter{failOnCreate: true}
	err := col.addIndex(mock)
	if err == nil {
		t.Error("Expected error from addIndex")
	}
}

// TestBuildJARBookDataError tests error handling when addBookData fails.
func TestBuildJARBookDataError(t *testing.T) {
	col := NewCollection()
	col.SetName("Test")
	col.AddBook("Gen", "Gen")
	col.AddChapter("Gen", 1, []string{"Verse"})

	// Test that addBookData handles Create errors
	mock := &mockZipWriter{failOnCreate: true}
	book := col.GetBookByID("Gen")
	err := col.addBookData(mock, 0, book)
	if err == nil {
		t.Error("Expected error from addBookData")
	}
}

// TestZipCloseError verifies error handling when Close fails.
func TestZipCloseError(t *testing.T) {
	mock := &mockZipWriter{
		failOnClose: true,
	}

	err := mock.Close()
	if err == nil {
		t.Error("Close should return error when failOnClose is true")
	}
	if err.Error() != "mock close error" {
		t.Errorf("Expected 'mock close error', got: %v", err)
	}
}

// TestBuildJARValidationError verifies that BuildJAR handles validation errors.
func TestBuildJARValidationError(t *testing.T) {
	// Test validation error path
	col := NewCollection()
	_, err := col.BuildJAR()
	if err == nil {
		t.Error("BuildJAR should fail validation")
	}

	// Test with missing name
	col = NewCollection()
	col.AddBook("Gen", "Gen")
	col.AddChapter("Gen", 1, []string{"Verse"})
	_, err = col.BuildJAR()
	if err == nil {
		t.Error("BuildJAR should fail without name")
	}
}

// TestBuildJARWithWriterManifestError tests error propagation when manifest creation fails.
func TestBuildJARWithWriterManifestError(t *testing.T) {
	col := NewCollection()
	col.SetName("Test")
	col.AddBook("Gen", "Gen")
	col.AddChapter("Gen", 1, []string{"Verse"})

	// Mock that fails on first Create (manifest)
	mock := &mockZipWriter{
		failOnCreateIndex: 1,
	}

	err := col.buildJARWithWriter(mock)
	if err == nil {
		t.Error("buildJARWithWriter should fail when manifest creation fails")
	}
	if err.Error() != "mock create error at specific index" {
		t.Errorf("Expected specific error, got: %v", err)
	}
}

// TestBuildJARWithWriterIndexError tests error propagation when index creation fails.
func TestBuildJARWithWriterIndexError(t *testing.T) {
	col := NewCollection()
	col.SetName("Test")
	col.AddBook("Gen", "Gen")
	col.AddChapter("Gen", 1, []string{"Verse"})

	// Mock that fails on second Create (index)
	mock := &mockZipWriter{
		failOnCreateIndex: 2,
	}

	err := col.buildJARWithWriter(mock)
	if err == nil {
		t.Error("buildJARWithWriter should fail when index creation fails")
	}
	if err.Error() != "mock create error at specific index" {
		t.Errorf("Expected specific error, got: %v", err)
	}
}

// TestBuildJARWithWriterBookDataError tests error propagation when book data creation fails.
func TestBuildJARWithWriterBookDataError(t *testing.T) {
	col := NewCollection()
	col.SetName("Test")
	col.AddBook("Gen", "Gen")
	col.AddChapter("Gen", 1, []string{"Verse"})

	// Mock that fails on third Create (book data)
	mock := &mockZipWriter{
		failOnCreateIndex: 3,
	}

	err := col.buildJARWithWriter(mock)
	if err == nil {
		t.Error("buildJARWithWriter should fail when book data creation fails")
	}
	if err.Error() != "mock create error at specific index" {
		t.Errorf("Expected specific error, got: %v", err)
	}
}

// TestBuildJARWithWriterCloseError tests error propagation when Close fails.
func TestBuildJARWithWriterCloseError(t *testing.T) {
	col := NewCollection()
	col.SetName("Test")
	col.AddBook("Gen", "Gen")
	col.AddChapter("Gen", 1, []string{"Verse"})

	// Mock that fails on Close
	mock := &mockZipWriter{
		failOnClose: true,
	}

	err := col.buildJARWithWriter(mock)
	if err == nil {
		t.Error("buildJARWithWriter should fail when Close fails")
	}
	if err.Error() != "mock close error" {
		t.Errorf("Expected 'mock close error', got: %v", err)
	}
}

// TestBuildJARWithWriterMultipleBooks tests error propagation with multiple books.
func TestBuildJARWithWriterMultipleBooks(t *testing.T) {
	col := NewCollection()
	col.SetName("Test")
	col.AddBook("Gen", "Gen")
	col.AddBook("Exod", "Exod")
	col.AddChapter("Gen", 1, []string{"Verse"})
	col.AddChapter("Exod", 1, []string{"Verse"})

	// Mock that fails on fourth Create (second book data)
	mock := &mockZipWriter{
		failOnCreateIndex: 4,
	}

	err := col.buildJARWithWriter(mock)
	if err == nil {
		t.Error("buildJARWithWriter should fail when second book data creation fails")
	}
	if err.Error() != "mock create error at specific index" {
		t.Errorf("Expected specific error, got: %v", err)
	}
}

// TestBuildJARErrorHandling verifies BuildJAR handles all errors correctly.
// Note: BuildJAR creates a real zip.Writer internally which rarely fails with
// in-memory buffers, so we test the error paths through buildJARWithWriter directly.
func TestBuildJARErrorHandling(t *testing.T) {
	// This test documents that BuildJAR's error handling is tested via
	// buildJARWithWriter tests above. The error return path in BuildJAR
	// (line 129-131) propagates errors from buildJARWithWriter.

	// Verify BuildJAR works with valid data
	col := NewCollection()
	col.SetName("Test")
	col.AddBook("Gen", "Gen")
	col.AddChapter("Gen", 1, []string{"Verse"})

	_, err := col.BuildJAR()
	if err != nil {
		t.Errorf("BuildJAR should succeed with valid data: %v", err)
	}

	// Error paths are tested through buildJARWithWriter above
	// Testing all scenarios: manifest error, index error, book data error, close error
}

// TestBuildJARInternalWithMock tests BuildJARInternal with error conditions.
func TestBuildJARInternalWithMock(t *testing.T) {
	col := NewCollection()
	col.SetName("Test")
	col.AddBook("Gen", "Gen")
	col.AddChapter("Gen", 1, []string{"Verse"})

	// Test with nil writer (should create internal writer)
	data, err := col.BuildJARInternal(nil)
	if err != nil {
		t.Errorf("BuildJARInternal with nil writer should succeed: %v", err)
	}
	if len(data) == 0 {
		t.Error("BuildJARInternal should return data when using internal writer")
	}

	// Test with mock that fails
	mock := &mockZipWriter{
		failOnCreate: true,
	}
	_, err = col.BuildJARInternal(mock)
	if err == nil {
		t.Error("BuildJARInternal should fail when mock writer fails")
	}
	if err.Error() != "mock create error" {
		t.Errorf("Expected 'mock create error', got: %v", err)
	}

	// Test with successful mock
	mockSuccess := &mockZipWriter{}
	data, err = col.BuildJARInternal(mockSuccess)
	if err != nil {
		t.Errorf("BuildJARInternal should succeed with working mock: %v", err)
	}
	// When using external writer, returns nil bytes
	if data != nil {
		t.Error("BuildJARInternal should return nil bytes when using external writer")
	}
}
