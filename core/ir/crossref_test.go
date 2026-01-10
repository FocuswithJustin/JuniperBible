package ir

import (
	"encoding/json"
	"testing"
)

// Phase 16.2: Test CrossRefType constants
func TestCrossRefTypeConstants(t *testing.T) {
	tests := []struct {
		crt  CrossRefType
		want string
	}{
		{CrossRefQuotation, "quotation"},
		{CrossRefAllusion, "allusion"},
		{CrossRefParallel, "parallel"},
		{CrossRefProphecy, "prophecy"},
		{CrossRefTypology, "typology"},
		{CrossRefGeneral, "general"},
	}

	for _, tt := range tests {
		if string(tt.crt) != tt.want {
			t.Errorf("CrossRefType = %q, want %q", tt.crt, tt.want)
		}
	}
}

// Test CrossReference JSON serialization
func TestCrossReferenceJSON(t *testing.T) {
	cr := &CrossReference{
		ID:         "cr1",
		SourceRef:  &Ref{Book: "Matt", Chapter: 5, Verse: 21, OSISID: "Matt.5.21"},
		TargetRef:  &Ref{Book: "Exod", Chapter: 20, Verse: 13, OSISID: "Exod.20.13"},
		Type:       CrossRefQuotation,
		Label:      "cf.",
		Notes:      "Jesus quoting the law",
		Confidence: 1.0,
		Source:     "TSK",
	}

	// Marshal to JSON
	data, err := json.Marshal(cr)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	// Unmarshal back
	var decoded CrossReference
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	// Verify fields
	if decoded.ID != cr.ID {
		t.Errorf("ID = %q, want %q", decoded.ID, cr.ID)
	}
	if decoded.SourceRef.OSISID != cr.SourceRef.OSISID {
		t.Errorf("SourceRef.OSISID = %q, want %q", decoded.SourceRef.OSISID, cr.SourceRef.OSISID)
	}
	if decoded.TargetRef.OSISID != cr.TargetRef.OSISID {
		t.Errorf("TargetRef.OSISID = %q, want %q", decoded.TargetRef.OSISID, cr.TargetRef.OSISID)
	}
	if decoded.Type != cr.Type {
		t.Errorf("Type = %q, want %q", decoded.Type, cr.Type)
	}
	if decoded.Confidence != cr.Confidence {
		t.Errorf("Confidence = %f, want %f", decoded.Confidence, cr.Confidence)
	}
	if decoded.Source != cr.Source {
		t.Errorf("Source = %q, want %q", decoded.Source, cr.Source)
	}
}

// Test CrossRefIndex operations
func TestCrossRefIndex(t *testing.T) {
	index := NewCrossRefIndex()

	// Add some cross-references
	cr1 := &CrossReference{
		ID:        "cr1",
		SourceRef: &Ref{Book: "Matt", Chapter: 5, Verse: 21, OSISID: "Matt.5.21"},
		TargetRef: &Ref{Book: "Exod", Chapter: 20, Verse: 13, OSISID: "Exod.20.13"},
		Type:      CrossRefQuotation,
	}
	cr2 := &CrossReference{
		ID:        "cr2",
		SourceRef: &Ref{Book: "Matt", Chapter: 5, Verse: 27, OSISID: "Matt.5.27"},
		TargetRef: &Ref{Book: "Exod", Chapter: 20, Verse: 14, OSISID: "Exod.20.14"},
		Type:      CrossRefQuotation,
	}
	cr3 := &CrossReference{
		ID:        "cr3",
		SourceRef: &Ref{Book: "Rom", Chapter: 13, Verse: 9, OSISID: "Rom.13.9"},
		TargetRef: &Ref{Book: "Exod", Chapter: 20, Verse: 13, OSISID: "Exod.20.13"},
		Type:      CrossRefQuotation,
	}

	index.Add(cr1)
	index.Add(cr2)
	index.Add(cr3)

	// Test GetBySource
	refs := index.GetBySource("Matt.5.21")
	if len(refs) != 1 {
		t.Errorf("GetBySource returned %d refs, want 1", len(refs))
	}
	if refs[0].ID != "cr1" {
		t.Errorf("GetBySource[0].ID = %q, want %q", refs[0].ID, "cr1")
	}

	// Test GetByTarget
	refs = index.GetByTarget("Exod.20.13")
	if len(refs) != 2 {
		t.Errorf("GetByTarget returned %d refs, want 2", len(refs))
	}

	// Test All
	if len(index.All) != 3 {
		t.Errorf("len(All) = %d, want 3", len(index.All))
	}
}

// Test ParseCrossRefString
func TestParseCrossRefString(t *testing.T) {
	tests := []struct {
		input   string
		wantLen int
		wantRef string // first ref OSISID
	}{
		{"Gen 1:1", 1, "Gen.1.1"},
		{"Gen 1:1-3", 1, "Gen.1.1-3"},     // Range is correctly preserved
		{"Matt 5:3-12", 1, "Matt.5.3-12"}, // Range is correctly preserved
		{"Gen 1:1; Exod 2:3", 2, "Gen.1.1"},
		{"cf. Matt 5:3-12", 1, "Matt.5.3-12"}, // Range is correctly preserved
	}

	for _, tt := range tests {
		refs, err := ParseCrossRefString(tt.input)
		if err != nil {
			t.Errorf("ParseCrossRefString(%q) error: %v", tt.input, err)
			continue
		}
		if len(refs) != tt.wantLen {
			t.Errorf("ParseCrossRefString(%q) returned %d refs, want %d", tt.input, len(refs), tt.wantLen)
			continue
		}
		if refs[0].OSISID != tt.wantRef {
			t.Errorf("ParseCrossRefString(%q)[0].OSISID = %q, want %q", tt.input, refs[0].OSISID, tt.wantRef)
		}
	}
}

// Test AddCrossReference to Corpus
func TestCorpusAddCrossReference(t *testing.T) {
	corpus := &Corpus{
		ID: "test-corpus",
	}

	cr := &CrossReference{
		ID:        "cr1",
		SourceRef: &Ref{Book: "Matt", Chapter: 5, Verse: 21, OSISID: "Matt.5.21"},
		TargetRef: &Ref{Book: "Exod", Chapter: 20, Verse: 13, OSISID: "Exod.20.13"},
		Type:      CrossRefQuotation,
	}

	corpus.AddCrossReference(cr)

	if len(corpus.CrossReferences) != 1 {
		t.Errorf("len(CrossReferences) = %d, want 1", len(corpus.CrossReferences))
	}
}

// Test BuildCrossRefIndex
func TestCorpusBuildCrossRefIndex(t *testing.T) {
	corpus := &Corpus{
		ID: "test-corpus",
		CrossReferences: []*CrossReference{
			{
				ID:        "cr1",
				SourceRef: &Ref{Book: "Matt", Chapter: 5, Verse: 21, OSISID: "Matt.5.21"},
				TargetRef: &Ref{Book: "Exod", Chapter: 20, Verse: 13, OSISID: "Exod.20.13"},
			},
		},
	}

	index := corpus.BuildCrossRefIndex()

	if index == nil {
		t.Fatal("BuildCrossRefIndex returned nil")
	}
	if len(index.All) != 1 {
		t.Errorf("len(index.All) = %d, want 1", len(index.All))
	}
}

// TestParseCrossRefStringEmptyParts tests parsing with empty parts (e.g., "Gen 1:1; ; Gen 1:2")
func TestParseCrossRefStringEmptyParts(t *testing.T) {
	// Input with empty parts between semicolons
	refs, err := ParseCrossRefString("Gen 1:1; ; Gen 1:2")
	if err != nil {
		t.Fatalf("ParseCrossRefString failed: %v", err)
	}

	// Should skip empty parts and return 2 refs
	if len(refs) != 2 {
		t.Errorf("got %d refs, want 2", len(refs))
	}
}

// TestParseCrossRefStringOnlyWhitespace tests parsing with only whitespace parts
func TestParseCrossRefStringOnlyWhitespace(t *testing.T) {
	refs, err := ParseCrossRefString("Gen 1:1;   ;  \t  ; Gen 1:2")
	if err != nil {
		t.Fatalf("ParseCrossRefString failed: %v", err)
	}

	if len(refs) != 2 {
		t.Errorf("got %d refs, want 2", len(refs))
	}
}

// TestParseCrossRefStringInvalidRef tests parsing with invalid reference format
func TestParseCrossRefStringInvalidRef(t *testing.T) {
	// An invalid reference should return nil refs (parseSimpleRef returns nil)
	refs, err := ParseCrossRefString("not a valid ref")
	if err != nil {
		t.Fatalf("ParseCrossRefString failed: %v", err)
	}
	// parseSimpleRef returns nil for invalid refs, which are filtered out
	if len(refs) != 0 {
		t.Errorf("got %d refs, want 0 for invalid ref", len(refs))
	}
}

// TestNormalizeBookNameUnknown tests normalizing an unknown book name
func TestNormalizeBookNameUnknown(t *testing.T) {
	// Parse a valid ref with an unknown book name - it should pass through unchanged
	refs, err := ParseCrossRefString("UnknownBook 1:1")
	if err != nil {
		t.Fatalf("ParseCrossRefString failed: %v", err)
	}
	if len(refs) != 1 {
		t.Fatalf("got %d refs, want 1", len(refs))
	}
	// Unknown book name should be passed through as-is
	if refs[0].Book != "UnknownBook" {
		t.Errorf("Book = %q, want %q", refs[0].Book, "UnknownBook")
	}
}
