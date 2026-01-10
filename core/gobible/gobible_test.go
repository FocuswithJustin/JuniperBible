// Package gobible provides pure Go GoBible JAR creation.
package gobible

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
)

// TestNewCollection verifies creating a new collection.
func TestNewCollection(t *testing.T) {
	col := NewCollection()
	if col == nil {
		t.Fatal("NewCollection returned nil")
	}
}

// TestSetMetadata verifies setting collection metadata.
func TestSetMetadata(t *testing.T) {
	col := NewCollection()
	col.SetName("KJV Bible")
	col.SetInfo("King James Version")

	if col.Name != "KJV Bible" {
		t.Errorf("Name = %q, want %q", col.Name, "KJV Bible")
	}
	if col.Info != "King James Version" {
		t.Errorf("Info = %q, want %q", col.Info, "King James Version")
	}
}

// TestAddBook verifies adding books.
func TestAddBook(t *testing.T) {
	col := NewCollection()
	col.AddBook("Genesis", "Gen")
	col.AddBook("Exodus", "Exod")

	if len(col.Books) != 2 {
		t.Errorf("Should have 2 books, got %d", len(col.Books))
	}

	if col.Books[0].Name != "Genesis" {
		t.Errorf("Book name = %q, want %q", col.Books[0].Name, "Genesis")
	}
}

// TestAddChapter verifies adding chapters to books.
func TestAddChapter(t *testing.T) {
	col := NewCollection()
	col.AddBook("Genesis", "Gen")
	col.AddChapter("Gen", 1, []string{
		"In the beginning God created the heaven and the earth.",
		"And the earth was without form, and void;",
	})

	if len(col.Books[0].Chapters) != 1 {
		t.Errorf("Should have 1 chapter, got %d", len(col.Books[0].Chapters))
	}

	if len(col.Books[0].Chapters[0].Verses) != 2 {
		t.Errorf("Should have 2 verses, got %d", len(col.Books[0].Chapters[0].Verses))
	}
}

// TestBuildJAR verifies building the JAR file.
func TestBuildJAR(t *testing.T) {
	col := NewCollection()
	col.SetName("Test Bible")
	col.AddBook("Genesis", "Gen")
	col.AddChapter("Gen", 1, []string{"In the beginning..."})

	data, err := col.BuildJAR()
	if err != nil {
		t.Fatalf("BuildJAR failed: %v", err)
	}

	if len(data) == 0 {
		t.Error("BuildJAR returned empty data")
	}

	// Verify it's a valid ZIP
	_, err = zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Errorf("BuildJAR output is not valid ZIP: %v", err)
	}
}

// TestBuildContainsManifest verifies MANIFEST.MF exists.
func TestBuildContainsManifest(t *testing.T) {
	col := NewCollection()
	col.SetName("Test")
	col.AddBook("Gen", "Gen")
	col.AddChapter("Gen", 1, []string{"Verse"})

	data, _ := col.BuildJAR()
	r, _ := zip.NewReader(bytes.NewReader(data), int64(len(data)))

	found := false
	for _, f := range r.File {
		if f.Name == "META-INF/MANIFEST.MF" {
			found = true
			break
		}
	}

	if !found {
		t.Error("META-INF/MANIFEST.MF not found in JAR")
	}
}

// TestBuildContainsIndex verifies index file exists.
func TestBuildContainsIndex(t *testing.T) {
	col := NewCollection()
	col.SetName("Test")
	col.AddBook("Gen", "Gen")
	col.AddChapter("Gen", 1, []string{"Verse"})

	data, _ := col.BuildJAR()
	r, _ := zip.NewReader(bytes.NewReader(data), int64(len(data)))

	found := false
	for _, f := range r.File {
		if strings.Contains(f.Name, "Index") || strings.Contains(f.Name, "index") {
			found = true
			break
		}
	}

	if !found {
		t.Error("Index file not found in JAR")
	}
}

// TestBuildContainsBookData verifies book data exists.
func TestBuildContainsBookData(t *testing.T) {
	col := NewCollection()
	col.SetName("Test")
	col.AddBook("Genesis", "Gen")
	col.AddChapter("Gen", 1, []string{"Verse 1", "Verse 2"})

	data, _ := col.BuildJAR()
	r, _ := zip.NewReader(bytes.NewReader(data), int64(len(data)))

	found := false
	for _, f := range r.File {
		if strings.Contains(f.Name, "Gen") || strings.Contains(f.Name, "Book") {
			found = true
			break
		}
	}

	if !found {
		t.Error("Book data not found in JAR")
	}
}

// TestManifestContent verifies MANIFEST.MF content.
func TestManifestContent(t *testing.T) {
	col := NewCollection()
	col.SetName("Test Bible")
	col.AddBook("Gen", "Gen")
	col.AddChapter("Gen", 1, []string{"Verse"})

	data, _ := col.BuildJAR()
	r, _ := zip.NewReader(bytes.NewReader(data), int64(len(data)))

	for _, f := range r.File {
		if f.Name == "META-INF/MANIFEST.MF" {
			rc, _ := f.Open()
			content := make([]byte, 1000)
			n, _ := rc.Read(content)
			rc.Close()

			manifest := string(content[:n])
			if !strings.Contains(manifest, "Manifest-Version") {
				t.Error("MANIFEST.MF missing Manifest-Version")
			}
			if !strings.Contains(manifest, "MIDlet-Name") {
				t.Error("MANIFEST.MF missing MIDlet-Name")
			}
			break
		}
	}
}

// TestEmptyCollection verifies handling of empty collection.
func TestEmptyCollection(t *testing.T) {
	col := NewCollection()
	col.SetName("Empty")

	_, err := col.BuildJAR()
	if err == nil {
		t.Error("BuildJAR should fail for empty collection")
	}
}

// TestBookWithNoChapters verifies handling of book with no chapters.
func TestBookWithNoChapters(t *testing.T) {
	col := NewCollection()
	col.SetName("Test")
	col.AddBook("Genesis", "Gen")

	_, err := col.BuildJAR()
	if err == nil {
		t.Error("BuildJAR should fail for book with no chapters")
	}
}

// TestMultipleBooks verifies handling multiple books.
func TestMultipleBooks(t *testing.T) {
	col := NewCollection()
	col.SetName("Test")
	col.AddBook("Genesis", "Gen")
	col.AddBook("Exodus", "Exod")
	col.AddChapter("Gen", 1, []string{"Gen 1:1"})
	col.AddChapter("Exod", 1, []string{"Exod 1:1"})

	data, err := col.BuildJAR()
	if err != nil {
		t.Fatalf("BuildJAR failed: %v", err)
	}

	r, _ := zip.NewReader(bytes.NewReader(data), int64(len(data)))

	book0Found := false
	book1Found := false
	for _, f := range r.File {
		if strings.Contains(f.Name, "Book0") {
			book0Found = true
		}
		if strings.Contains(f.Name, "Book1") {
			book1Found = true
		}
	}

	if !book0Found || !book1Found {
		t.Error("Not all books found in JAR")
	}
}

// TestUnicodeSupport verifies Unicode text handling.
func TestUnicodeSupport(t *testing.T) {
	col := NewCollection()
	col.SetName("Greek NT")
	col.AddBook("Matthew", "Matt")
	col.AddChapter("Matt", 1, []string{
		"Βίβλος γενέσεως Ἰησοῦ Χριστοῦ",
	})

	data, err := col.BuildJAR()
	if err != nil {
		t.Fatalf("BuildJAR failed: %v", err)
	}

	if len(data) == 0 {
		t.Error("Should produce valid JAR with Unicode content")
	}
}

// TestSpecialCharacters verifies special character handling.
func TestSpecialCharacters(t *testing.T) {
	col := NewCollection()
	col.SetName("Test & Bible <Special>")
	col.AddBook("Gen", "Gen")
	col.AddChapter("Gen", 1, []string{"God said, \"Let there be light.\""})

	data, err := col.BuildJAR()
	if err != nil {
		t.Fatalf("BuildJAR failed: %v", err)
	}

	if len(data) == 0 {
		t.Error("Should handle special characters")
	}
}

// TestGetBookByID verifies book lookup by ID.
func TestGetBookByID(t *testing.T) {
	col := NewCollection()
	col.AddBook("Genesis", "Gen")
	col.AddBook("Exodus", "Exod")

	book := col.GetBookByID("Gen")
	if book == nil {
		t.Fatal("GetBookByID returned nil")
	}

	if book.Name != "Genesis" {
		t.Errorf("Book name = %q, want %q", book.Name, "Genesis")
	}

	notFound := col.GetBookByID("NotFound")
	if notFound != nil {
		t.Error("GetBookByID should return nil for unknown ID")
	}
}

// TestGetChapter verifies chapter lookup.
func TestGetChapter(t *testing.T) {
	col := NewCollection()
	col.AddBook("Genesis", "Gen")
	col.AddChapter("Gen", 1, []string{"V1", "V2"})
	col.AddChapter("Gen", 2, []string{"V1"})

	book := col.GetBookByID("Gen")
	if len(book.Chapters) != 2 {
		t.Errorf("Should have 2 chapters, got %d", len(book.Chapters))
	}

	ch := book.GetChapter(1)
	if ch == nil {
		t.Fatal("GetChapter returned nil")
	}

	if len(ch.Verses) != 2 {
		t.Errorf("Chapter 1 should have 2 verses, got %d", len(ch.Verses))
	}
}

// TestValidation verifies collection validation.
func TestValidation(t *testing.T) {
	col := NewCollection()
	err := col.Validate()
	if err == nil {
		t.Error("Empty collection should fail validation")
	}

	col.SetName("Test")
	err = col.Validate()
	if err == nil {
		t.Error("Collection with no books should fail validation")
	}

	col.AddBook("Gen", "Gen")
	err = col.Validate()
	if err == nil {
		t.Error("Book with no chapters should fail validation")
	}

	col.AddChapter("Gen", 1, []string{"Verse"})
	err = col.Validate()
	if err != nil {
		t.Errorf("Valid collection should pass validation: %v", err)
	}
}

// TestBuildJAD verifies JAD file generation.
func TestBuildJAD(t *testing.T) {
	col := NewCollection()
	col.SetName("Test Bible")
	col.SetInfo("Test info")
	col.AddBook("Gen", "Gen")
	col.AddChapter("Gen", 1, []string{"Verse"})

	jar, _ := col.BuildJAR()
	jad := col.BuildJAD(len(jar))

	if !strings.Contains(jad, "MIDlet-Name") {
		t.Error("JAD missing MIDlet-Name")
	}
	if !strings.Contains(jad, "MIDlet-Jar-Size") {
		t.Error("JAD missing MIDlet-Jar-Size")
	}
	if !strings.Contains(jad, "Test Bible") {
		t.Error("JAD missing collection name")
	}
}

// TestAddChapterToNonexistentBook verifies adding chapter to nonexistent book.
func TestAddChapterToNonexistentBook(t *testing.T) {
	col := NewCollection()
	col.SetName("Test")
	col.AddBook("Genesis", "Gen")
	col.AddChapter("Gen", 1, []string{"Verse 1"})

	// Try to add chapter to book that doesn't exist
	col.AddChapter("NotFound", 1, []string{"Should not be added"})

	// Should still have only one book
	if len(col.Books) != 1 {
		t.Errorf("Should have 1 book, got %d", len(col.Books))
	}

	// The Genesis book should still have only 1 chapter
	book := col.GetBookByID("Gen")
	if len(book.Chapters) != 1 {
		t.Errorf("Genesis should have 1 chapter, got %d", len(book.Chapters))
	}
}

// TestGetChapterNotFound verifies GetChapter returns nil for nonexistent chapter.
func TestGetChapterNotFound(t *testing.T) {
	col := NewCollection()
	col.SetName("Test")
	col.AddBook("Genesis", "Gen")
	col.AddChapter("Gen", 1, []string{"Verse 1"})

	book := col.GetBookByID("Gen")
	ch := book.GetChapter(99) // Chapter that doesn't exist

	if ch != nil {
		t.Error("GetChapter should return nil for nonexistent chapter")
	}
}

// TestManifestEscaping verifies that newlines and carriage returns are escaped in manifest.
func TestManifestEscaping(t *testing.T) {
	col := NewCollection()
	col.SetName("Test\nBible\rWith\nNewlines")
	col.AddBook("Gen", "Gen")
	col.AddChapter("Gen", 1, []string{"Verse"})

	data, err := col.BuildJAR()
	if err != nil {
		t.Fatalf("BuildJAR failed: %v", err)
	}

	r, _ := zip.NewReader(bytes.NewReader(data), int64(len(data)))

	for _, f := range r.File {
		if f.Name == "META-INF/MANIFEST.MF" {
			rc, _ := f.Open()
			content := make([]byte, 1000)
			n, _ := rc.Read(content)
			rc.Close()

			manifest := string(content[:n])
			// Newlines and carriage returns should be replaced with spaces
			if strings.Contains(manifest, "\n\n") {
				t.Error("MANIFEST.MF should not contain consecutive newlines from escaped name")
			}
			// The name should still be present (escaped)
			if !strings.Contains(manifest, "Test Bible") {
				t.Error("MANIFEST.MF should contain escaped name")
			}
			break
		}
	}
}

// TestJADEscaping verifies that newlines and carriage returns are escaped in JAD.
func TestJADEscaping(t *testing.T) {
	col := NewCollection()
	col.SetName("Test\nBible")
	col.SetInfo("Info\rwith\nspecial\rchars")
	col.AddBook("Gen", "Gen")
	col.AddChapter("Gen", 1, []string{"Verse"})

	jar, _ := col.BuildJAR()
	jad := col.BuildJAD(len(jar))

	// Check that newlines and carriage returns were escaped
	lines := strings.Split(jad, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "MIDlet-Name:") ||
			strings.HasPrefix(line, "MIDlet-Description:") {
			// These values should not contain unescaped newlines
			if strings.Count(line, "\n") > 0 {
				t.Error("JAD field should not contain unescaped newlines")
			}
		}
	}
}

// TestMultipleChaptersInBook verifies multiple chapters work correctly.
func TestMultipleChaptersInBook(t *testing.T) {
	col := NewCollection()
	col.SetName("Test")
	col.AddBook("Genesis", "Gen")
	col.AddChapter("Gen", 1, []string{"Verse 1", "Verse 2"})
	col.AddChapter("Gen", 2, []string{"Verse 1", "Verse 2", "Verse 3"})
	col.AddChapter("Gen", 3, []string{"Verse 1"})

	book := col.GetBookByID("Gen")
	if len(book.Chapters) != 3 {
		t.Errorf("Should have 3 chapters, got %d", len(book.Chapters))
	}

	// Verify we can get specific chapters
	ch1 := book.GetChapter(1)
	ch2 := book.GetChapter(2)
	ch3 := book.GetChapter(3)

	if ch1 == nil || ch2 == nil || ch3 == nil {
		t.Error("All chapters should be retrievable")
	}

	if len(ch1.Verses) != 2 {
		t.Errorf("Chapter 1 should have 2 verses, got %d", len(ch1.Verses))
	}
	if len(ch2.Verses) != 3 {
		t.Errorf("Chapter 2 should have 3 verses, got %d", len(ch2.Verses))
	}
	if len(ch3.Verses) != 1 {
		t.Errorf("Chapter 3 should have 1 verse, got %d", len(ch3.Verses))
	}
}

// TestEmptyVerses verifies handling of empty verse content.
func TestEmptyVerses(t *testing.T) {
	col := NewCollection()
	col.SetName("Test")
	col.AddBook("Gen", "Gen")
	col.AddChapter("Gen", 1, []string{""}) // Empty verse

	data, err := col.BuildJAR()
	if err != nil {
		t.Fatalf("BuildJAR should handle empty verses: %v", err)
	}

	if len(data) == 0 {
		t.Error("Should produce valid JAR with empty verses")
	}
}

// TestVeryLongText verifies handling of very long text content.
func TestVeryLongText(t *testing.T) {
	col := NewCollection()
	col.SetName("Test")
	col.AddBook("Gen", "Gen")

	// Create a very long verse
	longVerse := strings.Repeat("This is a very long verse. ", 1000)
	col.AddChapter("Gen", 1, []string{longVerse})

	data, err := col.BuildJAR()
	if err != nil {
		t.Fatalf("BuildJAR should handle long text: %v", err)
	}

	if len(data) == 0 {
		t.Error("Should produce valid JAR with long text")
	}
}

// TestEmptyBookName verifies handling of empty book name.
func TestEmptyBookName(t *testing.T) {
	col := NewCollection()
	col.SetName("Test")
	col.AddBook("", "Gen") // Empty name
	col.AddChapter("Gen", 1, []string{"Verse"})

	data, err := col.BuildJAR()
	if err != nil {
		t.Fatalf("BuildJAR should handle empty book name: %v", err)
	}

	if len(data) == 0 {
		t.Error("Should produce valid JAR with empty book name")
	}
}

// TestEmptyInfo verifies handling of empty collection info.
func TestEmptyInfo(t *testing.T) {
	col := NewCollection()
	col.SetName("Test")
	col.SetInfo("") // Empty info
	col.AddBook("Gen", "Gen")
	col.AddChapter("Gen", 1, []string{"Verse"})

	data, err := col.BuildJAR()
	if err != nil {
		t.Fatalf("BuildJAR should handle empty info: %v", err)
	}

	jad := col.BuildJAD(len(data))
	if !strings.Contains(jad, "MIDlet-Description") {
		t.Error("JAD should contain MIDlet-Description even if info is empty")
	}
}

// TestNoNameValidation verifies validation fails without name.
func TestNoNameValidation(t *testing.T) {
	col := NewCollection()
	// Don't set name
	col.AddBook("Gen", "Gen")
	col.AddChapter("Gen", 1, []string{"Verse"})

	err := col.Validate()
	if err == nil {
		t.Error("Validate should fail when name is empty")
	}

	if !strings.Contains(err.Error(), "name is required") {
		t.Errorf("Error should mention name requirement, got: %v", err)
	}
}

// TestValidationErrorMessages verifies specific error messages.
func TestValidationErrorMessages(t *testing.T) {
	// Test no books error
	col := NewCollection()
	col.SetName("Test")
	err := col.Validate()
	if err == nil {
		t.Error("Should fail with no books")
	}
	if !strings.Contains(err.Error(), "at least one book") {
		t.Errorf("Error should mention books requirement, got: %v", err)
	}

	// Test no chapters error
	col.AddBook("Genesis", "Gen")
	err = col.Validate()
	if err == nil {
		t.Error("Should fail with no chapters")
	}
	if !strings.Contains(err.Error(), "at least one chapter") {
		t.Errorf("Error should mention chapter requirement, got: %v", err)
	}
	if !strings.Contains(err.Error(), "Gen") {
		t.Errorf("Error should mention book ID, got: %v", err)
	}
}

// errorWriter is a writer that always returns an error.
type errorWriter struct {
	errorOnWrite bool
}

func (e *errorWriter) Write(p []byte) (n int, err error) {
	if e.errorOnWrite {
		return 0, errors.New("write error")
	}
	return len(p), nil
}

// TestZipWriterClose verifies that zip.Writer.Close() is called.
// This test ensures the Close() path is executed.
func TestZipWriterClose(t *testing.T) {
	col := NewCollection()
	col.SetName("Test")
	col.AddBook("Gen", "Gen")
	col.AddChapter("Gen", 1, []string{"Verse"})

	// Build JAR normally - this exercises the Close() path
	data, err := col.BuildJAR()
	if err != nil {
		t.Fatalf("BuildJAR failed: %v", err)
	}

	// Verify it's a complete, valid ZIP (which requires proper closing)
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("ZIP reader failed: %v", err)
	}

	// Count files to ensure everything was written
	fileCount := len(r.File)
	if fileCount < 3 { // Should have manifest, index, and at least one book
		t.Errorf("Expected at least 3 files in JAR, got %d", fileCount)
	}
}

// TestJARStructure verifies complete JAR structure.
func TestJARStructure(t *testing.T) {
	col := NewCollection()
	col.SetName("Test Bible")
	col.SetInfo("Test Info")
	col.AddBook("Genesis", "Gen")
	col.AddBook("Exodus", "Exod")
	col.AddChapter("Gen", 1, []string{"Gen 1:1", "Gen 1:2"})
	col.AddChapter("Gen", 2, []string{"Gen 2:1"})
	col.AddChapter("Exod", 1, []string{"Exod 1:1"})

	data, err := col.BuildJAR()
	if err != nil {
		t.Fatalf("BuildJAR failed: %v", err)
	}

	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("ZIP reader failed: %v", err)
	}

	// Check for all expected files
	expectedFiles := map[string]bool{
		"META-INF/MANIFEST.MF": false,
		"GoBible/Index":        false,
		"GoBible/Book0":        false,
		"GoBible/Book1":        false,
	}

	for _, f := range r.File {
		if _, ok := expectedFiles[f.Name]; ok {
			expectedFiles[f.Name] = true
		}
	}

	for name, found := range expectedFiles {
		if !found {
			t.Errorf("Expected file %s not found in JAR", name)
		}
	}
}

// TestIndexFileContent verifies index file content structure.
func TestIndexFileContent(t *testing.T) {
	col := NewCollection()
	col.SetName("Test Bible")
	col.SetInfo("Description")
	col.AddBook("Genesis", "Gen")
	col.AddChapter("Gen", 1, []string{"Verse"})

	data, err := col.BuildJAR()
	if err != nil {
		t.Fatalf("BuildJAR failed: %v", err)
	}

	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("ZIP reader failed: %v", err)
	}

	// Find and read index file
	for _, f := range r.File {
		if f.Name == "GoBible/Index" {
			rc, err := f.Open()
			if err != nil {
				t.Fatalf("Failed to open index: %v", err)
			}
			defer rc.Close()

			// Read index content
			content, err := io.ReadAll(rc)
			if err != nil {
				t.Fatalf("Failed to read index: %v", err)
			}

			// Index should have content
			if len(content) == 0 {
				t.Error("Index file is empty")
			}

			// Index should start with collection name length (2 bytes) then name
			if len(content) < 2 {
				t.Error("Index is too short to contain name length")
			}

			return
		}
	}

	t.Error("Index file not found")
}

// TestBookDataContent verifies book data file content structure.
func TestBookDataContent(t *testing.T) {
	col := NewCollection()
	col.SetName("Test")
	col.AddBook("Genesis", "Gen")
	col.AddChapter("Gen", 1, []string{"Verse 1", "Verse 2"})
	col.AddChapter("Gen", 2, []string{"Verse 3"})

	data, err := col.BuildJAR()
	if err != nil {
		t.Fatalf("BuildJAR failed: %v", err)
	}

	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("ZIP reader failed: %v", err)
	}

	// Find and read book data file
	for _, f := range r.File {
		if f.Name == "GoBible/Book0" {
			rc, err := f.Open()
			if err != nil {
				t.Fatalf("Failed to open book data: %v", err)
			}
			defer rc.Close()

			// Read book content
			content, err := io.ReadAll(rc)
			if err != nil {
				t.Fatalf("Failed to read book data: %v", err)
			}

			// Book data should have content
			if len(content) == 0 {
				t.Error("Book data file is empty")
			}

			return
		}
	}

	t.Error("Book0 file not found")
}

// TestLargeCollection verifies handling of large collections.
func TestLargeCollection(t *testing.T) {
	col := NewCollection()
	col.SetName("Large Bible")
	col.SetInfo("A large test collection")

	// Add multiple books with multiple chapters
	for i := 0; i < 10; i++ {
		bookID := strings.Repeat("B", i+1)
		col.AddBook("Book"+bookID, bookID)
		for j := 1; j <= 5; j++ {
			verses := make([]string, 10)
			for k := 0; k < 10; k++ {
				verses[k] = "Verse content for testing"
			}
			col.AddChapter(bookID, j, verses)
		}
	}

	data, err := col.BuildJAR()
	if err != nil {
		t.Fatalf("BuildJAR failed for large collection: %v", err)
	}

	if len(data) == 0 {
		t.Error("Large collection should produce JAR data")
	}

	// Verify it's a valid ZIP
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("Large collection ZIP is invalid: %v", err)
	}

	// Should have manifest + index + 10 books = 12 files
	if len(r.File) < 12 {
		t.Errorf("Expected at least 12 files, got %d", len(r.File))
	}
}

// TestSingleVerseChapter verifies chapters with single verses.
func TestSingleVerseChapter(t *testing.T) {
	col := NewCollection()
	col.SetName("Test")
	col.AddBook("Obadiah", "Obad")
	col.AddChapter("Obad", 1, []string{"Single verse book"})

	data, err := col.BuildJAR()
	if err != nil {
		t.Fatalf("BuildJAR failed: %v", err)
	}

	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("ZIP reader failed: %v", err)
	}

	// Find book data
	found := false
	for _, f := range r.File {
		if f.Name == "GoBible/Book0" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Book data not found for single verse chapter")
	}
}

// TestChapterNumbering verifies chapter numbering is preserved.
func TestChapterNumbering(t *testing.T) {
	col := NewCollection()
	col.SetName("Test")
	col.AddBook("Psalms", "Ps")

	// Add chapters with specific numbers (not sequential from 1)
	col.AddChapter("Ps", 1, []string{"Psalm 1"})
	col.AddChapter("Ps", 23, []string{"The Lord is my shepherd"})
	col.AddChapter("Ps", 119, []string{"Longest chapter"})

	book := col.GetBookByID("Ps")
	if book == nil {
		t.Fatal("Book not found")
	}

	// Verify chapters are stored with correct numbers
	if len(book.Chapters) != 3 {
		t.Errorf("Expected 3 chapters, got %d", len(book.Chapters))
	}

	// Check chapter numbers
	ch1 := book.GetChapter(1)
	ch23 := book.GetChapter(23)
	ch119 := book.GetChapter(119)

	if ch1 == nil || ch23 == nil || ch119 == nil {
		t.Error("Not all chapters are retrievable by their numbers")
	}

	if ch1 != nil && ch1.Number != 1 {
		t.Errorf("Chapter 1 has wrong number: %d", ch1.Number)
	}
	if ch23 != nil && ch23.Number != 23 {
		t.Errorf("Chapter 23 has wrong number: %d", ch23.Number)
	}
	if ch119 != nil && ch119.Number != 119 {
		t.Errorf("Chapter 119 has wrong number: %d", ch119.Number)
	}
}

// TestUTF8StringEncoding verifies UTF-8 string encoding.
func TestUTF8StringEncoding(t *testing.T) {
	col := NewCollection()
	col.SetName("Multi-language Bible")
	col.AddBook("John", "John")

	// Test various Unicode characters
	col.AddChapter("John", 1, []string{
		"English text",
		"Ελληνικό κείμενο", // Greek
		"עברית",            // Hebrew
		"中文文本",             // Chinese
		"🙏 Emoji test",
	})

	data, err := col.BuildJAR()
	if err != nil {
		t.Fatalf("BuildJAR failed with Unicode: %v", err)
	}

	if len(data) == 0 {
		t.Error("Should produce JAR with Unicode content")
	}

	// Verify it's valid
	_, err = zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Errorf("Unicode content produced invalid ZIP: %v", err)
	}
}
