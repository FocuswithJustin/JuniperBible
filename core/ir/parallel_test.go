package ir

import (
	"encoding/json"
	"testing"
)

// Phase 16.3: Test AlignmentLevel constants
func TestAlignmentLevelConstants(t *testing.T) {
	tests := []struct {
		al   AlignmentLevel
		want string
	}{
		{AlignBook, "book"},
		{AlignChapter, "chapter"},
		{AlignVerse, "verse"},
		{AlignToken, "token"},
	}

	for _, tt := range tests {
		if string(tt.al) != tt.want {
			t.Errorf("AlignmentLevel = %q, want %q", tt.al, tt.want)
		}
	}
}

// Test ParallelCorpus JSON serialization
func TestParallelCorpusJSON(t *testing.T) {
	pc := &ParallelCorpus{
		ID:      "kjv-niv-parallel",
		Version: "1.0.0",
		BaseCorpus: &CorpusRef{
			ID:       "KJV",
			Language: "en",
		},
		Corpora: []*CorpusRef{
			{ID: "KJV", Language: "en"},
			{ID: "NIV", Language: "en"},
		},
		DefaultAlignment: AlignVerse,
	}

	// Marshal to JSON
	data, err := json.Marshal(pc)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	// Unmarshal back
	var decoded ParallelCorpus
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	// Verify fields
	if decoded.ID != pc.ID {
		t.Errorf("ID = %q, want %q", decoded.ID, pc.ID)
	}
	if decoded.Version != pc.Version {
		t.Errorf("Version = %q, want %q", decoded.Version, pc.Version)
	}
	if decoded.BaseCorpus.ID != pc.BaseCorpus.ID {
		t.Errorf("BaseCorpus.ID = %q, want %q", decoded.BaseCorpus.ID, pc.BaseCorpus.ID)
	}
	if len(decoded.Corpora) != 2 {
		t.Errorf("len(Corpora) = %d, want 2", len(decoded.Corpora))
	}
	if decoded.DefaultAlignment != pc.DefaultAlignment {
		t.Errorf("DefaultAlignment = %q, want %q", decoded.DefaultAlignment, pc.DefaultAlignment)
	}
}

// Test AlignedUnit JSON serialization
func TestAlignedUnitJSON(t *testing.T) {
	au := &AlignedUnit{
		ID:  "au1",
		Ref: &Ref{Book: "Gen", Chapter: 1, Verse: 1, OSISID: "Gen.1.1"},
		Texts: map[string]string{
			"KJV": "In the beginning God created the heaven and the earth.",
			"NIV": "In the beginning God created the heavens and the earth.",
		},
		Level: AlignVerse,
	}

	// Marshal to JSON
	data, err := json.Marshal(au)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	// Unmarshal back
	var decoded AlignedUnit
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	// Verify fields
	if decoded.ID != au.ID {
		t.Errorf("ID = %q, want %q", decoded.ID, au.ID)
	}
	if decoded.Ref.OSISID != au.Ref.OSISID {
		t.Errorf("Ref.OSISID = %q, want %q", decoded.Ref.OSISID, au.Ref.OSISID)
	}
	if len(decoded.Texts) != 2 {
		t.Errorf("len(Texts) = %d, want 2", len(decoded.Texts))
	}
	if decoded.Texts["KJV"] != au.Texts["KJV"] {
		t.Errorf("Texts[KJV] = %q, want %q", decoded.Texts["KJV"], au.Texts["KJV"])
	}
}

// Test AlignByVerse
func TestAlignByVerse(t *testing.T) {
	// Create two corpora
	kjv := &Corpus{
		ID:       "KJV",
		Language: "en",
		Documents: []*Document{
			{
				ID:    "Gen",
				Title: "Genesis",
				ContentBlocks: []*ContentBlock{
					{ID: "cb1", Sequence: 0, Text: "In the beginning God created the heaven and the earth."},
				},
			},
		},
	}
	niv := &Corpus{
		ID:       "NIV",
		Language: "en",
		Documents: []*Document{
			{
				ID:    "Gen",
				Title: "Genesis",
				ContentBlocks: []*ContentBlock{
					{ID: "cb1", Sequence: 0, Text: "In the beginning God created the heavens and the earth."},
				},
			},
		},
	}

	// Align the corpora
	pc, err := AlignByVerse([]*Corpus{kjv, niv})
	if err != nil {
		t.Fatalf("AlignByVerse failed: %v", err)
	}

	// Verify parallel corpus
	if pc == nil {
		t.Fatal("AlignByVerse returned nil")
	}
	if len(pc.Corpora) != 2 {
		t.Errorf("len(Corpora) = %d, want 2", len(pc.Corpora))
	}
	if pc.DefaultAlignment != AlignVerse {
		t.Errorf("DefaultAlignment = %q, want %q", pc.DefaultAlignment, AlignVerse)
	}
}

// Test GetAlignedVerses
func TestGetAlignedVerses(t *testing.T) {
	pc := &ParallelCorpus{
		ID:               "test-parallel",
		DefaultAlignment: AlignVerse,
		Alignments: []*Alignment{
			{
				ID:    "a1",
				Level: AlignVerse,
				Units: []*AlignedUnit{
					{
						ID:  "au1",
						Ref: &Ref{Book: "Gen", Chapter: 1, Verse: 1, OSISID: "Gen.1.1"},
						Texts: map[string]string{
							"KJV": "In the beginning...",
							"NIV": "In the beginning...",
						},
					},
					{
						ID:  "au2",
						Ref: &Ref{Book: "Gen", Chapter: 1, Verse: 2, OSISID: "Gen.1.2"},
						Texts: map[string]string{
							"KJV": "And the earth was without form...",
							"NIV": "Now the earth was formless...",
						},
					},
				},
			},
		},
	}

	// Get aligned verses for Gen.1.1
	ref := &Ref{Book: "Gen", Chapter: 1, Verse: 1, OSISID: "Gen.1.1"}
	units := pc.GetAlignedVerses(ref)

	if len(units) != 1 {
		t.Fatalf("GetAlignedVerses returned %d units, want 1", len(units))
	}
	if units[0].ID != "au1" {
		t.Errorf("units[0].ID = %q, want %q", units[0].ID, "au1")
	}
}

// Test TokenAlignment
func TestTokenAlignmentJSON(t *testing.T) {
	ta := &TokenAlignment{
		ID:           "ta1",
		SourceTokens: []string{"t1", "t2"},
		TargetTokens: []string{"t3"},
		Confidence:   0.95,
		AlignType:    "one-to-many",
	}

	// Marshal to JSON
	data, err := json.Marshal(ta)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	// Unmarshal back
	var decoded TokenAlignment
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	// Verify fields
	if decoded.ID != ta.ID {
		t.Errorf("ID = %q, want %q", decoded.ID, ta.ID)
	}
	if len(decoded.SourceTokens) != 2 {
		t.Errorf("len(SourceTokens) = %d, want 2", len(decoded.SourceTokens))
	}
	if decoded.Confidence != ta.Confidence {
		t.Errorf("Confidence = %f, want %f", decoded.Confidence, ta.Confidence)
	}
}

// Test InterlinearLine
func TestInterlinearLineJSON(t *testing.T) {
	il := &InterlinearLine{
		Ref: &Ref{Book: "Gen", Chapter: 1, Verse: 1, OSISID: "Gen.1.1"},
		Layers: map[string]*InterlinearLayer{
			"hebrew": {
				CorpusID: "OSHB",
				Tokens:   []string{"בְּרֵאשִׁית", "בָּרָא", "אֱלֹהִים"},
				Label:    "Hebrew",
			},
			"english": {
				CorpusID: "KJV",
				Tokens:   []string{"In the beginning", "created", "God"},
				Label:    "English",
			},
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(il)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	// Unmarshal back
	var decoded InterlinearLine
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	// Verify fields
	if decoded.Ref.OSISID != il.Ref.OSISID {
		t.Errorf("Ref.OSISID = %q, want %q", decoded.Ref.OSISID, il.Ref.OSISID)
	}
	if len(decoded.Layers) != 2 {
		t.Errorf("len(Layers) = %d, want 2", len(decoded.Layers))
	}
	if decoded.Layers["hebrew"].CorpusID != "OSHB" {
		t.Errorf("Layers[hebrew].CorpusID = %q, want %q", decoded.Layers["hebrew"].CorpusID, "OSHB")
	}
}

// TestAlignTokens tests the AlignTokens function.
func TestAlignTokens(t *testing.T) {
	source := &ContentBlock{
		ID:   "cb1",
		Text: "In the beginning",
		Tokens: []*Token{
			{ID: "t1", Text: "In", CharStart: 0, CharEnd: 2},
			{ID: "t2", Text: "the", CharStart: 3, CharEnd: 6},
			{ID: "t3", Text: "beginning", CharStart: 7, CharEnd: 16},
		},
	}

	target := &ContentBlock{
		ID:   "cb2",
		Text: "Al principio",
		Tokens: []*Token{
			{ID: "t4", Text: "Al", CharStart: 0, CharEnd: 2},
			{ID: "t5", Text: "principio", CharStart: 3, CharEnd: 12},
		},
	}

	// Align with default options
	alignments, err := AlignTokens(source, target, nil)
	if err != nil {
		t.Fatalf("AlignTokens failed: %v", err)
	}

	// Should align min(3, 2) = 2 tokens
	if len(alignments) != 2 {
		t.Errorf("AlignTokens returned %d alignments, want 2", len(alignments))
	}

	// Check first alignment
	if len(alignments) > 0 {
		if alignments[0].SourceTokens[0] != "t1" {
			t.Errorf("alignments[0].SourceTokens[0] = %q, want t1", alignments[0].SourceTokens[0])
		}
		if alignments[0].TargetTokens[0] != "t4" {
			t.Errorf("alignments[0].TargetTokens[0] = %q, want t4", alignments[0].TargetTokens[0])
		}
	}
}

// TestAlignTokensWithOptions tests AlignTokens with custom options.
func TestAlignTokensWithOptions(t *testing.T) {
	source := &ContentBlock{
		ID:   "cb1",
		Text: "Hello world",
		Tokens: []*Token{
			{ID: "t1", Text: "Hello", CharStart: 0, CharEnd: 5},
			{ID: "t2", Text: "world", CharStart: 6, CharEnd: 11},
		},
	}

	target := &ContentBlock{
		ID:   "cb2",
		Text: "Hola mundo",
		Tokens: []*Token{
			{ID: "t3", Text: "Hola", CharStart: 0, CharEnd: 4},
			{ID: "t4", Text: "mundo", CharStart: 5, CharEnd: 10},
		},
	}

	opts := &AlignOptions{
		MinConfidence:  0.8,
		AllowUnaligned: false,
	}

	alignments, err := AlignTokens(source, target, opts)
	if err != nil {
		t.Fatalf("AlignTokens with options failed: %v", err)
	}

	// Should align 2 tokens
	if len(alignments) != 2 {
		t.Errorf("AlignTokens returned %d alignments, want 2", len(alignments))
	}
}

// TestAlignTokensEmpty tests AlignTokens with empty content blocks.
func TestAlignTokensEmpty(t *testing.T) {
	source := &ContentBlock{
		ID:     "cb1",
		Text:   "No tokens",
		Tokens: []*Token{},
	}

	target := &ContentBlock{
		ID:     "cb2",
		Text:   "Also no tokens",
		Tokens: []*Token{},
	}

	alignments, err := AlignTokens(source, target, nil)
	if err != nil {
		t.Fatalf("AlignTokens failed: %v", err)
	}

	if len(alignments) != 0 {
		t.Errorf("AlignTokens returned %d alignments for empty blocks, want 0", len(alignments))
	}
}

// TestGetAlignedVersesEmpty tests GetAlignedVerses when no match found.
func TestGetAlignedVersesEmpty(t *testing.T) {
	pc := &ParallelCorpus{
		ID:               "test-parallel",
		DefaultAlignment: AlignVerse,
		Alignments: []*Alignment{
			{
				ID:    "a1",
				Level: AlignVerse,
				Units: []*AlignedUnit{
					{
						ID:  "au1",
						Ref: &Ref{Book: "Gen", Chapter: 1, Verse: 1, OSISID: "Gen.1.1"},
						Texts: map[string]string{
							"KJV": "In the beginning...",
						},
					},
				},
			},
		},
	}

	// Get aligned verses for a ref that doesn't exist
	ref := &Ref{Book: "Exod", Chapter: 1, Verse: 1, OSISID: "Exod.1.1"}
	units := pc.GetAlignedVerses(ref)

	if len(units) != 0 {
		t.Errorf("GetAlignedVerses returned %d units for non-existent ref, want 0", len(units))
	}
}

// TestAlignByVerseEmpty tests AlignByVerse with empty corpora.
func TestAlignByVerseEmpty(t *testing.T) {
	_, err := AlignByVerse([]*Corpus{})
	if err == nil {
		t.Error("AlignByVerse should return error for empty corpora")
	}
}

// TestGetAlignedVersesNonVerseLevel tests GetAlignedVerses with non-verse alignments.
func TestGetAlignedVersesNonVerseLevel(t *testing.T) {
	pc := &ParallelCorpus{
		ID: "test",
		Alignments: []*Alignment{
			{
				ID:    "chapter-align",
				Level: AlignChapter, // Not AlignVerse, should be skipped
				Units: []*AlignedUnit{
					{
						ID:  "au1",
						Ref: &Ref{Book: "Gen", Chapter: 1, Verse: 1, OSISID: "Gen.1.1"},
					},
				},
			},
			{
				ID:    "verse-align",
				Level: AlignVerse, // This one should be checked
				Units: []*AlignedUnit{
					{
						ID:  "au2",
						Ref: &Ref{Book: "Gen", Chapter: 1, Verse: 1, OSISID: "Gen.1.1"},
					},
				},
			},
		},
	}

	ref := &Ref{Book: "Gen", Chapter: 1, Verse: 1, OSISID: "Gen.1.1"}
	units := pc.GetAlignedVerses(ref)

	// Should only return the verse-level alignment, not the chapter-level one
	if len(units) != 1 {
		t.Errorf("GetAlignedVerses returned %d units, want 1", len(units))
	}
}
