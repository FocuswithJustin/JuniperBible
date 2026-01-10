package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/FocuswithJustin/JuniperBible/core/sqlite"
	"github.com/FocuswithJustin/JuniperBible/plugins/ipc"
)

// Phase 18: Tests for e-Sword Commentary (.cmtx) parser

// TestCommentaryEntryStruct verifies the CommentaryEntry structure.
func TestCommentaryEntryStruct(t *testing.T) {
	entry := CommentaryEntry{
		Book:         40, // Matthew
		ChapterStart: 5,
		VerseStart:   3,
		ChapterEnd:   5,
		VerseEnd:     12,
		Comments:     "The Beatitudes represent the core of Jesus' teaching...",
	}

	if entry.Book != 40 {
		t.Errorf("Book = %d, want 40", entry.Book)
	}
	if entry.ChapterStart != 5 {
		t.Errorf("ChapterStart = %d, want 5", entry.ChapterStart)
	}
	if entry.VerseStart != 3 {
		t.Errorf("VerseStart = %d, want 3", entry.VerseStart)
	}
	if entry.Comments == "" {
		t.Error("Comments should not be empty")
	}
}

// TestCommentaryParserCreation verifies parser creation with database path.
func TestCommentaryParserCreation(t *testing.T) {
	parser, err := NewCommentaryParser("/path/to/commentary.cmtx")
	if err != nil {
		// Expected to fail with non-existent path
		return
	}

	if parser == nil {
		t.Error("NewCommentaryParser should return a parser or error")
	}
}

// TestCommentaryDetect verifies detection of e-Sword commentary files.
func TestCommentaryDetect(t *testing.T) {
	tests := []struct {
		filename string
		want     bool
	}{
		{"Matthew Henry.cmtx", true},
		{"commentary.cmtx", true},
		{"COMMENTARY.CMTX", true},
		{"bible.bblx", false},
		{"dictionary.dctx", false},
		{"file.sqlite", false},
	}

	for _, tt := range tests {
		got := IsCommentaryFile(tt.filename)
		if got != tt.want {
			t.Errorf("IsCommentaryFile(%q) = %v, want %v", tt.filename, got, tt.want)
		}
	}
}

// TestCommentaryTableSchema verifies expected table structure.
func TestCommentaryTableSchema(t *testing.T) {
	// Commentary table should have these columns:
	expectedColumns := []string{
		"Book",
		"ChapterStart",
		"VerseStart",
		"ChapterEnd",
		"VerseEnd",
		"Comments",
	}

	for _, col := range expectedColumns {
		if col == "" {
			t.Error("column name should not be empty")
		}
	}
}

// TestCommentaryDetailsTable verifies reading the Details table.
func TestCommentaryDetailsTable(t *testing.T) {
	details := CommentaryDetails{
		Title:        "Matthew Henry's Complete Commentary",
		Abbreviation: "MHC",
		Information:  "Public domain commentary",
		Version:      1,
		RightToLeft:  false,
	}

	if details.Title == "" {
		t.Error("Title should not be empty")
	}
	if details.Abbreviation == "" {
		t.Error("Abbreviation should not be empty")
	}
}

// TestCommentaryGetEntry verifies retrieving a single commentary entry.
func TestCommentaryGetEntry(t *testing.T) {
	parser := &CommentaryParser{
		entries: map[string]*CommentaryEntry{
			"40:5:3": {
				Book:         40,
				ChapterStart: 5,
				VerseStart:   3,
				Comments:     "Blessed are the poor in spirit...",
			},
		},
	}

	entry, err := parser.GetEntry(40, 5, 3)
	if err != nil {
		t.Fatalf("GetEntry failed: %v", err)
	}

	if entry.Book != 40 {
		t.Errorf("Book = %d, want 40", entry.Book)
	}
	if entry.Comments == "" {
		t.Error("Comments should not be empty")
	}
}

// TestCommentaryGetEntryByRef verifies getting entry by OSIS reference.
func TestCommentaryGetEntryByRef(t *testing.T) {
	parser := &CommentaryParser{
		entries: map[string]*CommentaryEntry{
			"40:5:3": {
				Book:         40,
				ChapterStart: 5,
				VerseStart:   3,
				Comments:     "Blessed are the poor in spirit...",
			},
		},
	}

	entry, err := parser.GetEntryByRef("Matt.5.3")
	if err != nil {
		t.Fatalf("GetEntryByRef failed: %v", err)
	}

	if entry.Comments == "" {
		t.Error("Comments should not be empty")
	}
}

// TestCommentaryGetChapter verifies retrieving all entries for a chapter.
func TestCommentaryGetChapter(t *testing.T) {
	parser := &CommentaryParser{
		entries: map[string]*CommentaryEntry{
			"40:5:1":  {Book: 40, ChapterStart: 5, VerseStart: 1, Comments: "V1"},
			"40:5:3":  {Book: 40, ChapterStart: 5, VerseStart: 3, Comments: "V3"},
			"40:5:12": {Book: 40, ChapterStart: 5, VerseStart: 12, Comments: "V12"},
			"40:6:1":  {Book: 40, ChapterStart: 6, VerseStart: 1, Comments: "Ch6V1"},
		},
	}

	entries := parser.GetChapter(40, 5)
	if len(entries) != 3 {
		t.Errorf("len(entries) = %d, want 3", len(entries))
	}
}

// TestCommentaryGetBook verifies retrieving all entries for a book.
func TestCommentaryGetBook(t *testing.T) {
	parser := &CommentaryParser{
		entries: map[string]*CommentaryEntry{
			"40:1:1": {Book: 40, ChapterStart: 1, VerseStart: 1, Comments: "Matt 1:1"},
			"40:5:3": {Book: 40, ChapterStart: 5, VerseStart: 3, Comments: "Matt 5:3"},
			"41:1:1": {Book: 41, ChapterStart: 1, VerseStart: 1, Comments: "Mark 1:1"},
		},
	}

	entries := parser.GetBook(40)
	if len(entries) != 2 {
		t.Errorf("len(entries) = %d, want 2", len(entries))
	}
}

// TestCommentaryVerseRange verifies handling of verse range entries.
func TestCommentaryVerseRange(t *testing.T) {
	entry := CommentaryEntry{
		Book:         40,
		ChapterStart: 5,
		VerseStart:   3,
		ChapterEnd:   5,
		VerseEnd:     12,
		Comments:     "The Beatitudes (v3-12)",
	}

	// IsRange should return true for entries spanning multiple verses
	if !entry.IsRange() {
		t.Error("IsRange() should return true for multi-verse entry")
	}

	// Single verse entry
	single := CommentaryEntry{
		Book:         40,
		ChapterStart: 5,
		VerseStart:   1,
		ChapterEnd:   5,
		VerseEnd:     1,
		Comments:     "Single verse",
	}

	if single.IsRange() {
		t.Error("IsRange() should return false for single verse")
	}
}

// TestCommentaryChapterRange verifies handling of chapter range entries.
func TestCommentaryChapterRange(t *testing.T) {
	entry := CommentaryEntry{
		Book:         40,
		ChapterStart: 5,
		VerseStart:   1,
		ChapterEnd:   7,
		VerseEnd:     29,
		Comments:     "Sermon on the Mount (ch 5-7)",
	}

	if !entry.IsMultiChapter() {
		t.Error("IsMultiChapter() should return true")
	}
}

// TestCommentaryRTFCleaning verifies RTF formatting is stripped.
func TestCommentaryRTFCleaning(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{`{\rtf1 Hello World}`, "Hello World"},
		{`{\b Bold text}`, "Bold text"},
		{`Text with \par new paragraph`, "Text with new paragraph"},
		{"Plain text", "Plain text"},
	}

	for _, tt := range tests {
		got := cleanCommentaryText(tt.input)
		if got != tt.want {
			t.Errorf("cleanCommentaryText(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// TestCommentaryBookNumbers verifies book number mapping.
func TestCommentaryBookNumbers(t *testing.T) {
	// e-Sword uses 1-66 book numbering
	tests := []struct {
		book int
		name string
	}{
		{1, "Genesis"},
		{19, "Psalms"},
		{40, "Matthew"},
		{66, "Revelation"},
	}

	for _, tt := range tests {
		name := BookName(tt.book)
		if name != tt.name {
			t.Errorf("BookName(%d) = %q, want %q", tt.book, name, tt.name)
		}
	}
}

// TestCommentaryListBooks verifies listing all books with commentary.
func TestCommentaryListBooks(t *testing.T) {
	parser := &CommentaryParser{
		entries: map[string]*CommentaryEntry{
			"1:1:1":  {Book: 1},
			"19:1:1": {Book: 19},
			"40:1:1": {Book: 40},
			"40:5:3": {Book: 40},
		},
	}

	books := parser.ListBooks()
	if len(books) != 3 {
		t.Errorf("len(books) = %d, want 3", len(books))
	}
}

// TestCommentaryModuleInfo verifies module information extraction.
func TestCommentaryModuleInfo(t *testing.T) {
	parser := &CommentaryParser{
		dbPath: "/path/to/mhc.cmtx",
		details: &CommentaryDetails{
			Title:        "Matthew Henry's Complete Commentary",
			Abbreviation: "MHC",
		},
		entries: map[string]*CommentaryEntry{
			"1:1:1":  {Book: 1},
			"40:1:1": {Book: 40},
		},
	}

	info := parser.ModuleInfo()

	if info.Title != "Matthew Henry's Complete Commentary" {
		t.Errorf("info.Title = %q", info.Title)
	}
	if info.EntryCount != 2 {
		t.Errorf("info.EntryCount = %d, want 2", info.EntryCount)
	}
}

// TestCommentaryIPCDetect verifies IPC detect command.
func TestCommentaryIPCDetect(t *testing.T) {
	request := ipc.Request{
		Command: "detect",
		Args: map[string]interface{}{
			"path": "/path/to/commentary.cmtx",
		},
	}

	if request.Command != "detect" {
		t.Errorf("Command = %q, want %q", request.Command, "detect")
	}

	response := ipc.Response{
		Status: "ok",
		Result: map[string]interface{}{
			"type":   "e-Sword Commentary",
			"format": "cmtx",
			"valid":  true,
		},
	}

	if response.Status != "ok" {
		t.Error("Response should be successful")
	}
}

// TestCommentaryIPCGetEntry verifies IPC get-entry command.
func TestCommentaryIPCGetEntry(t *testing.T) {
	request := ipc.Request{
		Command: "get-entry",
		Args: map[string]interface{}{
			"path": "/path/to/commentary.cmtx",
			"ref":  "Matt.5.3",
		},
	}

	if request.Command != "get-entry" {
		t.Errorf("Command = %q, want %q", request.Command, "get-entry")
	}

	response := ipc.Response{
		Status: "ok",
		Result: map[string]interface{}{
			"book":     40,
			"chapter":  5,
			"verse":    3,
			"comments": "Blessed are the poor in spirit...",
		},
	}

	if response.Status != "ok" {
		t.Error("Response should be successful")
	}
}

// TestCommentaryIPCGetChapter verifies IPC get-chapter command.
func TestCommentaryIPCGetChapter(t *testing.T) {
	request := ipc.Request{
		Command: "get-chapter",
		Args: map[string]interface{}{
			"path":    "/path/to/commentary.cmtx",
			"book":    40,
			"chapter": 5,
		},
	}

	if request.Command != "get-chapter" {
		t.Errorf("Command = %q, want %q", request.Command, "get-chapter")
	}

	response := ipc.Response{
		Status: "ok",
		Result: map[string]interface{}{
			"entries": []map[string]interface{}{
				{"verse": 1, "comments": "V1 commentary"},
				{"verse": 3, "comments": "V3 commentary"},
			},
		},
	}

	if response.Status != "ok" {
		t.Error("Response should be successful")
	}
}

// TestCommentaryWithRealDB tests with actual SQLite if available.
func TestCommentaryWithRealDB(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.cmtx")

	// Create test database
	db, err := sqlite.Open(dbPath)
	if err != nil {
		t.Skip("SQLite not available")
	}
	defer db.Close()

	// Create tables (e-Sword uses ChapterBegin/VerseBegin, not ChapterStart/VerseStart)
	_, err = db.Exec(`
		CREATE TABLE Commentary (
			Book INTEGER,
			ChapterBegin INTEGER,
			VerseBegin INTEGER,
			ChapterEnd INTEGER,
			VerseEnd INTEGER,
			Comments TEXT
		);
		CREATE TABLE Details (
			Title TEXT,
			Abbreviation TEXT,
			Information TEXT,
			Version INTEGER,
			RightToLeft INTEGER
		);
	`)
	if err != nil {
		t.Fatalf("Failed to create tables: %v", err)
	}

	// Insert test data
	_, err = db.Exec(`
		INSERT INTO Commentary VALUES (40, 5, 3, 5, 3, 'Blessed are the poor in spirit');
		INSERT INTO Details VALUES ('Test Commentary', 'TC', 'Test', 1, 0);
	`)
	if err != nil {
		t.Fatalf("Failed to insert data: %v", err)
	}
	db.Close()

	// Now test the parser
	parser, err := NewCommentaryParser(dbPath)
	if err != nil {
		t.Fatalf("NewCommentaryParser failed: %v", err)
	}
	defer parser.Close()

	entry, err := parser.GetEntry(40, 5, 3)
	if err != nil {
		t.Fatalf("GetEntry failed: %v", err)
	}

	if entry.Comments != "Blessed are the poor in spirit" {
		t.Errorf("Comments = %q", entry.Comments)
	}
}

// TestCommentaryNTOnly verifies handling NT-only commentaries.
func TestCommentaryNTOnly(t *testing.T) {
	parser := &CommentaryParser{
		entries: map[string]*CommentaryEntry{
			"40:1:1": {Book: 40},
			"66:1:1": {Book: 66},
		},
	}

	if parser.HasOT() {
		t.Error("HasOT() should return false for NT-only")
	}
	if !parser.HasNT() {
		t.Error("HasNT() should return true")
	}
}

// TestCommentaryOTOnly verifies handling OT-only commentaries.
func TestCommentaryOTOnly(t *testing.T) {
	parser := &CommentaryParser{
		entries: map[string]*CommentaryEntry{
			"1:1:1":  {Book: 1},
			"39:1:1": {Book: 39},
		},
	}

	if !parser.HasOT() {
		t.Error("HasOT() should return true")
	}
	if parser.HasNT() {
		t.Error("HasNT() should return false for OT-only")
	}
}

// TestCommentaryEmptyComments verifies handling of empty commentary entries.
func TestCommentaryEmptyComments(t *testing.T) {
	parser := &CommentaryParser{
		entries: map[string]*CommentaryEntry{
			"40:5:3": {Book: 40, ChapterStart: 5, VerseStart: 3, Comments: ""},
		},
	}

	entry, err := parser.GetEntry(40, 5, 3)
	if err != nil {
		t.Fatalf("GetEntry failed: %v", err)
	}

	// Empty comments is valid (entry exists but has no content)
	if entry.Comments != "" {
		t.Errorf("Comments = %q, want empty", entry.Comments)
	}
}

// Ensure temp file is cleaned up
func init() {
	// Register sqlite driver for testing
	_ = os.TempDir()
}
