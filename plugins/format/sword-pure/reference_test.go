package main

import (
	"testing"
)

// TestParseRefOSIS tests parsing OSIS-style references.
func TestParseRefOSIS(t *testing.T) {
	tests := []struct {
		input    string
		expected *Ref
	}{
		{
			input:    "Gen.1.1",
			expected: &Ref{Book: "Gen", Chapter: 1, Verse: 1},
		},
		{
			input:    "Matt.5.3",
			expected: &Ref{Book: "Matt", Chapter: 5, Verse: 3},
		},
		{
			input:    "Rev.22.21",
			expected: &Ref{Book: "Rev", Chapter: 22, Verse: 21},
		},
		{
			input:    "Gen.1.1-5",
			expected: &Ref{Book: "Gen", Chapter: 1, Verse: 1, VerseEnd: 5},
		},
		{
			input:    "Gen.1.1-2.3",
			expected: &Ref{Book: "Gen", Chapter: 1, Verse: 1, ChapterEnd: 2, VerseEnd: 3},
		},
	}

	for _, tt := range tests {
		ref, err := ParseRef(tt.input)
		if err != nil {
			t.Errorf("ParseRef(%q) failed: %v", tt.input, err)
			continue
		}

		if ref.Book != tt.expected.Book {
			t.Errorf("ParseRef(%q).Book: expected %q, got %q", tt.input, tt.expected.Book, ref.Book)
		}
		if ref.Chapter != tt.expected.Chapter {
			t.Errorf("ParseRef(%q).Chapter: expected %d, got %d", tt.input, tt.expected.Chapter, ref.Chapter)
		}
		if ref.Verse != tt.expected.Verse {
			t.Errorf("ParseRef(%q).Verse: expected %d, got %d", tt.input, tt.expected.Verse, ref.Verse)
		}
		if ref.VerseEnd != tt.expected.VerseEnd {
			t.Errorf("ParseRef(%q).VerseEnd: expected %d, got %d", tt.input, tt.expected.VerseEnd, ref.VerseEnd)
		}
		if ref.ChapterEnd != tt.expected.ChapterEnd {
			t.Errorf("ParseRef(%q).ChapterEnd: expected %d, got %d", tt.input, tt.expected.ChapterEnd, ref.ChapterEnd)
		}
	}
}

// TestParseRefHuman tests parsing human-readable references.
func TestParseRefHuman(t *testing.T) {
	tests := []struct {
		input    string
		expected *Ref
	}{
		{
			input:    "Gen 1:1",
			expected: &Ref{Book: "Gen", Chapter: 1, Verse: 1},
		},
		{
			input:    "Genesis 1:1",
			expected: &Ref{Book: "Gen", Chapter: 1, Verse: 1},
		},
		{
			input:    "Matt 5:3",
			expected: &Ref{Book: "Matt", Chapter: 5, Verse: 3},
		},
		{
			input:    "Matthew 5:3",
			expected: &Ref{Book: "Matt", Chapter: 5, Verse: 3},
		},
		{
			input:    "1 John 3:16",
			expected: &Ref{Book: "1John", Chapter: 3, Verse: 16},
		},
		{
			input:    "Rev 22:21",
			expected: &Ref{Book: "Rev", Chapter: 22, Verse: 21},
		},
		{
			input:    "Revelation 22:21",
			expected: &Ref{Book: "Rev", Chapter: 22, Verse: 21},
		},
		{
			input:    "Gen 1:1-5",
			expected: &Ref{Book: "Gen", Chapter: 1, Verse: 1, VerseEnd: 5},
		},
		{
			input:    "Ps 23:1",
			expected: &Ref{Book: "Ps", Chapter: 23, Verse: 1},
		},
		{
			input:    "Psalm 23:1",
			expected: &Ref{Book: "Ps", Chapter: 23, Verse: 1},
		},
	}

	for _, tt := range tests {
		ref, err := ParseRef(tt.input)
		if err != nil {
			t.Errorf("ParseRef(%q) failed: %v", tt.input, err)
			continue
		}

		if ref.Book != tt.expected.Book {
			t.Errorf("ParseRef(%q).Book: expected %q, got %q", tt.input, tt.expected.Book, ref.Book)
		}
		if ref.Chapter != tt.expected.Chapter {
			t.Errorf("ParseRef(%q).Chapter: expected %d, got %d", tt.input, tt.expected.Chapter, ref.Chapter)
		}
		if ref.Verse != tt.expected.Verse {
			t.Errorf("ParseRef(%q).Verse: expected %d, got %d", tt.input, tt.expected.Verse, ref.Verse)
		}
		if ref.VerseEnd != tt.expected.VerseEnd {
			t.Errorf("ParseRef(%q).VerseEnd: expected %d, got %d", tt.input, tt.expected.VerseEnd, ref.VerseEnd)
		}
	}
}

// TestRefString tests the String() method.
func TestRefString(t *testing.T) {
	tests := []struct {
		ref      *Ref
		expected string
	}{
		{
			ref:      &Ref{Book: "Gen", Chapter: 1, Verse: 1},
			expected: "Gen.1.1",
		},
		{
			ref:      &Ref{Book: "Gen", Chapter: 1, Verse: 1, VerseEnd: 5},
			expected: "Gen.1.1-5",
		},
		{
			ref:      &Ref{Book: "Gen", Chapter: 1, Verse: 1, ChapterEnd: 2, VerseEnd: 3},
			expected: "Gen.1.1-2.3",
		},
	}

	for _, tt := range tests {
		got := tt.ref.String()
		if got != tt.expected {
			t.Errorf("Ref.String(): expected %q, got %q", tt.expected, got)
		}
	}
}

// TestNormalizeBookName tests book name normalization.
func TestNormalizeBookName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Gen", "Gen"},
		{"gen", "Gen"},
		{"Genesis", "Gen"},
		{"genesis", "Gen"},
		{"Matt", "Matt"},
		{"matthew", "Matt"},
		{"Mt", "Matt"},
		{"1 John", "1John"},
		{"1john", "1John"},
		{"Ps", "Ps"},
		{"Psalm", "Ps"},
		{"Psalms", "Ps"},
		{"Rev", "Rev"},
		{"Revelation", "Rev"},
		{"Apocalypse", "Rev"},
	}

	for _, tt := range tests {
		got := normalizeBookName(tt.input)
		if got != tt.expected {
			t.Errorf("normalizeBookName(%q): expected %q, got %q", tt.input, tt.expected, got)
		}
	}
}

// TestBookIndex tests book index lookup.
func TestBookIndex(t *testing.T) {
	tests := []struct {
		book     string
		expected int
	}{
		{"Gen", 0},
		{"Exod", 1},
		{"Rev", 65},
		{"Matt", 39},
		{"Unknown", -1},
	}

	for _, tt := range tests {
		got := BookIndex(tt.book)
		if got != tt.expected {
			t.Errorf("BookIndex(%q): expected %d, got %d", tt.book, tt.expected, got)
		}
	}
}

// TestParseRefInvalid tests parsing invalid references.
func TestParseRefInvalid(t *testing.T) {
	invalidRefs := []string{
		"",
		"NotABook 1:1",
		"Gen",
		"Gen 1",
		"1:1",
	}

	for _, input := range invalidRefs {
		_, err := ParseRef(input)
		if err == nil {
			t.Errorf("ParseRef(%q) should have returned an error", input)
		}
	}
}
