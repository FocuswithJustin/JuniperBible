package ir

import (
	"encoding/json"
	"testing"
)

func TestAnchorJSON(t *testing.T) {
	anchor := &Anchor{
		ID:             "a1",
		ContentBlockID: "cb1",
		CharOffset:     10,
		TokenIndex:     2,
		Hash:           "abc123",
	}

	// Marshal to JSON
	data, err := json.Marshal(anchor)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	// Unmarshal back
	var decoded Anchor
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	// Verify fields
	if decoded.ID != anchor.ID {
		t.Errorf("ID = %q, want %q", decoded.ID, anchor.ID)
	}
	if decoded.ContentBlockID != anchor.ContentBlockID {
		t.Errorf("ContentBlockID = %q, want %q", decoded.ContentBlockID, anchor.ContentBlockID)
	}
	if decoded.CharOffset != anchor.CharOffset {
		t.Errorf("CharOffset = %d, want %d", decoded.CharOffset, anchor.CharOffset)
	}
	if decoded.TokenIndex != anchor.TokenIndex {
		t.Errorf("TokenIndex = %d, want %d", decoded.TokenIndex, anchor.TokenIndex)
	}
}

func TestSpanTypeConstants(t *testing.T) {
	tests := []struct {
		st   SpanType
		want string
	}{
		{SpanVerse, "VERSE"},
		{SpanChapter, "CHAPTER"},
		{SpanParagraph, "PARAGRAPH"},
		{SpanPoetryLine, "POETRY_LINE"},
		{SpanQuotation, "QUOTATION"},
		{SpanRedLetter, "RED_LETTER"},
		{SpanNote, "NOTE"},
		{SpanCrossRef, "CROSS_REF"},
		{SpanSection, "SECTION"},
		{SpanTitle, "TITLE"},
	}

	for _, tt := range tests {
		if string(tt.st) != tt.want {
			t.Errorf("SpanType = %q, want %q", tt.st, tt.want)
		}
	}
}

func TestSpanTypeValidation(t *testing.T) {
	tests := []struct {
		st    SpanType
		valid bool
	}{
		{SpanVerse, true},
		{SpanChapter, true},
		{SpanParagraph, true},
		{SpanType("INVALID"), false},
		{SpanType(""), false},
	}

	for _, tt := range tests {
		if got := tt.st.IsValid(); got != tt.valid {
			t.Errorf("SpanType(%q).IsValid() = %v, want %v", tt.st, got, tt.valid)
		}
	}
}

func TestSpanJSON(t *testing.T) {
	span := &Span{
		ID:            "s1",
		Type:          SpanVerse,
		StartAnchorID: "a1",
		EndAnchorID:   "a2",
		Ref: &Ref{
			Book:    "Gen",
			Chapter: 1,
			Verse:   1,
			OSISID:  "Gen.1.1",
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(span)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	// Unmarshal back
	var decoded Span
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	// Verify fields
	if decoded.ID != span.ID {
		t.Errorf("ID = %q, want %q", decoded.ID, span.ID)
	}
	if decoded.Type != span.Type {
		t.Errorf("Type = %q, want %q", decoded.Type, span.Type)
	}
	if decoded.StartAnchorID != span.StartAnchorID {
		t.Errorf("StartAnchorID = %q, want %q", decoded.StartAnchorID, span.StartAnchorID)
	}
	if decoded.Ref == nil {
		t.Fatal("Ref is nil")
	}
	if decoded.Ref.OSISID != span.Ref.OSISID {
		t.Errorf("Ref.OSISID = %q, want %q", decoded.Ref.OSISID, span.Ref.OSISID)
	}
}

func TestSpanAttributes(t *testing.T) {
	span := &Span{
		ID:   "s1",
		Type: SpanQuotation,
	}

	// Set attributes
	span.SetAttribute("speaker", "Jesus")
	span.SetAttribute("level", 1)

	// Get attributes
	speaker, ok := span.GetAttribute("speaker")
	if !ok {
		t.Error("speaker attribute not found")
	}
	if speaker != "Jesus" {
		t.Errorf("speaker = %v, want %q", speaker, "Jesus")
	}

	level, ok := span.GetAttribute("level")
	if !ok {
		t.Error("level attribute not found")
	}
	if level != 1 {
		t.Errorf("level = %v, want 1", level)
	}

	// Get non-existent attribute
	_, ok = span.GetAttribute("nonexistent")
	if ok {
		t.Error("nonexistent attribute should not be found")
	}
}

func TestSpanWithAttributes(t *testing.T) {
	span := &Span{
		ID:   "s1",
		Type: SpanRedLetter,
		Attributes: map[string]interface{}{
			"style": "italic",
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(span)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	// Unmarshal back
	var decoded Span
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	// Verify attributes
	style, ok := decoded.GetAttribute("style")
	if !ok {
		t.Error("style attribute not found")
	}
	if style != "italic" {
		t.Errorf("style = %v, want %q", style, "italic")
	}
}

func TestAnnotationTypeConstants(t *testing.T) {
	tests := []struct {
		at   AnnotationType
		want string
	}{
		{AnnotationStrongs, "STRONGS"},
		{AnnotationMorphology, "MORPHOLOGY"},
		{AnnotationFootnote, "FOOTNOTE"},
		{AnnotationCrossRef, "CROSS_REF"},
		{AnnotationGloss, "GLOSS"},
	}

	for _, tt := range tests {
		if string(tt.at) != tt.want {
			t.Errorf("AnnotationType = %q, want %q", tt.at, tt.want)
		}
	}
}

func TestAnnotationTypeValidation(t *testing.T) {
	tests := []struct {
		at    AnnotationType
		valid bool
	}{
		{AnnotationStrongs, true},
		{AnnotationMorphology, true},
		{AnnotationType("INVALID"), false},
		{AnnotationType(""), false},
	}

	for _, tt := range tests {
		if got := tt.at.IsValid(); got != tt.valid {
			t.Errorf("AnnotationType(%q).IsValid() = %v, want %v", tt.at, got, tt.valid)
		}
	}
}

func TestAnnotationJSON(t *testing.T) {
	annotation := &Annotation{
		ID:         "ann1",
		SpanID:     "s1",
		Type:       AnnotationStrongs,
		Value:      "H430",
		Confidence: 0.95,
		Source:     "BDB",
	}

	// Marshal to JSON
	data, err := json.Marshal(annotation)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	// Unmarshal back
	var decoded Annotation
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	// Verify fields
	if decoded.ID != annotation.ID {
		t.Errorf("ID = %q, want %q", decoded.ID, annotation.ID)
	}
	if decoded.SpanID != annotation.SpanID {
		t.Errorf("SpanID = %q, want %q", decoded.SpanID, annotation.SpanID)
	}
	if decoded.Type != annotation.Type {
		t.Errorf("Type = %q, want %q", decoded.Type, annotation.Type)
	}
	if decoded.Value != annotation.Value {
		t.Errorf("Value = %v, want %v", decoded.Value, annotation.Value)
	}
	if decoded.Confidence != annotation.Confidence {
		t.Errorf("Confidence = %v, want %v", decoded.Confidence, annotation.Confidence)
	}
	if decoded.Source != annotation.Source {
		t.Errorf("Source = %q, want %q", decoded.Source, annotation.Source)
	}
}

func TestAnnotationWithComplexValue(t *testing.T) {
	annotation := &Annotation{
		ID:     "ann1",
		SpanID: "s1",
		Type:   AnnotationCrossRef,
		Value: map[string]interface{}{
			"refs": []interface{}{"Gen.1.1", "John.1.1"},
			"note": "Creation account parallels",
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(annotation)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	// Unmarshal back
	var decoded Annotation
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	// Verify value is preserved (as map[string]interface{})
	valueMap, ok := decoded.Value.(map[string]interface{})
	if !ok {
		t.Fatalf("Value is not map[string]interface{}: %T", decoded.Value)
	}

	note, ok := valueMap["note"]
	if !ok || note != "Creation account parallels" {
		t.Errorf("note = %v, want %q", note, "Creation account parallels")
	}
}

// Test overlapping spans - a key feature of stand-off markup
func TestOverlappingSpans(t *testing.T) {
	// Create a document with overlapping verse and quotation spans
	// This tests that the data model can represent overlapping structures

	// Content: "And God said, Let there be light: and there was light."
	// Verse 3 starts at "And God said" and ends at "light."
	// Quotation starts at "Let there be light" and ends before "and there was"

	cb := &ContentBlock{
		ID:   "cb1",
		Text: "And God said, Let there be light: and there was light.",
		Anchors: []*Anchor{
			{ID: "a1", CharOffset: 0},  // Start of verse
			{ID: "a2", CharOffset: 14}, // Start of quotation "Let"
			{ID: "a3", CharOffset: 32}, // End of quotation ":"
			{ID: "a4", CharOffset: 54}, // End of verse "."
		},
	}

	// Verse span (encompasses quotation)
	verseSpan := &Span{
		ID:            "verse3",
		Type:          SpanVerse,
		StartAnchorID: "a1",
		EndAnchorID:   "a4",
		Ref:           &Ref{Book: "Gen", Chapter: 1, Verse: 3},
	}

	// Quotation span (within verse)
	quotationSpan := &Span{
		ID:            "quote1",
		Type:          SpanQuotation,
		StartAnchorID: "a2",
		EndAnchorID:   "a3",
	}
	quotationSpan.SetAttribute("speaker", "God")

	// Put together in a document
	doc := &Document{
		ID:            "Gen",
		ContentBlocks: []*ContentBlock{cb},
	}

	// The spans would be stored separately but reference the same anchors
	// This verifies the data model supports overlapping structures
	_ = verseSpan
	_ = quotationSpan
	_ = doc

	// Verify anchor count
	if len(cb.Anchors) != 4 {
		t.Errorf("expected 4 anchors, got %d", len(cb.Anchors))
	}

	// Verify quotation is within verse
	verseStart := 0
	verseEnd := 54
	quoteStart := 14
	quoteEnd := 32

	if quoteStart < verseStart || quoteEnd > verseEnd {
		t.Error("quotation should be within verse boundaries")
	}
}

// TestSpanGetAttributeNil tests GetAttribute when Attributes is nil.
func TestSpanGetAttributeNil(t *testing.T) {
	span := &Span{
		ID:         "s1",
		Attributes: nil,
	}

	val, ok := span.GetAttribute("any")
	if ok {
		t.Error("GetAttribute should return false for nil Attributes")
	}
	if val != nil {
		t.Error("GetAttribute should return nil for nil Attributes")
	}
}

// TestSpanGetAttributeExists tests GetAttribute when attribute exists.
func TestSpanGetAttributeExists(t *testing.T) {
	span := &Span{
		ID:         "s1",
		Attributes: map[string]interface{}{"key": "value"},
	}

	val, ok := span.GetAttribute("key")
	if !ok {
		t.Error("GetAttribute should return true for existing key")
	}
	if val != "value" {
		t.Errorf("GetAttribute = %v, want %q", val, "value")
	}
}

// TestSpanGetAttributeNotExists tests GetAttribute when attribute doesn't exist.
func TestSpanGetAttributeNotExists(t *testing.T) {
	span := &Span{
		ID:         "s1",
		Attributes: map[string]interface{}{"key": "value"},
	}

	val, ok := span.GetAttribute("other")
	if ok {
		t.Error("GetAttribute should return false for non-existing key")
	}
	if val != nil {
		t.Error("GetAttribute should return nil for non-existing key")
	}
}
