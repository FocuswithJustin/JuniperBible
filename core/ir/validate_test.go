package ir

import (
	"errors"
	"strings"
	"testing"
)

func TestValidateCorpusValid(t *testing.T) {
	corpus := &Corpus{
		ID:            "KJV",
		Version:       "1.0.0",
		ModuleType:    ModuleBible,
		Versification: "KJV",
		Language:      "en",
		Title:         "King James Version",
	}

	errs := ValidateCorpus(corpus)
	if len(errs) > 0 {
		t.Errorf("ValidateCorpus returned errors for valid corpus: %v", errs)
	}
}

func TestValidateCorpusMissingID(t *testing.T) {
	corpus := &Corpus{
		Version:    "1.0.0",
		ModuleType: ModuleBible,
	}

	errs := ValidateCorpus(corpus)
	if len(errs) == 0 {
		t.Error("ValidateCorpus should return error for missing ID")
	}

	// Check error message
	found := false
	for _, err := range errs {
		if strings.Contains(err.Error(), "ID") {
			found = true
			break
		}
	}
	if !found {
		t.Error("error should mention ID")
	}
}

func TestValidateCorpusMissingVersion(t *testing.T) {
	corpus := &Corpus{
		ID:         "test",
		ModuleType: ModuleBible,
	}

	errs := ValidateCorpus(corpus)
	if len(errs) == 0 {
		t.Error("ValidateCorpus should return error for missing Version")
	}
}

func TestValidateCorpusInvalidModuleType(t *testing.T) {
	corpus := &Corpus{
		ID:         "test",
		Version:    "1.0.0",
		ModuleType: ModuleType("INVALID"),
	}

	errs := ValidateCorpus(corpus)
	if len(errs) == 0 {
		t.Error("ValidateCorpus should return error for invalid ModuleType")
	}
}

func TestValidateDocumentValid(t *testing.T) {
	doc := &Document{
		ID:    "Gen",
		Title: "Genesis",
		Order: 1,
	}

	errs := ValidateDocument(doc)
	if len(errs) > 0 {
		t.Errorf("ValidateDocument returned errors for valid document: %v", errs)
	}
}

func TestValidateDocumentMissingID(t *testing.T) {
	doc := &Document{
		Title: "Genesis",
		Order: 1,
	}

	errs := ValidateDocument(doc)
	if len(errs) == 0 {
		t.Error("ValidateDocument should return error for missing ID")
	}
}

func TestValidateContentBlockValid(t *testing.T) {
	cb := &ContentBlock{
		ID:       "cb1",
		Sequence: 0,
		Text:     "In the beginning",
	}
	cb.ComputeHash()

	errs := ValidateContentBlock(cb)
	if len(errs) > 0 {
		t.Errorf("ValidateContentBlock returned errors for valid block: %v", errs)
	}
}

func TestValidateContentBlockMissingID(t *testing.T) {
	cb := &ContentBlock{
		Sequence: 0,
		Text:     "Some text",
	}

	errs := ValidateContentBlock(cb)
	if len(errs) == 0 {
		t.Error("ValidateContentBlock should return error for missing ID")
	}
}

func TestValidateContentBlockInvalidHash(t *testing.T) {
	cb := &ContentBlock{
		ID:   "cb1",
		Text: "Some text",
		Hash: "invalid_hash",
	}

	errs := ValidateContentBlock(cb)
	if len(errs) == 0 {
		t.Error("ValidateContentBlock should return error for invalid hash")
	}
}

func TestValidateSpanValid(t *testing.T) {
	span := &Span{
		ID:            "s1",
		Type:          SpanVerse,
		StartAnchorID: "a1",
		EndAnchorID:   "a2",
	}

	errs := ValidateSpan(span)
	if len(errs) > 0 {
		t.Errorf("ValidateSpan returned errors for valid span: %v", errs)
	}
}

func TestValidateSpanMissingID(t *testing.T) {
	span := &Span{
		Type:          SpanVerse,
		StartAnchorID: "a1",
		EndAnchorID:   "a2",
	}

	errs := ValidateSpan(span)
	if len(errs) == 0 {
		t.Error("ValidateSpan should return error for missing ID")
	}
}

func TestValidateSpanInvalidType(t *testing.T) {
	span := &Span{
		ID:            "s1",
		Type:          SpanType("INVALID"),
		StartAnchorID: "a1",
		EndAnchorID:   "a2",
	}

	errs := ValidateSpan(span)
	if len(errs) == 0 {
		t.Error("ValidateSpan should return error for invalid type")
	}
}

func TestValidateSpanMissingAnchors(t *testing.T) {
	span := &Span{
		ID:   "s1",
		Type: SpanVerse,
	}

	errs := ValidateSpan(span)
	if len(errs) < 2 {
		t.Error("ValidateSpan should return errors for missing anchors")
	}
}

func TestValidateRefValid(t *testing.T) {
	ref := &Ref{
		Book:    "Gen",
		Chapter: 1,
		Verse:   1,
	}

	errs := ValidateRef(ref)
	if len(errs) > 0 {
		t.Errorf("ValidateRef returned errors for valid ref: %v", errs)
	}
}

func TestValidateRefMissingBook(t *testing.T) {
	ref := &Ref{
		Chapter: 1,
		Verse:   1,
	}

	errs := ValidateRef(ref)
	if len(errs) == 0 {
		t.Error("ValidateRef should return error for missing Book")
	}
}

func TestValidateRefNegativeChapter(t *testing.T) {
	ref := &Ref{
		Book:    "Gen",
		Chapter: -1,
	}

	errs := ValidateRef(ref)
	if len(errs) == 0 {
		t.Error("ValidateRef should return error for negative chapter")
	}
}

func TestValidateRefInvalidRange(t *testing.T) {
	ref := &Ref{
		Book:     "Gen",
		Chapter:  1,
		Verse:    10,
		VerseEnd: 5, // End before start
	}

	errs := ValidateRef(ref)
	if len(errs) == 0 {
		t.Error("ValidateRef should return error for invalid range")
	}
}

func TestValidateAnnotationValid(t *testing.T) {
	ann := &Annotation{
		ID:     "a1",
		SpanID: "s1",
		Type:   AnnotationStrongs,
		Value:  "H430",
	}

	errs := ValidateAnnotation(ann)
	if len(errs) > 0 {
		t.Errorf("ValidateAnnotation returned errors for valid annotation: %v", errs)
	}
}

func TestValidateAnnotationMissingID(t *testing.T) {
	ann := &Annotation{
		SpanID: "s1",
		Type:   AnnotationStrongs,
		Value:  "H430",
	}

	errs := ValidateAnnotation(ann)
	if len(errs) == 0 {
		t.Error("ValidateAnnotation should return error for missing ID")
	}
}

func TestValidateAnnotationInvalidType(t *testing.T) {
	ann := &Annotation{
		ID:     "a1",
		SpanID: "s1",
		Type:   AnnotationType("INVALID"),
		Value:  "test",
	}

	errs := ValidateAnnotation(ann)
	if len(errs) == 0 {
		t.Error("ValidateAnnotation should return error for invalid type")
	}
}

func TestValidateAnnotationInvalidConfidence(t *testing.T) {
	ann := &Annotation{
		ID:         "a1",
		SpanID:     "s1",
		Type:       AnnotationStrongs,
		Value:      "H430",
		Confidence: 1.5, // Invalid: should be 0-1
	}

	errs := ValidateAnnotation(ann)
	if len(errs) == 0 {
		t.Error("ValidateAnnotation should return error for invalid confidence")
	}
}

func TestValidateLossReportValid(t *testing.T) {
	report := &LossReport{
		SourceFormat: "SWORD",
		TargetFormat: "IR",
		LossClass:    LossL1,
	}

	errs := ValidateLossReport(report)
	if len(errs) > 0 {
		t.Errorf("ValidateLossReport returned errors for valid report: %v", errs)
	}
}

func TestValidateLossReportMissingFormats(t *testing.T) {
	report := &LossReport{
		LossClass: LossL1,
	}

	errs := ValidateLossReport(report)
	if len(errs) < 2 {
		t.Error("ValidateLossReport should return errors for missing formats")
	}
}

func TestValidateLossReportInvalidClass(t *testing.T) {
	report := &LossReport{
		SourceFormat: "SWORD",
		TargetFormat: "IR",
		LossClass:    LossClass("INVALID"),
	}

	errs := ValidateLossReport(report)
	if len(errs) == 0 {
		t.Error("ValidateLossReport should return error for invalid LossClass")
	}
}

func TestValidateCorpusWithDocuments(t *testing.T) {
	corpus := &Corpus{
		ID:         "test",
		Version:    "1.0.0",
		ModuleType: ModuleBible,
		Documents: []*Document{
			{ID: "Gen", Title: "Genesis", Order: 1},
			{ID: "", Title: "Invalid", Order: 2}, // Invalid - missing ID
		},
	}

	errs := ValidateCorpus(corpus)
	if len(errs) == 0 {
		t.Error("ValidateCorpus should return error for invalid document")
	}
}

func TestValidateCorpusWithContentBlocks(t *testing.T) {
	corpus := &Corpus{
		ID:         "test",
		Version:    "1.0.0",
		ModuleType: ModuleBible,
		Documents: []*Document{
			{
				ID:    "Gen",
				Title: "Genesis",
				Order: 1,
				ContentBlocks: []*ContentBlock{
					{ID: "cb1", Text: "Valid"},
					{ID: "", Text: "Invalid"}, // Invalid - missing ID
				},
			},
		},
	}

	errs := ValidateCorpus(corpus)
	if len(errs) == 0 {
		t.Error("ValidateCorpus should return error for invalid content block")
	}
}

// TestValidate tests the convenience Validate function.
func TestValidate(t *testing.T) {
	corpus := &Corpus{
		ID:         "test",
		Version:    "1.0.0",
		ModuleType: ModuleBible,
	}

	errs := Validate(corpus)
	if len(errs) > 0 {
		t.Errorf("Validate returned errors for valid corpus: %v", errs)
	}
}

// TestIsValid tests the convenience IsValid function.
func TestIsValid(t *testing.T) {
	valid := &Corpus{
		ID:         "test",
		Version:    "1.0.0",
		ModuleType: ModuleBible,
	}

	if !IsValid(valid) {
		t.Error("IsValid returned false for valid corpus")
	}

	invalid := &Corpus{
		Version: "1.0.0", // Missing ID
	}

	if IsValid(invalid) {
		t.Error("IsValid returned true for invalid corpus")
	}
}

// TestValidateMappingTable tests ValidateMappingTable with valid input.
func TestValidateMappingTableValid(t *testing.T) {
	mt := &MappingTable{
		ID:         "mt1",
		FromSystem: VersificationKJV,
		ToSystem:   VersificationLXX,
		Mappings: []*RefMapping{
			{
				From: &Ref{Book: "Gen", Chapter: 1, Verse: 1},
				To:   &Ref{Book: "Gen", Chapter: 1, Verse: 1},
			},
		},
	}

	errs := ValidateMappingTable(mt)
	if len(errs) > 0 {
		t.Errorf("ValidateMappingTable returned errors for valid table: %v", errs)
	}
}

// TestValidateMappingTableMissingID tests ValidateMappingTable with missing ID.
func TestValidateMappingTableMissingID(t *testing.T) {
	mt := &MappingTable{
		FromSystem: VersificationKJV,
		ToSystem:   VersificationLXX,
	}

	errs := ValidateMappingTable(mt)
	if len(errs) == 0 {
		t.Error("ValidateMappingTable should return error for missing ID")
	}
}

// TestValidateMappingTableInvalidSystems tests invalid versification systems.
func TestValidateMappingTableInvalidSystems(t *testing.T) {
	mt := &MappingTable{
		ID:         "mt1",
		FromSystem: VersificationID("INVALID"),
		ToSystem:   VersificationID("ALSO_INVALID"),
	}

	errs := ValidateMappingTable(mt)
	if len(errs) < 2 {
		t.Error("ValidateMappingTable should return errors for invalid systems")
	}
}

// TestValidateMappingTableInvalidRefs tests invalid refs in mappings.
func TestValidateMappingTableInvalidRefs(t *testing.T) {
	mt := &MappingTable{
		ID:         "mt1",
		FromSystem: VersificationKJV,
		ToSystem:   VersificationLXX,
		Mappings: []*RefMapping{
			{
				From: &Ref{Book: "", Chapter: -1}, // Invalid from ref
				To:   &Ref{Book: "", Verse: -1},   // Invalid to ref
			},
		},
	}

	errs := ValidateMappingTable(mt)
	if len(errs) < 2 {
		t.Error("ValidateMappingTable should return errors for invalid refs")
	}
}

// TestValidateCorpusWithMappingTables tests corpus with mapping tables.
func TestValidateCorpusWithMappingTables(t *testing.T) {
	corpus := &Corpus{
		ID:         "test",
		Version:    "1.0.0",
		ModuleType: ModuleBible,
		MappingTables: []*MappingTable{
			{ID: "mt1", FromSystem: VersificationKJV, ToSystem: VersificationLXX},
			{ID: "", FromSystem: VersificationKJV}, // Invalid - missing ID
		},
	}

	errs := ValidateCorpus(corpus)
	if len(errs) == 0 {
		t.Error("ValidateCorpus should return error for invalid mapping table")
	}
}

// TestValidateDocumentWithCanonicalRef tests document validation with canonical ref.
func TestValidateDocumentWithCanonicalRef(t *testing.T) {
	doc := &Document{
		ID:    "Gen",
		Title: "Genesis",
		Order: 1,
		CanonicalRef: &Ref{
			Book:    "", // Invalid - missing book
			Chapter: -1, // Invalid - negative chapter
		},
	}

	errs := ValidateDocument(doc)
	if len(errs) == 0 {
		t.Error("ValidateDocument should return errors for invalid canonical ref")
	}
}

// TestValidateDocumentWithAnnotations tests document validation with annotations.
func TestValidateDocumentWithAnnotations(t *testing.T) {
	doc := &Document{
		ID:    "Gen",
		Title: "Genesis",
		Order: 1,
		Annotations: []*Annotation{
			{ID: "a1", SpanID: "s1", Type: AnnotationStrongs, Value: "H430"},
			{ID: "", SpanID: "s2"}, // Invalid - missing ID
		},
	}

	errs := ValidateDocument(doc)
	if len(errs) == 0 {
		t.Error("ValidateDocument should return errors for invalid annotation")
	}
}

// TestValidateContentBlockNegativeSequence tests negative sequence validation.
func TestValidateContentBlockNegativeSequence(t *testing.T) {
	cb := &ContentBlock{
		ID:       "cb1",
		Sequence: -1,
		Text:     "Test",
	}

	errs := ValidateContentBlock(cb)
	if len(errs) == 0 {
		t.Error("ValidateContentBlock should return error for negative sequence")
	}
}

// TestValidateContentBlockInvalidTokens tests invalid token offsets.
func TestValidateContentBlockInvalidTokens(t *testing.T) {
	cb := &ContentBlock{
		ID:       "cb1",
		Sequence: 0,
		Text:     "Hello world",
		Tokens: []*Token{
			{ID: "t1", CharStart: -1, CharEnd: 5}, // Invalid start
			{ID: "t2", CharStart: 10, CharEnd: 5}, // End before start
		},
	}

	errs := ValidateContentBlock(cb)
	if len(errs) < 2 {
		t.Error("ValidateContentBlock should return errors for invalid tokens")
	}
}

// TestValidateContentBlockInvalidAnchors tests invalid anchor offsets.
func TestValidateContentBlockInvalidAnchors(t *testing.T) {
	cb := &ContentBlock{
		ID:       "cb1",
		Sequence: 0,
		Text:     "Hello world",
		Anchors: []*Anchor{
			{ID: "a1", CharOffset: -1}, // Invalid negative offset
		},
	}

	errs := ValidateContentBlock(cb)
	if len(errs) == 0 {
		t.Error("ValidateContentBlock should return error for negative anchor offset")
	}
}

// TestValidateSpanWithRef tests span validation with ref.
func TestValidateSpanWithRef(t *testing.T) {
	span := &Span{
		ID:            "s1",
		Type:          SpanVerse,
		StartAnchorID: "a1",
		EndAnchorID:   "a2",
		Ref: &Ref{
			Book:    "", // Invalid - missing book
			Chapter: -1, // Invalid
		},
	}

	errs := ValidateSpan(span)
	if len(errs) == 0 {
		t.Error("ValidateSpan should return errors for invalid ref")
	}
}

// TestValidateRefNegativeVerse tests negative verse validation.
func TestValidateRefNegativeVerse(t *testing.T) {
	ref := &Ref{
		Book:    "Gen",
		Chapter: 1,
		Verse:   -1,
	}

	errs := ValidateRef(ref)
	if len(errs) == 0 {
		t.Error("ValidateRef should return error for negative verse")
	}
}

// TestValidateAnnotationMissingSpanID tests missing span ID validation.
func TestValidateAnnotationMissingSpanID(t *testing.T) {
	ann := &Annotation{
		ID:    "a1",
		Type:  AnnotationStrongs,
		Value: "H430",
	}

	errs := ValidateAnnotation(ann)
	if len(errs) == 0 {
		t.Error("ValidateAnnotation should return error for missing SpanID")
	}
}

// TestValidateAnnotationNegativeConfidence tests negative confidence validation.
func TestValidateAnnotationNegativeConfidence(t *testing.T) {
	ann := &Annotation{
		ID:         "a1",
		SpanID:     "s1",
		Type:       AnnotationStrongs,
		Value:      "H430",
		Confidence: -0.5,
	}

	errs := ValidateAnnotation(ann)
	if len(errs) == 0 {
		t.Error("ValidateAnnotation should return error for negative confidence")
	}
}

// TestValidationErrorWithoutPath tests ValidationError with empty path.
func TestValidationErrorWithoutPath(t *testing.T) {
	err := &ValidationError{
		Path:    "",
		Message: "test error",
	}

	if err.Error() != "test error" {
		t.Errorf("Error() = %q, want %q", err.Error(), "test error")
	}
}

// TestValidationErrorWithPath tests ValidationError with path.
func TestValidationErrorWithPath(t *testing.T) {
	err := &ValidationError{
		Path:    "corpus.documents[0]",
		Message: "ID is required",
	}

	want := "corpus.documents[0]: ID is required"
	if err.Error() != want {
		t.Errorf("Error() = %q, want %q", err.Error(), want)
	}
}

// TestValidateCorpusNonValidationError tests ValidateCorpus when document validation returns non-ValidationError.
func TestValidateCorpusNonValidationError(t *testing.T) {
	orig := validateDocumentFn
	defer func() { validateDocumentFn = orig }()

	// Inject function that returns a regular error
	validateDocumentFn = func(d *Document) []error {
		return []error{errors.New("regular error")}
	}

	corpus := &Corpus{
		ID:         "test",
		Version:    "1.0.0",
		ModuleType: ModuleBible,
		Documents: []*Document{
			{ID: "Gen", Title: "Genesis"},
		},
	}

	errs := ValidateCorpus(corpus)
	if len(errs) == 0 {
		t.Error("ValidateCorpus should propagate non-ValidationError")
	}
}

// TestValidateCorpusMappingTableNonValidationError tests ValidateCorpus when mapping table returns non-ValidationError.
func TestValidateCorpusMappingTableNonValidationError(t *testing.T) {
	orig := validateMappingTableFn
	defer func() { validateMappingTableFn = orig }()

	// Inject function that returns a regular error
	validateMappingTableFn = func(mt *MappingTable) []error {
		return []error{errors.New("mapping table error")}
	}

	corpus := &Corpus{
		ID:         "test",
		Version:    "1.0.0",
		ModuleType: ModuleBible,
		MappingTables: []*MappingTable{
			{ID: "mt1", FromSystem: VersificationKJV},
		},
	}

	errs := ValidateCorpus(corpus)
	if len(errs) == 0 {
		t.Error("ValidateCorpus should propagate mapping table non-ValidationError")
	}
}

// TestValidateDocumentRefNonValidationError tests ValidateDocument when ref validation returns non-ValidationError.
func TestValidateDocumentRefNonValidationError(t *testing.T) {
	orig := validateRefFn
	defer func() { validateRefFn = orig }()

	// Inject function that returns a regular error
	validateRefFn = func(r *Ref) []error {
		return []error{errors.New("ref error")}
	}

	doc := &Document{
		ID:           "Gen",
		Title:        "Genesis",
		CanonicalRef: &Ref{Book: "Gen", Chapter: 1},
	}

	errs := ValidateDocument(doc)
	if len(errs) == 0 {
		t.Error("ValidateDocument should propagate ref non-ValidationError")
	}
}

// TestValidateDocumentContentBlockNonValidationError tests ValidateDocument when content block returns non-ValidationError.
func TestValidateDocumentContentBlockNonValidationError(t *testing.T) {
	orig := validateContentBlockFn
	defer func() { validateContentBlockFn = orig }()

	// Inject function that returns a regular error
	validateContentBlockFn = func(cb *ContentBlock) []error {
		return []error{errors.New("content block error")}
	}

	doc := &Document{
		ID:    "Gen",
		Title: "Genesis",
		ContentBlocks: []*ContentBlock{
			{ID: "cb1", Text: "Test"},
		},
	}

	errs := ValidateDocument(doc)
	if len(errs) == 0 {
		t.Error("ValidateDocument should propagate content block non-ValidationError")
	}
}

// TestValidateDocumentAnnotationNonValidationError tests ValidateDocument when annotation returns non-ValidationError.
func TestValidateDocumentAnnotationNonValidationError(t *testing.T) {
	orig := validateAnnotationFn
	defer func() { validateAnnotationFn = orig }()

	// Inject function that returns a regular error
	validateAnnotationFn = func(a *Annotation) []error {
		return []error{errors.New("annotation error")}
	}

	doc := &Document{
		ID:    "Gen",
		Title: "Genesis",
		Annotations: []*Annotation{
			{ID: "a1", SpanID: "s1", Type: AnnotationStrongs},
		},
	}

	errs := ValidateDocument(doc)
	if len(errs) == 0 {
		t.Error("ValidateDocument should propagate annotation non-ValidationError")
	}
}

// TestValidateSpanRefNonValidationError tests ValidateSpan when ref validation returns non-ValidationError.
func TestValidateSpanRefNonValidationError(t *testing.T) {
	orig := validateRefFn
	defer func() { validateRefFn = orig }()

	// Inject function that returns a regular error
	validateRefFn = func(r *Ref) []error {
		return []error{errors.New("span ref error")}
	}

	span := &Span{
		ID:            "s1",
		Type:          SpanVerse,
		StartAnchorID: "a1",
		EndAnchorID:   "a2",
		Ref:           &Ref{Book: "Gen", Chapter: 1, Verse: 1},
	}

	errs := ValidateSpan(span)
	if len(errs) == 0 {
		t.Error("ValidateSpan should propagate ref non-ValidationError")
	}
}

// TestValidateMappingTableRefNonValidationError tests ValidateMappingTable when ref validation returns non-ValidationError.
func TestValidateMappingTableRefNonValidationError(t *testing.T) {
	orig := validateRefFn
	defer func() { validateRefFn = orig }()

	callCount := 0
	// Inject function that returns a regular error
	validateRefFn = func(r *Ref) []error {
		callCount++
		return []error{errors.New("mapping ref error")}
	}

	mt := &MappingTable{
		ID:         "mt1",
		FromSystem: VersificationKJV,
		ToSystem:   VersificationLXX,
		Mappings: []*RefMapping{
			{
				From: &Ref{Book: "Gen", Chapter: 1, Verse: 1},
				To:   &Ref{Book: "Gen", Chapter: 1, Verse: 1},
			},
		},
	}

	errs := ValidateMappingTable(mt)
	if len(errs) == 0 {
		t.Error("ValidateMappingTable should propagate ref non-ValidationError")
	}
}
