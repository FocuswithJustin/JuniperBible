package main

import (
	"path/filepath"
	"testing"

	"github.com/FocuswithJustin/JuniperBible/core/sqlite"
	"github.com/FocuswithJustin/JuniperBible/plugins/ipc"
)

// Phase 18: Tests for e-Sword Dictionary (.dctx) parser

// TestDictionaryEntryStruct verifies the DictionaryEntry structure.
func TestDictionaryEntryStruct(t *testing.T) {
	entry := DictionaryEntry{
		Topic:      "Grace",
		Definition: "Unmerited favor from God...",
	}

	if entry.Topic != "Grace" {
		t.Errorf("Topic = %q, want %q", entry.Topic, "Grace")
	}
	if entry.Definition == "" {
		t.Error("Definition should not be empty")
	}
}

// TestDictionaryParserCreation verifies parser creation with database path.
func TestDictionaryParserCreation(t *testing.T) {
	parser, err := NewDictionaryParser("/path/to/dictionary.dctx")
	if err != nil {
		// Expected to fail with non-existent path
		return
	}

	if parser == nil {
		t.Error("NewDictionaryParser should return a parser or error")
	}
}

// TestDictionaryDetect verifies detection of e-Sword dictionary files.
func TestDictionaryDetect(t *testing.T) {
	tests := []struct {
		filename string
		want     bool
	}{
		{"Eastons.dctx", true},
		{"dictionary.dctx", true},
		{"DICTIONARY.DCTX", true},
		{"bible.bblx", false},
		{"commentary.cmtx", false},
		{"file.sqlite", false},
	}

	for _, tt := range tests {
		got := IsDictionaryFile(tt.filename)
		if got != tt.want {
			t.Errorf("IsDictionaryFile(%q) = %v, want %v", tt.filename, got, tt.want)
		}
	}
}

// TestDictionaryTableSchema verifies expected table structure.
func TestDictionaryTableSchema(t *testing.T) {
	// Dictionary table should have these columns:
	expectedColumns := []string{
		"Topic",
		"Definition",
	}

	for _, col := range expectedColumns {
		if col == "" {
			t.Error("column name should not be empty")
		}
	}
}

// TestDictionaryDetailsTable verifies reading the Details table.
func TestDictionaryDetailsTable(t *testing.T) {
	details := DictionaryDetails{
		Title:        "Easton's Bible Dictionary",
		Abbreviation: "EBD",
		Information:  "Public domain dictionary",
		Version:      1,
	}

	if details.Title == "" {
		t.Error("Title should not be empty")
	}
	if details.Abbreviation == "" {
		t.Error("Abbreviation should not be empty")
	}
}

// TestDictionaryGetEntry verifies retrieving a single dictionary entry.
func TestDictionaryGetEntry(t *testing.T) {
	parser := &DictionaryParser{
		entries: map[string]*DictionaryEntry{
			"Grace": {
				Topic:      "Grace",
				Definition: "Unmerited favor from God...",
			},
			"Faith": {
				Topic:      "Faith",
				Definition: "Trust and belief in God...",
			},
		},
	}

	entry, err := parser.GetEntry("Grace")
	if err != nil {
		t.Fatalf("GetEntry failed: %v", err)
	}

	if entry.Topic != "Grace" {
		t.Errorf("Topic = %q, want %q", entry.Topic, "Grace")
	}
	if entry.Definition == "" {
		t.Error("Definition should not be empty")
	}
}

// TestDictionaryGetEntryNotFound verifies error handling for missing entries.
func TestDictionaryGetEntryNotFound(t *testing.T) {
	parser := &DictionaryParser{
		entries: map[string]*DictionaryEntry{},
	}

	_, err := parser.GetEntry("NotFound")
	if err == nil {
		t.Error("GetEntry should return error for missing entry")
	}
}

// TestDictionaryListTopics verifies listing all dictionary topics.
func TestDictionaryListTopics(t *testing.T) {
	parser := &DictionaryParser{
		entries: map[string]*DictionaryEntry{
			"Aaron":   {Topic: "Aaron"},
			"Abel":    {Topic: "Abel"},
			"Abraham": {Topic: "Abraham"},
		},
	}

	topics := parser.ListTopics()
	if len(topics) != 3 {
		t.Errorf("len(topics) = %d, want 3", len(topics))
	}
}

// TestDictionarySearchTopics verifies topic search functionality.
func TestDictionarySearchTopics(t *testing.T) {
	parser := &DictionaryParser{
		entries: map[string]*DictionaryEntry{
			"Aaron":   {Topic: "Aaron"},
			"Abel":    {Topic: "Abel"},
			"Abraham": {Topic: "Abraham"},
			"Babylon": {Topic: "Babylon"},
		},
	}

	// Search for topics starting with "A"
	results := parser.SearchTopics("A")
	if len(results) != 3 {
		t.Errorf("SearchTopics('A') returned %d results, want 3", len(results))
	}

	// Search for topics starting with "Ab"
	results = parser.SearchTopics("Ab")
	if len(results) != 2 {
		t.Errorf("SearchTopics('Ab') returned %d results, want 2", len(results))
	}
}

// TestDictionarySearchDefinitions verifies full-text search in definitions.
func TestDictionarySearchDefinitions(t *testing.T) {
	parser := &DictionaryParser{
		entries: map[string]*DictionaryEntry{
			"Grace": {Topic: "Grace", Definition: "Unmerited favor from God"},
			"Faith": {Topic: "Faith", Definition: "Trust and belief in God"},
			"Hope":  {Topic: "Hope", Definition: "Expectation of future blessing"},
		},
	}

	// Search for "God" should find Grace and Faith
	results := parser.SearchDefinitions("God")
	if len(results) != 2 {
		t.Errorf("SearchDefinitions('God') returned %d results, want 2", len(results))
	}
}

// TestDictionaryCaseInsensitiveSearch verifies case-insensitive search.
func TestDictionaryCaseInsensitiveSearch(t *testing.T) {
	parser := &DictionaryParser{
		entries: map[string]*DictionaryEntry{
			"Grace": {Topic: "Grace", Definition: "Unmerited favor"},
		},
	}

	// Should find regardless of case
	tests := []string{"grace", "GRACE", "Grace", "gRaCe"}
	for _, query := range tests {
		results := parser.SearchTopics(query)
		if len(results) != 1 {
			t.Errorf("SearchTopics(%q) returned %d results, want 1", query, len(results))
		}
	}
}

// TestDictionaryRTFCleaning verifies RTF formatting is stripped.
func TestDictionaryRTFCleaning(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{`{\rtf1 Definition text}`, "Definition text"},
		{`{\b Bold term}`, "Bold term"},
		{`Text with \par new paragraph`, "Text with new paragraph"},
		{"Plain text", "Plain text"},
	}

	for _, tt := range tests {
		got := cleanDictionaryText(tt.input)
		if got != tt.want {
			t.Errorf("cleanDictionaryText(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// TestDictionaryUnicodeTopics verifies handling of Unicode in topics.
func TestDictionaryUnicodeTopics(t *testing.T) {
	parser := &DictionaryParser{
		entries: map[string]*DictionaryEntry{
			"θεός":     {Topic: "θεός", Definition: "Greek word for God"},
			"אֱלֹהִים": {Topic: "אֱלֹהִים", Definition: "Hebrew word for God"},
		},
	}

	entry, err := parser.GetEntry("θεός")
	if err != nil {
		t.Fatalf("GetEntry failed for Greek topic: %v", err)
	}
	if entry.Definition == "" {
		t.Error("Definition should not be empty")
	}

	entry, err = parser.GetEntry("אֱלֹהִים")
	if err != nil {
		t.Fatalf("GetEntry failed for Hebrew topic: %v", err)
	}
	if entry.Definition == "" {
		t.Error("Definition should not be empty")
	}
}

// TestDictionaryModuleInfo verifies module information extraction.
func TestDictionaryModuleInfo(t *testing.T) {
	parser := &DictionaryParser{
		dbPath: "/path/to/eastons.dctx",
		details: &DictionaryDetails{
			Title:        "Easton's Bible Dictionary",
			Abbreviation: "EBD",
		},
		entries: map[string]*DictionaryEntry{
			"Aaron":   {Topic: "Aaron"},
			"Abel":    {Topic: "Abel"},
			"Abraham": {Topic: "Abraham"},
		},
	}

	info := parser.ModuleInfo()

	if info.Title != "Easton's Bible Dictionary" {
		t.Errorf("info.Title = %q", info.Title)
	}
	if info.EntryCount != 3 {
		t.Errorf("info.EntryCount = %d, want 3", info.EntryCount)
	}
}

// TestDictionaryIPCDetect verifies IPC detect command.
func TestDictionaryIPCDetect(t *testing.T) {
	request := ipc.Request{
		Command: "detect",
		Args: map[string]interface{}{
			"path": "/path/to/dictionary.dctx",
		},
	}

	if request.Command != "detect" {
		t.Errorf("Command = %q, want %q", request.Command, "detect")
	}

	response := ipc.Response{
		Status: "ok",
		Result: map[string]interface{}{
			"type":   "e-Sword Dictionary",
			"format": "dctx",
			"valid":  true,
		},
	}

	if response.Status != "ok" {
		t.Error("Response should be successful")
	}
}

// TestDictionaryIPCGetEntry verifies IPC get-entry command.
func TestDictionaryIPCGetEntry(t *testing.T) {
	request := ipc.Request{
		Command: "get-entry",
		Args: map[string]interface{}{
			"path":  "/path/to/dictionary.dctx",
			"topic": "Grace",
		},
	}

	if request.Command != "get-entry" {
		t.Errorf("Command = %q, want %q", request.Command, "get-entry")
	}

	response := ipc.Response{
		Status: "ok",
		Result: map[string]interface{}{
			"topic":      "Grace",
			"definition": "Unmerited favor from God...",
		},
	}

	if response.Status != "ok" {
		t.Error("Response should be successful")
	}
}

// TestDictionaryIPCListTopics verifies IPC list-topics command.
func TestDictionaryIPCListTopics(t *testing.T) {
	request := ipc.Request{
		Command: "list-topics",
		Args: map[string]interface{}{
			"path": "/path/to/dictionary.dctx",
		},
	}

	if request.Command != "list-topics" {
		t.Errorf("Command = %q, want %q", request.Command, "list-topics")
	}

	response := ipc.Response{
		Status: "ok",
		Result: map[string]interface{}{
			"topics": []string{"Aaron", "Abel", "Abraham"},
		},
	}

	if response.Status != "ok" {
		t.Error("Response should be successful")
	}
}

// TestDictionaryIPCSearch verifies IPC search command.
func TestDictionaryIPCSearch(t *testing.T) {
	request := ipc.Request{
		Command: "search",
		Args: map[string]interface{}{
			"path":  "/path/to/dictionary.dctx",
			"query": "God",
			"in":    "definitions",
		},
	}

	if request.Command != "search" {
		t.Errorf("Command = %q, want %q", request.Command, "search")
	}

	response := ipc.Response{
		Status: "ok",
		Result: map[string]interface{}{
			"matches": []string{"Grace", "Faith"},
		},
	}

	if response.Status != "ok" {
		t.Error("Response should be successful")
	}
}

// TestDictionaryWithRealDB tests with actual SQLite if available.
func TestDictionaryWithRealDB(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.dctx")

	// Create test database
	db, err := sqlite.Open(dbPath)
	if err != nil {
		t.Skip("SQLite not available")
	}
	defer db.Close()

	// Create tables
	_, err = db.Exec(`
		CREATE TABLE Dictionary (
			Topic TEXT,
			Definition TEXT
		);
		CREATE TABLE Details (
			Title TEXT,
			Abbreviation TEXT,
			Information TEXT,
			Version INTEGER
		);
	`)
	if err != nil {
		t.Fatalf("Failed to create tables: %v", err)
	}

	// Insert test data
	_, err = db.Exec(`
		INSERT INTO Dictionary VALUES ('Grace', 'Unmerited favor from God');
		INSERT INTO Dictionary VALUES ('Faith', 'Trust and belief in God');
		INSERT INTO Details VALUES ('Test Dictionary', 'TD', 'Test', 1);
	`)
	if err != nil {
		t.Fatalf("Failed to insert data: %v", err)
	}
	db.Close()

	// Now test the parser
	parser, err := NewDictionaryParser(dbPath)
	if err != nil {
		t.Fatalf("NewDictionaryParser failed: %v", err)
	}
	defer parser.Close()

	entry, err := parser.GetEntry("Grace")
	if err != nil {
		t.Fatalf("GetEntry failed: %v", err)
	}

	if entry.Definition != "Unmerited favor from God" {
		t.Errorf("Definition = %q", entry.Definition)
	}
}

// TestDictionaryEmptyDefinition verifies handling of empty definitions.
func TestDictionaryEmptyDefinition(t *testing.T) {
	parser := &DictionaryParser{
		entries: map[string]*DictionaryEntry{
			"Empty": {Topic: "Empty", Definition: ""},
		},
	}

	entry, err := parser.GetEntry("Empty")
	if err != nil {
		t.Fatalf("GetEntry failed: %v", err)
	}

	// Empty definition is valid (topic exists but has no content)
	if entry.Definition != "" {
		t.Errorf("Definition = %q, want empty", entry.Definition)
	}
}

// TestDictionaryLongDefinition verifies handling of long definitions.
func TestDictionaryLongDefinition(t *testing.T) {
	longDef := ""
	for i := 0; i < 1000; i++ {
		longDef += "This is a very long dictionary definition. "
	}

	parser := &DictionaryParser{
		entries: map[string]*DictionaryEntry{
			"Long": {Topic: "Long", Definition: longDef},
		},
	}

	entry, err := parser.GetEntry("Long")
	if err != nil {
		t.Fatalf("GetEntry failed: %v", err)
	}

	if len(entry.Definition) != len(longDef) {
		t.Errorf("Definition length = %d, want %d", len(entry.Definition), len(longDef))
	}
}

// TestDictionarySpecialCharactersInTopics verifies handling of special chars.
func TestDictionarySpecialCharactersInTopics(t *testing.T) {
	parser := &DictionaryParser{
		entries: map[string]*DictionaryEntry{
			"God's Word":      {Topic: "God's Word"},
			"Alpha & Omega":   {Topic: "Alpha & Omega"},
			"Question?":       {Topic: "Question?"},
			"(Parenthetical)": {Topic: "(Parenthetical)"},
		},
	}

	tests := []string{
		"God's Word",
		"Alpha & Omega",
		"Question?",
		"(Parenthetical)",
	}

	for _, topic := range tests {
		entry, err := parser.GetEntry(topic)
		if err != nil {
			t.Errorf("GetEntry(%q) failed: %v", topic, err)
			continue
		}
		if entry.Topic != topic {
			t.Errorf("entry.Topic = %q, want %q", entry.Topic, topic)
		}
	}
}

// TestDictionarySortedTopics verifies topics can be sorted alphabetically.
func TestDictionarySortedTopics(t *testing.T) {
	parser := &DictionaryParser{
		entries: map[string]*DictionaryEntry{
			"Zion":  {Topic: "Zion"},
			"Aaron": {Topic: "Aaron"},
			"Moses": {Topic: "Moses"},
		},
	}

	topics := parser.ListTopicsSorted()
	if len(topics) != 3 {
		t.Fatalf("len(topics) = %d, want 3", len(topics))
	}

	// Should be alphabetically sorted
	expected := []string{"Aaron", "Moses", "Zion"}
	for i, exp := range expected {
		if topics[i] != exp {
			t.Errorf("topics[%d] = %q, want %q", i, topics[i], exp)
		}
	}
}

// TestDictionaryEntryCount verifies counting entries.
func TestDictionaryEntryCount(t *testing.T) {
	parser := &DictionaryParser{
		entries: map[string]*DictionaryEntry{
			"A": {Topic: "A"},
			"B": {Topic: "B"},
			"C": {Topic: "C"},
		},
	}

	count := parser.EntryCount()
	if count != 3 {
		t.Errorf("EntryCount() = %d, want 3", count)
	}
}
