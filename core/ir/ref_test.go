package ir

import (
	"encoding/json"
	"testing"
)

func TestRefJSON(t *testing.T) {
	ref := &Ref{
		Book:     "Gen",
		Chapter:  1,
		Verse:    1,
		VerseEnd: 3,
		SubVerse: "a",
		OSISID:   "Gen.1.1a-3",
	}

	// Marshal to JSON
	data, err := json.Marshal(ref)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	// Unmarshal back
	var decoded Ref
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	// Verify fields
	if decoded.Book != ref.Book {
		t.Errorf("Book = %q, want %q", decoded.Book, ref.Book)
	}
	if decoded.Chapter != ref.Chapter {
		t.Errorf("Chapter = %d, want %d", decoded.Chapter, ref.Chapter)
	}
	if decoded.Verse != ref.Verse {
		t.Errorf("Verse = %d, want %d", decoded.Verse, ref.Verse)
	}
	if decoded.VerseEnd != ref.VerseEnd {
		t.Errorf("VerseEnd = %d, want %d", decoded.VerseEnd, ref.VerseEnd)
	}
	if decoded.SubVerse != ref.SubVerse {
		t.Errorf("SubVerse = %q, want %q", decoded.SubVerse, ref.SubVerse)
	}
	if decoded.OSISID != ref.OSISID {
		t.Errorf("OSISID = %q, want %q", decoded.OSISID, ref.OSISID)
	}
}

func TestParseRef(t *testing.T) {
	tests := []struct {
		input    string
		expected *Ref
		wantErr  bool
	}{
		// Book only
		{
			input: "Gen",
			expected: &Ref{
				Book:   "Gen",
				OSISID: "Gen",
			},
		},
		// Book and chapter
		{
			input: "Gen.1",
			expected: &Ref{
				Book:    "Gen",
				Chapter: 1,
				OSISID:  "Gen.1",
			},
		},
		// Book, chapter, and verse
		{
			input: "Gen.1.1",
			expected: &Ref{
				Book:    "Gen",
				Chapter: 1,
				Verse:   1,
				OSISID:  "Gen.1.1",
			},
		},
		// With sub-verse
		{
			input: "Gen.1.1a",
			expected: &Ref{
				Book:     "Gen",
				Chapter:  1,
				Verse:    1,
				SubVerse: "a",
				OSISID:   "Gen.1.1a",
			},
		},
		// Verse range
		{
			input: "Matt.5.3-12",
			expected: &Ref{
				Book:     "Matt",
				Chapter:  5,
				Verse:    3,
				VerseEnd: 12,
				OSISID:   "Matt.5.3-12",
			},
		},
		// Books with numbers
		{
			input: "1John.3.16",
			expected: &Ref{
				Book:    "1John",
				Chapter: 3,
				Verse:   16,
				OSISID:  "1John.3.16",
			},
		},
		{
			input: "2Cor.5.21",
			expected: &Ref{
				Book:    "2Cor",
				Chapter: 5,
				Verse:   21,
				OSISID:  "2Cor.5.21",
			},
		},
		// Error cases
		{
			input:   "",
			wantErr: true,
		},
		{
			input:   "123",
			wantErr: true,
		},
		{
			input:   "Gen.abc",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		ref, err := ParseRef(tt.input)
		if tt.wantErr {
			if err == nil {
				t.Errorf("ParseRef(%q) expected error", tt.input)
			}
			continue
		}

		if err != nil {
			t.Errorf("ParseRef(%q) error: %v", tt.input, err)
			continue
		}

		if ref.Book != tt.expected.Book {
			t.Errorf("ParseRef(%q).Book = %q, want %q", tt.input, ref.Book, tt.expected.Book)
		}
		if ref.Chapter != tt.expected.Chapter {
			t.Errorf("ParseRef(%q).Chapter = %d, want %d", tt.input, ref.Chapter, tt.expected.Chapter)
		}
		if ref.Verse != tt.expected.Verse {
			t.Errorf("ParseRef(%q).Verse = %d, want %d", tt.input, ref.Verse, tt.expected.Verse)
		}
		if ref.VerseEnd != tt.expected.VerseEnd {
			t.Errorf("ParseRef(%q).VerseEnd = %d, want %d", tt.input, ref.VerseEnd, tt.expected.VerseEnd)
		}
		if ref.SubVerse != tt.expected.SubVerse {
			t.Errorf("ParseRef(%q).SubVerse = %q, want %q", tt.input, ref.SubVerse, tt.expected.SubVerse)
		}
	}
}

func TestRefString(t *testing.T) {
	tests := []struct {
		ref      *Ref
		expected string
	}{
		{
			ref:      &Ref{Book: "Gen", OSISID: "Gen"},
			expected: "Gen",
		},
		{
			ref:      &Ref{Book: "Gen", Chapter: 1, OSISID: "Gen.1"},
			expected: "Gen.1",
		},
		{
			ref:      &Ref{Book: "Gen", Chapter: 1, Verse: 1},
			expected: "Gen.1.1",
		},
		{
			ref:      &Ref{Book: "Gen", Chapter: 1, Verse: 1, SubVerse: "a"},
			expected: "Gen.1.1a",
		},
		{
			ref:      &Ref{Book: "Matt", Chapter: 5, Verse: 3, VerseEnd: 12},
			expected: "Matt.5.3-12",
		},
	}

	for _, tt := range tests {
		// Clear OSISID to test String() generation
		if tt.ref.OSISID != tt.expected {
			tt.ref.OSISID = ""
		}
		got := tt.ref.String()
		if got != tt.expected {
			t.Errorf("Ref.String() = %q, want %q", got, tt.expected)
		}
	}
}

func TestRefIsRange(t *testing.T) {
	tests := []struct {
		ref     *Ref
		isRange bool
	}{
		{&Ref{Book: "Gen", Chapter: 1, Verse: 1}, false},
		{&Ref{Book: "Gen", Chapter: 1, Verse: 1, VerseEnd: 1}, false},
		{&Ref{Book: "Gen", Chapter: 1, Verse: 1, VerseEnd: 3}, true},
		{&Ref{Book: "Matt", Chapter: 5, Verse: 3, VerseEnd: 12}, true},
	}

	for _, tt := range tests {
		if got := tt.ref.IsRange(); got != tt.isRange {
			t.Errorf("Ref{%s}.IsRange() = %v, want %v", tt.ref.String(), got, tt.isRange)
		}
	}
}

func TestRefContains(t *testing.T) {
	tests := []struct {
		ref      *Ref
		other    *Ref
		contains bool
	}{
		// Book contains all chapters
		{
			ref:      &Ref{Book: "Gen"},
			other:    &Ref{Book: "Gen", Chapter: 1, Verse: 1},
			contains: true,
		},
		// Different book
		{
			ref:      &Ref{Book: "Gen"},
			other:    &Ref{Book: "Exod", Chapter: 1, Verse: 1},
			contains: false,
		},
		// Chapter contains all verses
		{
			ref:      &Ref{Book: "Gen", Chapter: 1},
			other:    &Ref{Book: "Gen", Chapter: 1, Verse: 5},
			contains: true,
		},
		// Different chapter
		{
			ref:      &Ref{Book: "Gen", Chapter: 1},
			other:    &Ref{Book: "Gen", Chapter: 2, Verse: 1},
			contains: false,
		},
		// Exact verse match
		{
			ref:      &Ref{Book: "Gen", Chapter: 1, Verse: 1},
			other:    &Ref{Book: "Gen", Chapter: 1, Verse: 1},
			contains: true,
		},
		// Verse range contains verse
		{
			ref:      &Ref{Book: "Matt", Chapter: 5, Verse: 3, VerseEnd: 12},
			other:    &Ref{Book: "Matt", Chapter: 5, Verse: 5},
			contains: true,
		},
		// Verse range doesn't contain verse
		{
			ref:      &Ref{Book: "Matt", Chapter: 5, Verse: 3, VerseEnd: 12},
			other:    &Ref{Book: "Matt", Chapter: 5, Verse: 15},
			contains: false,
		},
	}

	for _, tt := range tests {
		if got := tt.ref.Contains(tt.other); got != tt.contains {
			t.Errorf("Ref{%s}.Contains(Ref{%s}) = %v, want %v",
				tt.ref.String(), tt.other.String(), got, tt.contains)
		}
	}
}

func TestRefSetJSON(t *testing.T) {
	rs := &RefSet{
		ID:    "xref1",
		Label: "Creation parallels",
		Refs: []*Ref{
			{Book: "Gen", Chapter: 1, Verse: 1, OSISID: "Gen.1.1"},
			{Book: "John", Chapter: 1, Verse: 1, OSISID: "John.1.1"},
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(rs)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	// Unmarshal back
	var decoded RefSet
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	// Verify fields
	if decoded.ID != rs.ID {
		t.Errorf("ID = %q, want %q", decoded.ID, rs.ID)
	}
	if decoded.Label != rs.Label {
		t.Errorf("Label = %q, want %q", decoded.Label, rs.Label)
	}
	if len(decoded.Refs) != 2 {
		t.Fatalf("len(Refs) = %d, want 2", len(decoded.Refs))
	}
}

func TestRefSetAdd(t *testing.T) {
	rs := &RefSet{ID: "xref1"}

	rs.Add(&Ref{Book: "Gen", Chapter: 1, Verse: 1})
	rs.Add(&Ref{Book: "John", Chapter: 1, Verse: 1})

	if len(rs.Refs) != 2 {
		t.Errorf("len(Refs) = %d, want 2", len(rs.Refs))
	}
}

func TestRefRangeContains(t *testing.T) {
	rr := &RefRange{
		Start: &Ref{Book: "Gen", Chapter: 1, Verse: 5},
		End:   &Ref{Book: "Gen", Chapter: 2, Verse: 5},
	}

	tests := []struct {
		ref      *Ref
		contains bool
	}{
		// Before start verse in start chapter
		{&Ref{Book: "Gen", Chapter: 1, Verse: 1}, false},
		{&Ref{Book: "Gen", Chapter: 1, Verse: 4}, false},
		// At start
		{&Ref{Book: "Gen", Chapter: 1, Verse: 5}, true},
		{&Ref{Book: "Gen", Chapter: 1, Verse: 31}, true},
		{&Ref{Book: "Gen", Chapter: 2, Verse: 1}, true},
		{&Ref{Book: "Gen", Chapter: 2, Verse: 5}, true},
		{&Ref{Book: "Gen", Chapter: 2, Verse: 10}, false},
		{&Ref{Book: "Gen", Chapter: 3, Verse: 1}, false},
		{&Ref{Book: "Exod", Chapter: 1, Verse: 1}, false},
	}

	for _, tt := range tests {
		if got := rr.Contains(tt.ref); got != tt.contains {
			t.Errorf("RefRange.Contains(Ref{%s}) = %v, want %v",
				tt.ref.String(), got, tt.contains)
		}
	}
}

func TestParseRefRoundTrip(t *testing.T) {
	inputs := []string{
		"Gen",
		"Gen.1",
		"Gen.1.1",
		"Gen.1.1a",
		"Matt.5.3-12",
		"1John.3.16",
		"2Cor.5.21",
	}

	for _, input := range inputs {
		ref, err := ParseRef(input)
		if err != nil {
			t.Errorf("ParseRef(%q) error: %v", input, err)
			continue
		}

		output := ref.String()
		if output != input {
			t.Errorf("ParseRef(%q).String() = %q, want %q", input, output, input)
		}
	}
}
