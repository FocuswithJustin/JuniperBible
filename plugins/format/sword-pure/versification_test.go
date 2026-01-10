package main

import (
	"testing"
)

// TestVersificationSystems tests that all standard systems are defined.
func TestVersificationSystems(t *testing.T) {
	systems := []VersificationID{
		VersKJV,
		VersNRSV,
		VersNRSVA,
		VersVulgate,
		VersCatholic,
		VersLXX,
		VersMT,
		VersSynodal,
		VersGerman,
		VersLuther,
	}

	for _, sys := range systems {
		if sys == "" {
			t.Error("expected non-empty versification ID")
		}
	}
}

// TestNewVersification tests creating versification instances.
func TestNewVersification(t *testing.T) {
	v, err := NewVersification(VersKJV)
	if err != nil {
		t.Fatalf("failed to create KJV versification: %v", err)
	}

	if v.ID != VersKJV {
		t.Errorf("expected ID %q, got %q", VersKJV, v.ID)
	}

	// KJV should have 66 books
	if len(v.Books) != 66 {
		t.Errorf("expected 66 books, got %d", len(v.Books))
	}
}

// TestVersificationGetBookIndex tests book index lookup.
func TestVersificationGetBookIndex(t *testing.T) {
	v, err := NewVersification(VersKJV)
	if err != nil {
		t.Fatalf("failed to create versification: %v", err)
	}

	tests := []struct {
		book     string
		expected int
	}{
		{"Gen", 0},
		{"Exod", 1},
		{"Ps", 18},
		{"Matt", 39},
		{"Rev", 65},
		{"Unknown", -1},
	}

	for _, tt := range tests {
		got := v.GetBookIndex(tt.book)
		if got != tt.expected {
			t.Errorf("GetBookIndex(%q): expected %d, got %d", tt.book, tt.expected, got)
		}
	}
}

// TestVersificationGetChapterCount tests chapter count lookup.
func TestVersificationGetChapterCount(t *testing.T) {
	v, err := NewVersification(VersKJV)
	if err != nil {
		t.Fatalf("failed to create versification: %v", err)
	}

	tests := []struct {
		book     string
		expected int
	}{
		{"Gen", 50},
		{"Ps", 150},
		{"Matt", 28},
		{"Jude", 1},
		{"Rev", 22},
	}

	for _, tt := range tests {
		got := v.GetChapterCount(tt.book)
		if got != tt.expected {
			t.Errorf("GetChapterCount(%q): expected %d, got %d", tt.book, tt.expected, got)
		}
	}
}

// TestVersificationGetVerseCount tests verse count lookup.
func TestVersificationGetVerseCount(t *testing.T) {
	v, err := NewVersification(VersKJV)
	if err != nil {
		t.Fatalf("failed to create versification: %v", err)
	}

	tests := []struct {
		book     string
		chapter  int
		expected int
	}{
		{"Gen", 1, 31},
		{"Gen", 2, 25},
		{"Ps", 23, 6},
		{"Ps", 119, 176},
		{"Matt", 1, 25},
		{"John", 3, 36},
	}

	for _, tt := range tests {
		got := v.GetVerseCount(tt.book, tt.chapter)
		if got != tt.expected {
			t.Errorf("GetVerseCount(%q, %d): expected %d, got %d", tt.book, tt.chapter, tt.expected, got)
		}
	}
}

// TestVersificationCalculateIndex tests calculating verse indices.
// SWORD uses a complex indexing scheme with headers for module, books, and chapters.
func TestVersificationCalculateIndex(t *testing.T) {
	v, err := NewVersification(VersKJV)
	if err != nil {
		t.Fatalf("failed to create versification: %v", err)
	}

	// The SWORD index structure for Genesis is:
	// [0] = empty
	// [1] = module header
	// [2] = Genesis book intro
	// [3] = Genesis chapter 1 heading
	// [4] = Genesis 1:1
	// [5] = Genesis 1:2
	// ...
	// [34] = Genesis 1:31
	// [35] = Genesis chapter 2 heading
	// [36] = Genesis 2:1

	tests := []struct {
		ref      *Ref
		expected int
	}{
		{&Ref{Book: "Gen", Chapter: 1, Verse: 1}, 4},  // Module(1) + BookIntro(1) + ChapterHeading(1) + Verse(1) = 4
		{&Ref{Book: "Gen", Chapter: 1, Verse: 2}, 5},  // 4 + 1 = 5
		{&Ref{Book: "Gen", Chapter: 2, Verse: 1}, 36}, // 4 + 31 (Gen 1 verses) + 1 (ch2 heading) = 36
	}

	for _, tt := range tests {
		got, err := v.CalculateIndex(tt.ref, false)
		if err != nil {
			t.Errorf("CalculateIndex(%v) error: %v", tt.ref, err)
			continue
		}
		if got != tt.expected {
			t.Errorf("CalculateIndex(%v): expected %d, got %d", tt.ref, tt.expected, got)
		}
	}
}

// TestVersificationCalculateNTIndex tests NT verse index calculation.
// SWORD NT uses the same header scheme: module(1) + book(1) + chapter(1) + verse.
func TestVersificationCalculateNTIndex(t *testing.T) {
	v, err := NewVersification(VersKJV)
	if err != nil {
		t.Fatalf("failed to create versification: %v", err)
	}

	// NT structure similar to OT:
	// [0] = empty
	// [1] = module header
	// [2] = Matthew book intro
	// [3] = Matthew chapter 1 heading
	// [4] = Matt 1:1
	tests := []struct {
		ref      *Ref
		expected int
	}{
		{&Ref{Book: "Matt", Chapter: 1, Verse: 1}, 4}, // Module(1) + BookIntro(1) + ChapterHeading(1) + Verse(1) = 4
		{&Ref{Book: "Matt", Chapter: 1, Verse: 2}, 5},
		{&Ref{Book: "Matt", Chapter: 2, Verse: 1}, 30}, // 4 + 25 (Matt 1 verses) + 1 (ch2 heading) = 30
	}

	for _, tt := range tests {
		got, err := v.CalculateIndex(tt.ref, true)
		if err != nil {
			t.Errorf("CalculateIndex(%v, NT) error: %v", tt.ref, err)
			continue
		}
		if got != tt.expected {
			t.Errorf("CalculateIndex(%v, NT): expected %d, got %d", tt.ref, tt.expected, got)
		}
	}
}

// TestVersificationIndexToRef tests converting indices back to references.
// Uses the SWORD header scheme: module(1) + book(1) + chapter(1) + verse.
func TestVersificationIndexToRef(t *testing.T) {
	v, err := NewVersification(VersKJV)
	if err != nil {
		t.Fatalf("failed to create versification: %v", err)
	}

	tests := []struct {
		index    int
		isNT     bool
		expected *Ref
	}{
		{4, false, &Ref{Book: "Gen", Chapter: 1, Verse: 1}},  // Index 4 = Gen 1:1
		{36, false, &Ref{Book: "Gen", Chapter: 2, Verse: 1}}, // Index 36 = Gen 2:1
		{4, true, &Ref{Book: "Matt", Chapter: 1, Verse: 1}},  // Index 4 = Matt 1:1
		{30, true, &Ref{Book: "Matt", Chapter: 2, Verse: 1}}, // Index 30 = Matt 2:1
	}

	for _, tt := range tests {
		got, err := v.IndexToRef(tt.index, tt.isNT)
		if err != nil {
			t.Errorf("IndexToRef(%d, %v) error: %v", tt.index, tt.isNT, err)
			continue
		}
		if got.Book != tt.expected.Book || got.Chapter != tt.expected.Chapter || got.Verse != tt.expected.Verse {
			t.Errorf("IndexToRef(%d, %v): expected %v, got %v", tt.index, tt.isNT, tt.expected, got)
		}
	}
}

// TestVersificationRoundTrip tests that index/ref conversions are consistent.
func TestVersificationRoundTrip(t *testing.T) {
	v, err := NewVersification(VersKJV)
	if err != nil {
		t.Fatalf("failed to create versification: %v", err)
	}

	refs := []*Ref{
		{Book: "Gen", Chapter: 1, Verse: 1},
		{Book: "Ps", Chapter: 23, Verse: 1},
		{Book: "Isa", Chapter: 53, Verse: 5},
		{Book: "Matt", Chapter: 5, Verse: 3},
		{Book: "John", Chapter: 3, Verse: 16},
		{Book: "Rev", Chapter: 22, Verse: 21},
	}

	for _, ref := range refs {
		isNT := BookIndex(ref.Book) >= 39

		idx, err := v.CalculateIndex(ref, isNT)
		if err != nil {
			t.Errorf("CalculateIndex(%v) error: %v", ref, err)
			continue
		}

		backRef, err := v.IndexToRef(idx, isNT)
		if err != nil {
			t.Errorf("IndexToRef(%d) error: %v", idx, err)
			continue
		}

		if backRef.Book != ref.Book || backRef.Chapter != ref.Chapter || backRef.Verse != ref.Verse {
			t.Errorf("round-trip failed: %v -> %d -> %v", ref, idx, backRef)
		}
	}
}

// TestVersificationTotalVerses tests total verse counts.
func TestVersificationTotalVerses(t *testing.T) {
	v, err := NewVersification(VersKJV)
	if err != nil {
		t.Fatalf("failed to create versification: %v", err)
	}

	tests := []struct {
		book     string
		expected int
	}{
		{"Gen", 1533},
		{"Ps", 2461},
		{"Matt", 1071},
		{"John", 879},
		{"Rev", 404},
	}

	for _, tt := range tests {
		got := v.GetTotalVerses(tt.book)
		if got != tt.expected {
			t.Errorf("GetTotalVerses(%q): expected %d, got %d", tt.book, tt.expected, got)
		}
	}
}

// TestVersificationFromConf tests getting versification from conf file.
func TestVersificationFromConf(t *testing.T) {
	conf := &ConfFile{
		Versification: "KJV",
	}

	v, err := VersificationFromConf(conf)
	if err != nil {
		t.Fatalf("VersificationFromConf failed: %v", err)
	}

	if v.ID != VersKJV {
		t.Errorf("expected VersKJV, got %v", v.ID)
	}
}

// TestVersificationFromConfDefault tests default versification.
func TestVersificationFromConfDefault(t *testing.T) {
	conf := &ConfFile{} // No versification specified

	v, err := VersificationFromConf(conf)
	if err != nil {
		t.Fatalf("VersificationFromConf failed: %v", err)
	}

	// Should default to KJV
	if v.ID != VersKJV {
		t.Errorf("expected default VersKJV, got %v", v.ID)
	}
}
