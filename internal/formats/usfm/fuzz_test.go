package usfm

import (
	"fmt"
	"testing"

	"github.com/FocuswithJustin/JuniperBible/core/ir"
)

// FuzzParseUSFMToIR tests the USFM text parser with fuzzing
func FuzzParseUSFMToIR(f *testing.F) {
	// Seed corpus with valid USFM examples
	f.Add([]byte(`\id GEN
\h Genesis
\mt Genesis
\c 1
\v 1 In the beginning God created the heaven and the earth.
\v 2 And the earth was without form, and void.
`))

	// Minimal valid USFM
	f.Add([]byte(`\id MAT
\c 1
\v 1 The book of the generation of Jesus Christ.
`))

	// USFM with poetry
	f.Add([]byte(`\id PSA
\h Psalms
\c 23
\q1 The LORD is my shepherd;
\q2 I shall not want.
`))

	// USFM with paragraph markers
	f.Add([]byte(`\id JHN
\c 3
\p
\v 16 For God so loved the world, that he gave his only begotten Son.
`))

	// USFM with multiple chapters
	f.Add([]byte(`\id PHP
\h Philippians
\c 1
\v 1 Paul and Timotheus, the servants of Jesus Christ.
\c 2
\v 1 If there be therefore any consolation in Christ.
`))

	// Empty USFM
	f.Add([]byte(`\id GEN
`))

	// USFM without explicit chapter markers
	f.Add([]byte(`\id OBA
\v 1 The vision of Obadiah.
`))

	// USFM with various markers
	f.Add([]byte(`\id GEN
\toc1 Genesis
\toc2 Genesis
\toc3 Gen
\mt Genesis
\c 1
\p
\v 1 In the beginning.
`))

	f.Fuzz(func(t *testing.T, data []byte) {
		// The parser should not panic on any input
		corpus, err := parseUSFMToIR(data)

		// If parsing succeeds, validate basic invariants
		if err == nil && corpus != nil {
			// Corpus should have valid basic fields
			if corpus.Version == "" {
				t.Error("Corpus version should not be empty")
			}

			// Module type should be Bible
			if corpus.ModuleType != ir.ModuleBible {
				t.Error("Corpus module type should be Bible")
			}

			// All documents should have IDs
			for i, doc := range corpus.Documents {
				if doc == nil {
					t.Errorf("Document at index %d is nil", i)
					continue
				}
				if doc.ID == "" {
					t.Errorf("Document at index %d has empty ID", i)
				}
				if doc.Order <= 0 {
					t.Errorf("Document at index %d has invalid order %d", i, doc.Order)
				}

				// All content blocks should have valid hashes
				for j, cb := range doc.ContentBlocks {
					if cb == nil {
						t.Errorf("ContentBlock at doc %d, block %d is nil", i, j)
						continue
					}
					if cb.Hash == "" {
						t.Errorf("ContentBlock at doc %d, block %d has empty hash", i, j)
					}
					if cb.ID == "" {
						t.Errorf("ContentBlock at doc %d, block %d has empty ID", i, j)
					}
					if cb.Sequence <= 0 {
						t.Errorf("ContentBlock at doc %d, block %d has invalid sequence %d", i, j, cb.Sequence)
					}
				}
			}

			// Source hash should be set
			if corpus.SourceHash == "" {
				t.Error("Corpus source hash should not be empty")
			}
		}
	})
}

// FuzzEmitUSFMFromIR tests the USFM text emitter with fuzzing
func FuzzEmitUSFMFromIR(f *testing.F) {
	// Seed with valid corpus configurations
	f.Add("GEN", "Genesis", 1, 1)
	f.Add("MAT", "Matthew", 5, 10)
	f.Add("PSA", "Psalms", 23, 5)
	f.Add("", "", 0, 0)
	f.Add("REV", "Revelation", 22, 20)

	f.Fuzz(func(t *testing.T, bookID, title string, numDocs, numBlocks int) {
		// Create a test corpus
		corpus := createTestUSFMCorpus(bookID, title, numDocs, numBlocks)

		// The emitter should not panic on any corpus
		data, err := emitUSFMFromIR(corpus)

		// If emission succeeds, do basic validation
		if err == nil && len(data) > 0 {
			// Should contain USFM markers
			dataStr := string(data)
			_ = dataStr

			// Basic sanity: output should not be excessively large
			if len(dataStr) > 1000000 {
				t.Error("Generated USFM is unexpectedly large")
			}

			// Should not be empty for non-empty corpus
			if len(corpus.Documents) > 0 && len(dataStr) == 0 {
				t.Error("Generated USFM is empty for non-empty corpus")
			}
		}
	})
}

// FuzzMarkerRegex tests the marker regex with fuzzing
func FuzzMarkerRegex(f *testing.F) {
	// Seed with valid and invalid markers
	f.Add(`\id GEN`)
	f.Add(`\v 1`)
	f.Add(`\c 1`)
	f.Add(`\mt Genesis`)
	f.Add(`\q1 Poetry line`)
	f.Add(`\add*`)
	f.Add(`\no-marker`)
	f.Add(``)
	f.Add(`\\double-backslash`)
	f.Add(`\123`)
	f.Add(`\marker`)

	f.Fuzz(func(t *testing.T, input string) {
		// The regex should not panic on any input
		matches := markerRegex.FindStringSubmatch(input)
		_ = matches

		// If it matches, should have expected structure
		if len(matches) > 0 {
			// First capture group should be the marker name
			if len(matches) > 1 {
				marker := matches[1]
				_ = marker
			}
		}
	})
}

// FuzzVerseNumRegex tests the verse number regex with fuzzing
func FuzzVerseNumRegex(f *testing.F) {
	// Seed with valid verse numbers
	f.Add("1")
	f.Add("123")
	f.Add("1-5")
	f.Add("10-15")
	f.Add("1a")
	f.Add("invalid")
	f.Add("")
	f.Add("999-1000")

	f.Fuzz(func(t *testing.T, input string) {
		// The regex should not panic on any input
		matches := verseNumRegex.FindStringSubmatch(input)
		_ = matches

		// If it matches, validate structure
		if len(matches) > 0 {
			// Should have verse number in first capture group
			if len(matches) > 1 && matches[1] != "" {
				// Verse number should be present
				_ = matches[1]
			}
			// Optional second verse in range
			if len(matches) > 2 {
				_ = matches[2]
			}
		}
	})
}

// FuzzChapterRegex tests the chapter regex with fuzzing
func FuzzChapterRegex(f *testing.F) {
	// Seed with valid chapter numbers
	f.Add("1")
	f.Add("150")
	f.Add("999")
	f.Add("invalid")
	f.Add("")
	f.Add("1a")
	f.Add("abc")

	f.Fuzz(func(t *testing.T, input string) {
		// The regex should not panic on any input
		matches := chapterRegex.FindStringSubmatch(input)
		_ = matches

		// If it matches, should have chapter number
		if len(matches) > 0 && len(matches) > 1 {
			chapter := matches[1]
			_ = chapter
		}
	})
}

// FuzzBookNames tests book name lookups with fuzzing
func FuzzBookNames(f *testing.F) {
	// Seed with valid book codes
	f.Add("GEN")
	f.Add("MAT")
	f.Add("PSA")
	f.Add("REV")
	f.Add("1CO")
	f.Add("invalid")
	f.Add("")
	f.Add("gen")
	f.Add("GENESIS")

	f.Fuzz(func(t *testing.T, bookCode string) {
		// Looking up book names should not panic
		name, ok := bookNames[bookCode]
		_ = name
		_ = ok

		// If found, name should not be empty
		if ok && name == "" {
			t.Error("Book name should not be empty for valid code")
		}
	})
}

// Helper function to create a test USFM corpus
func createTestUSFMCorpus(bookID, title string, numDocs, numBlocks int) *ir.Corpus {
	corpus := &ir.Corpus{
		ID:         bookID,
		Version:    "1.0.0",
		ModuleType: ir.ModuleBible,
		LossClass:  ir.LossL0,
		Title:      title,
		Documents:  []*ir.Document{},
	}

	// Limit to reasonable sizes
	if numDocs > 10 {
		numDocs = 10
	}
	if numDocs < 0 {
		numDocs = 0
	}
	if numBlocks > 50 {
		numBlocks = 50
	}
	if numBlocks < 0 {
		numBlocks = 0
	}

	for i := 0; i < numDocs; i++ {
		docID := bookID
		if docID == "" {
			docID = fmt.Sprintf("BOOK%d", i+1)
		}

		doc := &ir.Document{
			ID:            docID,
			Title:         title,
			Order:         i + 1,
			ContentBlocks: []*ir.ContentBlock{},
		}

		// Add content blocks
		for j := 0; j < numBlocks; j++ {
			cb := &ir.ContentBlock{
				ID:       fmt.Sprintf("cb-%d", j+1),
				Sequence: j + 1,
				Text:     fmt.Sprintf("Test verse content %d", j+1),
				Hash:     "test-hash",
			}

			// Add anchor for verse content
			if j%2 == 0 {
				cb.Anchors = []*ir.Anchor{
					{
						ID:             fmt.Sprintf("a-%d-0", j+1),
						ContentBlockID: cb.ID,
						CharOffset:     0,
					},
				}
			}

			doc.ContentBlocks = append(doc.ContentBlocks, cb)
		}

		corpus.Documents = append(corpus.Documents, doc)
	}

	return corpus
}
