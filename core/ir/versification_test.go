package ir

import (
	"encoding/json"
	"testing"
)

func TestVersificationIDConstants(t *testing.T) {
	tests := []struct {
		vid  VersificationID
		want string
	}{
		{VersificationKJV, "KJV"},
		{VersificationCatholic, "Catholic"},
		{VersificationLXX, "LXX"},
		{VersificationVulgate, "Vulgate"},
		{VersificationEthiopian, "Ethiopian"},
		{VersificationSynodal, "Synodal"},
		{VersificationMT, "MT"},
		{VersificationNRSV, "NRSV"},
	}

	for _, tt := range tests {
		if string(tt.vid) != tt.want {
			t.Errorf("VersificationID = %q, want %q", tt.vid, tt.want)
		}
	}
}

func TestVersificationIDValidation(t *testing.T) {
	tests := []struct {
		vid   VersificationID
		valid bool
	}{
		{VersificationKJV, true},
		{VersificationCatholic, true},
		{VersificationLXX, true},
		{VersificationID("INVALID"), false},
		{VersificationID(""), false},
	}

	for _, tt := range tests {
		if got := tt.vid.IsValid(); got != tt.valid {
			t.Errorf("VersificationID(%q).IsValid() = %v, want %v", tt.vid, got, tt.valid)
		}
	}
}

func TestMappingTypeConstants(t *testing.T) {
	tests := []struct {
		mt   MappingType
		want string
	}{
		{MappingExact, "exact"},
		{MappingSplit, "split"},
		{MappingMerge, "merge"},
		{MappingMissing, "missing"},
		{MappingAdded, "added"},
		{MappingReordered, "reordered"},
	}

	for _, tt := range tests {
		if string(tt.mt) != tt.want {
			t.Errorf("MappingType = %q, want %q", tt.mt, tt.want)
		}
	}
}

func TestRefMappingJSON(t *testing.T) {
	mapping := &RefMapping{
		From: &Ref{Book: "Ps", Chapter: 9, Verse: 1, OSISID: "Ps.9.1"},
		To:   &Ref{Book: "Ps", Chapter: 9, Verse: 1, OSISID: "Ps.9.1"},
		Type: MappingExact,
		Note: "Same in both systems",
	}

	// Marshal to JSON
	data, err := json.Marshal(mapping)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	// Unmarshal back
	var decoded RefMapping
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	// Verify fields
	if decoded.From.OSISID != mapping.From.OSISID {
		t.Errorf("From.OSISID = %q, want %q", decoded.From.OSISID, mapping.From.OSISID)
	}
	if decoded.To.OSISID != mapping.To.OSISID {
		t.Errorf("To.OSISID = %q, want %q", decoded.To.OSISID, mapping.To.OSISID)
	}
	if decoded.Type != mapping.Type {
		t.Errorf("Type = %q, want %q", decoded.Type, mapping.Type)
	}
	if decoded.Note != mapping.Note {
		t.Errorf("Note = %q, want %q", decoded.Note, mapping.Note)
	}
}

func TestRefMappingSplit(t *testing.T) {
	// A verse split into multiple verses
	mapping := &RefMapping{
		From: &Ref{Book: "Gen", Chapter: 31, Verse: 55, OSISID: "Gen.31.55"},
		Type: MappingSplit,
		ToRefs: []*Ref{
			{Book: "Gen", Chapter: 31, Verse: 55, OSISID: "Gen.31.55"},
			{Book: "Gen", Chapter: 32, Verse: 1, OSISID: "Gen.32.1"},
		},
		Note: "Verse numbering differs between MT and LXX",
	}

	// Marshal to JSON
	data, err := json.Marshal(mapping)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	// Unmarshal back
	var decoded RefMapping
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	// Verify ToRefs
	if len(decoded.ToRefs) != 2 {
		t.Fatalf("len(ToRefs) = %d, want 2", len(decoded.ToRefs))
	}
	if decoded.ToRefs[0].OSISID != "Gen.31.55" {
		t.Errorf("ToRefs[0].OSISID = %q, want %q", decoded.ToRefs[0].OSISID, "Gen.31.55")
	}
}

func TestMappingTableJSON(t *testing.T) {
	table := &MappingTable{
		ID:         "kjv-to-lxx",
		FromSystem: VersificationKJV,
		ToSystem:   VersificationLXX,
		Mappings: []*RefMapping{
			{
				From: &Ref{Book: "Ps", Chapter: 10, Verse: 1},
				To:   &Ref{Book: "Ps", Chapter: 9, Verse: 22},
				Type: MappingExact,
			},
		},
		Hash: "abc123",
	}

	// Marshal to JSON
	data, err := json.Marshal(table)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	// Unmarshal back
	var decoded MappingTable
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	// Verify fields
	if decoded.ID != table.ID {
		t.Errorf("ID = %q, want %q", decoded.ID, table.ID)
	}
	if decoded.FromSystem != table.FromSystem {
		t.Errorf("FromSystem = %q, want %q", decoded.FromSystem, table.FromSystem)
	}
	if decoded.ToSystem != table.ToSystem {
		t.Errorf("ToSystem = %q, want %q", decoded.ToSystem, table.ToSystem)
	}
	if len(decoded.Mappings) != 1 {
		t.Errorf("len(Mappings) = %d, want 1", len(decoded.Mappings))
	}
}

func TestMappingTableLookup(t *testing.T) {
	table := &MappingTable{
		ID:         "kjv-to-lxx",
		FromSystem: VersificationKJV,
		ToSystem:   VersificationLXX,
		Mappings: []*RefMapping{
			{
				From: &Ref{Book: "Ps", Chapter: 10, Verse: 1},
				To:   &Ref{Book: "Ps", Chapter: 9, Verse: 22},
				Type: MappingExact,
			},
			{
				From: &Ref{Book: "Ps", Chapter: 10, Verse: 2},
				To:   &Ref{Book: "Ps", Chapter: 9, Verse: 23},
				Type: MappingExact,
			},
		},
	}

	// Lookup existing mapping
	mapping := table.Lookup(&Ref{Book: "Ps", Chapter: 10, Verse: 1})
	if mapping == nil {
		t.Fatal("Lookup returned nil for existing mapping")
	}
	if mapping.To.Verse != 22 {
		t.Errorf("mapping.To.Verse = %d, want 22", mapping.To.Verse)
	}

	// Lookup non-existing mapping
	mapping = table.Lookup(&Ref{Book: "Ps", Chapter: 10, Verse: 99})
	if mapping != nil {
		t.Error("Lookup should return nil for non-existing mapping")
	}
}

func TestMappingTableAddMapping(t *testing.T) {
	table := &MappingTable{
		ID:         "test",
		FromSystem: VersificationKJV,
		ToSystem:   VersificationLXX,
	}

	from := &Ref{Book: "Ps", Chapter: 10, Verse: 1}
	to := &Ref{Book: "Ps", Chapter: 9, Verse: 22}
	table.AddMapping(from, to, MappingExact)

	if len(table.Mappings) != 1 {
		t.Fatalf("len(Mappings) = %d, want 1", len(table.Mappings))
	}

	mapping := table.Mappings[0]
	if mapping.From.Verse != 1 {
		t.Errorf("From.Verse = %d, want 1", mapping.From.Verse)
	}
	if mapping.To.Verse != 22 {
		t.Errorf("To.Verse = %d, want 22", mapping.To.Verse)
	}
	if mapping.Type != MappingExact {
		t.Errorf("Type = %q, want %q", mapping.Type, MappingExact)
	}
}

func TestMappingTableMapRef(t *testing.T) {
	table := &MappingTable{
		ID:         "kjv-to-lxx",
		FromSystem: VersificationKJV,
		ToSystem:   VersificationLXX,
		Mappings: []*RefMapping{
			{
				From: &Ref{Book: "Ps", Chapter: 10, Verse: 1},
				To:   &Ref{Book: "Ps", Chapter: 9, Verse: 22},
				Type: MappingExact,
			},
		},
	}

	// Map existing reference
	from := &Ref{Book: "Ps", Chapter: 10, Verse: 1}
	to := table.MapRef(from)
	if to.Chapter != 9 || to.Verse != 22 {
		t.Errorf("MapRef returned (%d, %d), want (9, 22)", to.Chapter, to.Verse)
	}

	// Map non-existing reference (identity)
	from = &Ref{Book: "Gen", Chapter: 1, Verse: 1}
	to = table.MapRef(from)
	if to != from {
		t.Error("MapRef should return identity for unmapped reference")
	}
}

func TestEmptyMappingTable(t *testing.T) {
	table := &MappingTable{}

	data, err := json.Marshal(table)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var decoded MappingTable
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	// Empty table should have nil mappings
	if decoded.Mappings != nil && len(decoded.Mappings) != 0 {
		t.Error("Mappings should be nil or empty")
	}
}

// Phase 16.1: Test new versification systems
func TestNewVersificationSystems(t *testing.T) {
	// Test that all new versification systems are valid
	newSystems := []VersificationID{
		VersificationArmenian,
		VersificationGeorgian,
		VersificationSlavonic,
		VersificationSyriac,
		VersificationArabic,
		VersificationDSS,
		VersificationSamaritan,
		VersificationBHS,
		VersificationNA28,
	}

	for _, sys := range newSystems {
		if !sys.IsValid() {
			t.Errorf("VersificationID(%q) should be valid", sys)
		}
	}
}

// Phase 16.1: Test MappingRegistry
func TestMappingRegistry(t *testing.T) {
	registry := NewMappingRegistry()

	// Create KJV -> LXX mapping table
	kjvToLXX := &MappingTable{
		ID:         "kjv-to-lxx",
		FromSystem: VersificationKJV,
		ToSystem:   VersificationLXX,
	}
	kjvToLXX.AddMapping(
		&Ref{Book: "Ps", Chapter: 10, Verse: 1},
		&Ref{Book: "Ps", Chapter: 9, Verse: 22},
		MappingExact,
	)

	// Register the table
	registry.RegisterTable(kjvToLXX)

	// Get the table
	table := registry.GetTable(VersificationKJV, VersificationLXX)
	if table == nil {
		t.Fatal("GetTable returned nil for registered mapping")
	}
	if table.ID != "kjv-to-lxx" {
		t.Errorf("table.ID = %q, want %q", table.ID, "kjv-to-lxx")
	}

	// Get non-existent table
	table = registry.GetTable(VersificationKJV, VersificationMT)
	if table != nil {
		t.Error("GetTable should return nil for unregistered mapping")
	}
}

func TestMappingRegistryChainedMapping(t *testing.T) {
	registry := NewMappingRegistry()

	// Create KJV -> MT mapping
	kjvToMT := &MappingTable{
		ID:         "kjv-to-mt",
		FromSystem: VersificationKJV,
		ToSystem:   VersificationMT,
	}
	kjvToMT.AddMapping(
		&Ref{Book: "Gen", Chapter: 32, Verse: 1},
		&Ref{Book: "Gen", Chapter: 31, Verse: 55},
		MappingExact,
	)

	// Create MT -> LXX mapping
	mtToLXX := &MappingTable{
		ID:         "mt-to-lxx",
		FromSystem: VersificationMT,
		ToSystem:   VersificationLXX,
	}
	mtToLXX.AddMapping(
		&Ref{Book: "Gen", Chapter: 31, Verse: 55},
		&Ref{Book: "Gen", Chapter: 31, Verse: 55},
		MappingExact,
	)

	registry.RegisterTable(kjvToMT)
	registry.RegisterTable(mtToLXX)

	// Get chained mapping KJV -> LXX via MT
	chainedTable := registry.GetChainedMapping(VersificationKJV, VersificationLXX)
	if chainedTable == nil {
		t.Fatal("GetChainedMapping returned nil")
	}

	// Verify the chain works
	ref := &Ref{Book: "Gen", Chapter: 32, Verse: 1}
	mapped := chainedTable.MapRef(ref)
	if mapped.Chapter != 31 || mapped.Verse != 55 {
		t.Errorf("Chained mapping failed: got (%d, %d), want (31, 55)", mapped.Chapter, mapped.Verse)
	}
}

func TestMapRefBetweenSystems(t *testing.T) {
	registry := NewMappingRegistry()

	// Set up mappings
	kjvToLXX := &MappingTable{
		ID:         "kjv-to-lxx",
		FromSystem: VersificationKJV,
		ToSystem:   VersificationLXX,
	}
	kjvToLXX.AddMapping(
		&Ref{Book: "Ps", Chapter: 10, Verse: 1},
		&Ref{Book: "Ps", Chapter: 9, Verse: 22},
		MappingExact,
	)
	registry.RegisterTable(kjvToLXX)

	// Map reference between systems
	from := &Ref{Book: "Ps", Chapter: 10, Verse: 1}
	mapped, err := registry.MapRefBetweenSystems(from, VersificationKJV, VersificationLXX)
	if err != nil {
		t.Fatalf("MapRefBetweenSystems failed: %v", err)
	}
	if mapped.Chapter != 9 || mapped.Verse != 22 {
		t.Errorf("MapRefBetweenSystems returned (%d, %d), want (9, 22)", mapped.Chapter, mapped.Verse)
	}
}

// Phase 16.1: Test SplitRef
func TestSplitRef(t *testing.T) {
	// Test splitting Gen.31.55 (MT) into Gen.31.55 + Gen.32.1 (KJV)
	mtRef := &Ref{Book: "Gen", Chapter: 31, Verse: 55, OSISID: "Gen.31.55"}

	// Create a split mapping
	mapping := &RefMapping{
		From: mtRef,
		Type: MappingSplit,
		ToRefs: []*Ref{
			{Book: "Gen", Chapter: 31, Verse: 55, OSISID: "Gen.31.55"},
			{Book: "Gen", Chapter: 32, Verse: 1, OSISID: "Gen.32.1"},
		},
	}

	// SplitRef should return the split references
	refs := SplitRef(mapping)
	if len(refs) != 2 {
		t.Fatalf("SplitRef returned %d refs, want 2", len(refs))
	}
	if refs[0].Chapter != 31 || refs[0].Verse != 55 {
		t.Errorf("First split ref: got (%d, %d), want (31, 55)", refs[0].Chapter, refs[0].Verse)
	}
	if refs[1].Chapter != 32 || refs[1].Verse != 1 {
		t.Errorf("Second split ref: got (%d, %d), want (32, 1)", refs[1].Chapter, refs[1].Verse)
	}
}

// Phase 16.1: Test MergeRefs
func TestMergeRefs(t *testing.T) {
	// Test merging Ps.9.22 + Ps.9.23 (LXX) into Ps.10.1 (KJV)
	lxxRefs := []*Ref{
		{Book: "Ps", Chapter: 9, Verse: 22, OSISID: "Ps.9.22"},
		{Book: "Ps", Chapter: 9, Verse: 23, OSISID: "Ps.9.23"},
	}

	// MergeRefs should return a single merged reference
	merged := MergeRefs(lxxRefs)
	if merged == nil {
		t.Fatal("MergeRefs returned nil")
	}
	// The merged ref should point to the first verse
	if merged.Chapter != 9 || merged.Verse != 22 {
		t.Errorf("MergeRefs: got (%d, %d), want (9, 22)", merged.Chapter, merged.Verse)
	}
}

// Phase 16.1: Test ApplyToCorpus
func TestMappingTableApplyToCorpus(t *testing.T) {
	// Create a simple corpus with KJV versification
	corpus := &Corpus{
		ID:            "test-corpus",
		Versification: string(VersificationKJV),
		Documents: []*Document{
			{
				ID:           "ps10",
				Title:        "Psalm 10",
				CanonicalRef: &Ref{Book: "Ps", Chapter: 10, Verse: 1},
				ContentBlocks: []*ContentBlock{
					{ID: "block1", Sequence: 0, Text: "Why standest thou afar off?"},
					{ID: "block2", Sequence: 1, Text: "In his pride the wicked persecutes the poor."},
				},
			},
		},
	}

	// Create KJV -> LXX mapping
	table := &MappingTable{
		ID:         "kjv-to-lxx",
		FromSystem: VersificationKJV,
		ToSystem:   VersificationLXX,
	}
	table.AddMapping(
		&Ref{Book: "Ps", Chapter: 10, Verse: 1},
		&Ref{Book: "Ps", Chapter: 9, Verse: 22},
		MappingExact,
	)

	// Apply mapping to corpus
	mappedCorpus, lossReport, err := table.ApplyToCorpus(corpus)
	if err != nil {
		t.Fatalf("ApplyToCorpus failed: %v", err)
	}
	if lossReport == nil {
		t.Error("ApplyToCorpus should return a LossReport")
	}

	// Verify mapped corpus has LXX versification
	if mappedCorpus.Versification != string(VersificationLXX) {
		t.Errorf("Versification = %q, want %q", mappedCorpus.Versification, VersificationLXX)
	}

	// Verify canonical reference was mapped
	if len(mappedCorpus.Documents) != 1 {
		t.Fatalf("len(Documents) = %d, want 1", len(mappedCorpus.Documents))
	}
	docRef := mappedCorpus.Documents[0].CanonicalRef
	if docRef == nil {
		t.Fatal("CanonicalRef should not be nil")
	}
	if docRef.Chapter != 9 || docRef.Verse != 22 {
		t.Errorf("CanonicalRef = (%d, %d), want (9, 22)", docRef.Chapter, docRef.Verse)
	}

	// Verify content blocks were copied
	blocks := mappedCorpus.Documents[0].ContentBlocks
	if len(blocks) != 2 {
		t.Fatalf("len(ContentBlocks) = %d, want 2", len(blocks))
	}
}

// Additional tests for SplitRef edge cases
func TestSplitRef_NonSplit(t *testing.T) {
	// Test a non-split mapping returns single ref slice
	mapping := &RefMapping{
		From: &Ref{Book: "Gen", Chapter: 1, Verse: 1},
		To:   &Ref{Book: "Gen", Chapter: 1, Verse: 1},
		Type: MappingExact,
	}

	refs := SplitRef(mapping)
	if len(refs) != 1 {
		t.Fatalf("SplitRef returned %d refs, want 1", len(refs))
	}
	if refs[0].Chapter != 1 || refs[0].Verse != 1 {
		t.Errorf("SplitRef returned wrong ref: got (%d, %d), want (1, 1)", refs[0].Chapter, refs[0].Verse)
	}
}

func TestSplitRef_NilTo(t *testing.T) {
	// Test a mapping with nil To field
	mapping := &RefMapping{
		From: &Ref{Book: "Gen", Chapter: 1, Verse: 1},
		Type: MappingMissing,
	}

	refs := SplitRef(mapping)
	if refs != nil {
		t.Errorf("SplitRef should return nil for mapping with no To, got %v", refs)
	}
}

func TestSplitRef_SplitWithEmptyToRefs(t *testing.T) {
	// Test a split mapping with empty ToRefs but valid To
	mapping := &RefMapping{
		From:   &Ref{Book: "Gen", Chapter: 1, Verse: 1},
		To:     &Ref{Book: "Gen", Chapter: 1, Verse: 1},
		Type:   MappingSplit,
		ToRefs: []*Ref{}, // Empty
	}

	refs := SplitRef(mapping)
	if len(refs) != 1 {
		t.Fatalf("SplitRef returned %d refs, want 1 (fallback to To)", len(refs))
	}
}

// Additional tests for MergeRefs edge cases
func TestMergeRefs_Empty(t *testing.T) {
	refs := MergeRefs([]*Ref{})
	if refs != nil {
		t.Error("MergeRefs of empty slice should return nil")
	}
}

func TestMergeRefs_SingleRef(t *testing.T) {
	single := &Ref{Book: "Ps", Chapter: 9, Verse: 22}
	merged := MergeRefs([]*Ref{single})
	if merged != single {
		t.Error("MergeRefs of single ref should return that ref")
	}
}

// Test MapRefBetweenSystems with identity mapping
func TestMapRefBetweenSystems_SameSystem(t *testing.T) {
	registry := NewMappingRegistry()

	ref := &Ref{Book: "Gen", Chapter: 1, Verse: 1}
	mapped, err := registry.MapRefBetweenSystems(ref, VersificationKJV, VersificationKJV)
	if err != nil {
		t.Fatalf("MapRefBetweenSystems failed: %v", err)
	}
	if mapped != ref {
		t.Error("MapRefBetweenSystems between same systems should return identity")
	}
}

// Test MapRefBetweenSystems with no mapping found - returns identity
func TestMapRefBetweenSystems_NoMapping(t *testing.T) {
	registry := NewMappingRegistry()

	ref := &Ref{Book: "Gen", Chapter: 1, Verse: 1}
	mapped, err := registry.MapRefBetweenSystems(ref, VersificationKJV, VersificationLXX)
	if err != nil {
		t.Fatalf("MapRefBetweenSystems failed: %v", err)
	}
	// When no mapping exists, identity is returned
	if mapped != ref {
		t.Error("MapRefBetweenSystems should return identity when no mapping exists")
	}
}

// TestGetChainedMappingDirect tests GetChainedMapping when direct mapping exists.
func TestGetChainedMappingDirect(t *testing.T) {
	registry := NewMappingRegistry()

	// Register a direct mapping
	directTable := &MappingTable{
		ID:         "kjv-to-lxx",
		FromSystem: VersificationKJV,
		ToSystem:   VersificationLXX,
	}
	directTable.AddMapping(
		&Ref{Book: "Ps", Chapter: 10, Verse: 1},
		&Ref{Book: "Ps", Chapter: 9, Verse: 22},
		MappingExact,
	)
	registry.RegisterTable(directTable)

	// GetChainedMapping should return the direct mapping
	chained := registry.GetChainedMapping(VersificationKJV, VersificationLXX)
	if chained == nil {
		t.Fatal("GetChainedMapping returned nil")
	}
	if chained.ID != "kjv-to-lxx" {
		t.Errorf("Expected direct table, got %q", chained.ID)
	}
}

// TestBuildChainedTableNilIntermediate tests buildChainedTable with nil intermediate refs.
func TestBuildChainedTableNilIntermediate(t *testing.T) {
	registry := NewMappingRegistry()

	// Create first table with a mapping that has nil To
	first := &MappingTable{
		ID:         "first",
		FromSystem: VersificationKJV,
		ToSystem:   VersificationMT,
		Mappings: []*RefMapping{
			{
				From: &Ref{Book: "Gen", Chapter: 1, Verse: 1},
				To:   nil, // No target - will be skipped in chain
				Type: MappingMissing,
			},
			{
				From: &Ref{Book: "Gen", Chapter: 1, Verse: 2},
				To:   &Ref{Book: "Gen", Chapter: 1, Verse: 2},
				Type: MappingExact,
			},
		},
	}

	second := &MappingTable{
		ID:         "second",
		FromSystem: VersificationMT,
		ToSystem:   VersificationLXX,
	}
	second.AddMapping(
		&Ref{Book: "Gen", Chapter: 1, Verse: 2},
		&Ref{Book: "Gen", Chapter: 1, Verse: 2},
		MappingExact,
	)

	registry.RegisterTable(first)
	registry.RegisterTable(second)

	// Get chained mapping
	chained := registry.GetChainedMapping(VersificationKJV, VersificationLXX)
	if chained == nil {
		t.Fatal("GetChainedMapping returned nil")
	}

	// The nil intermediate should be skipped, only the valid mapping should chain
	if len(chained.Mappings) != 1 {
		t.Errorf("Expected 1 chained mapping (nil skipped), got %d", len(chained.Mappings))
	}
}

// TestApplyToCorpusWithAnnotations tests ApplyToCorpus with annotations.
func TestApplyToCorpusWithAnnotations(t *testing.T) {
	corpus := &Corpus{
		ID:            "test-with-annotations",
		Versification: string(VersificationKJV),
		Documents: []*Document{
			{
				ID:           "ps10",
				Title:        "Psalm 10",
				CanonicalRef: &Ref{Book: "Ps", Chapter: 10, Verse: 1},
				ContentBlocks: []*ContentBlock{
					{ID: "block1", Sequence: 0, Text: "Why standest thou afar off?"},
				},
				Annotations: []*Annotation{
					{ID: "ann1", SpanID: "span1", Type: AnnotationStrongs, Value: "H3068"},
				},
			},
		},
	}

	table := &MappingTable{
		ID:         "kjv-to-lxx",
		FromSystem: VersificationKJV,
		ToSystem:   VersificationLXX,
	}

	mappedCorpus, _, err := table.ApplyToCorpus(corpus)
	if err != nil {
		t.Fatalf("ApplyToCorpus failed: %v", err)
	}

	// Verify annotations were copied
	doc := mappedCorpus.Documents[0]
	if len(doc.Annotations) != 1 {
		t.Errorf("Expected 1 annotation, got %d", len(doc.Annotations))
	}
}

// Test common versification differences
func TestPsalmVersificationDifferences(t *testing.T) {
	// Psalms often have different numbering between KJV and LXX
	// In LXX, Psalms 9-10 are combined into Psalm 9
	// Psalms 114-115 are combined into Psalm 113
	// etc.

	table := &MappingTable{
		ID:         "kjv-to-lxx-psalms",
		FromSystem: VersificationKJV,
		ToSystem:   VersificationLXX,
	}

	// KJV Psalm 10:1 -> LXX Psalm 9:22
	table.AddMapping(
		&Ref{Book: "Ps", Chapter: 10, Verse: 1},
		&Ref{Book: "Ps", Chapter: 9, Verse: 22},
		MappingExact,
	)

	// Verify mapping
	mapping := table.Lookup(&Ref{Book: "Ps", Chapter: 10, Verse: 1})
	if mapping == nil {
		t.Fatal("Psalm versification mapping not found")
	}
	if mapping.To.Chapter != 9 || mapping.To.Verse != 22 {
		t.Errorf("Expected LXX Ps.9.22, got Ps.%d.%d", mapping.To.Chapter, mapping.To.Verse)
	}
}
