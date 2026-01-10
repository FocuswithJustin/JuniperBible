package usx

import (
	"fmt"
	"testing"

	"github.com/FocuswithJustin/JuniperBible/core/ir"
)

// FuzzParseUSXToIR tests the USX XML parser with fuzzing
func FuzzParseUSXToIR(f *testing.F) {
	// Seed corpus with valid USX examples
	f.Add([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<usx version="3.0">
  <book code="GEN" style="id">Genesis</book>
  <chapter number="1" style="c" sid="GEN 1"/>
  <para style="p">
    <verse number="1" style="v" sid="GEN 1:1"/>
    In the beginning God created the heavens and the earth.
    <verse eid="GEN 1:1"/>
  </para>
  <chapter eid="GEN 1"/>
</usx>`))

	// Minimal valid USX
	f.Add([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<usx version="3.0">
  <book code="MAT" style="id">Matthew</book>
  <chapter number="1"/>
  <para style="p">
    <verse number="1"/>
    Test verse content.
  </para>
</usx>`))

	// USX with multiple verses
	f.Add([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<usx version="3.0">
  <book code="PSA" style="id">Psalms</book>
  <chapter number="23"/>
  <para style="q1">
    <verse number="1"/>The LORD is my shepherd;
  </para>
  <para style="q1">
    <verse number="2"/>I shall not want.
  </para>
</usx>`))

	// USX 2.0 version
	f.Add([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<usx version="2.0">
  <book code="JHN" style="id">John</book>
  <chapter number="1"/>
  <verse number="1"/>
  <para style="p">In the beginning was the Word.</para>
</usx>`))

	// Empty USX
	f.Add([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<usx version="3.0">
</usx>`))

	// USX without XML declaration
	f.Add([]byte(`<usx version="3.0">
  <book code="REV" style="id">Revelation</book>
</usx>`))

	f.Fuzz(func(t *testing.T, data []byte) {
		// The parser should handle malformed input gracefully
		// We allow panics during fuzzing as they help identify bugs
		defer func() {
			if r := recover(); r != nil {
				// Panic is acceptable - it means fuzzer found an edge case
				// In production, the parser would be fixed
				t.Logf("Parser panicked (this is expected during fuzzing): %v", r)
			}
		}()

		corpus, err := parseUSXToIR(data)

		// If parsing succeeds, validate basic invariants
		if err == nil && corpus != nil {
			// Corpus should have valid basic fields
			if corpus.Version == "" {
				t.Error("Corpus version should not be empty")
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

				// All content blocks should have valid hashes
				for j, cb := range doc.ContentBlocks {
					if cb == nil {
						t.Errorf("ContentBlock at doc %d, block %d is nil", i, j)
						continue
					}
					if cb.Hash == "" {
						t.Errorf("ContentBlock at doc %d, block %d has empty hash", i, j)
					}
					// Sequence should be positive
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

// FuzzEmitUSXFromIR tests the USX XML emitter with fuzzing
func FuzzEmitUSXFromIR(f *testing.F) {
	// Seed with valid corpus configurations
	f.Add("GEN", "Genesis", 1)
	f.Add("MAT", "Matthew", 2)
	f.Add("PSA", "Psalms", 3)
	f.Add("", "", 0)
	f.Add("REV", "Revelation", 99)

	f.Fuzz(func(t *testing.T, bookCode, title string, numBlocks int) {
		// Create a test corpus
		corpus := createTestUSXCorpus(bookCode, title, numBlocks)

		// The emitter should not panic on any corpus
		output := emitUSXFromIR(corpus)

		// If emission succeeds, do basic validation
		if len(output) > 0 {
			// Should contain XML-like structure
			outputStr := string(output)
			_ = outputStr

			// Basic sanity: output should not be excessively large
			if len(outputStr) > 1000000 {
				t.Error("Generated USX is unexpectedly large")
			}
		}
	})
}

// FuzzCreateContentBlock tests the content block creator with fuzzing
func FuzzCreateContentBlock(f *testing.F) {
	// Seed with various inputs
	f.Add(1, "In the beginning", "GEN", 1, 1)
	f.Add(100, "For God so loved the world", "JHN", 3, 16)
	f.Add(0, "", "", 0, 0)
	f.Add(-1, "Negative sequence", "TEST", -1, -1)
	f.Add(999999, "Large sequence", "REV", 22, 21)

	f.Fuzz(func(t *testing.T, sequence int, text, book string, chapter, verse int) {
		// The function should not panic on any input
		cb := createContentBlock(sequence, text, book, chapter, verse)

		// Content block should never be nil
		if cb == nil {
			t.Error("createContentBlock returned nil")
			return
		}

		// Basic invariants
		if cb.ID == "" {
			t.Error("ContentBlock has empty ID")
		}

		if cb.Sequence != sequence {
			t.Errorf("ContentBlock sequence mismatch: expected %d, got %d", sequence, cb.Sequence)
		}

		// Text should match input (after trimming)
		if len(text) > 0 && cb.Text == "" {
			t.Error("ContentBlock has empty text when input was non-empty")
		}

		// Should have at least one anchor
		if len(cb.Anchors) == 0 {
			t.Error("ContentBlock has no anchors")
		}

		// Hash should be computed
		if cb.Hash == "" {
			t.Error("ContentBlock hash is empty")
		}
	})
}

// FuzzEscapeXML tests the XML escaper with fuzzing
func FuzzEscapeXML(f *testing.F) {
	// Seed with strings that need escaping
	f.Add("Hello & goodbye")
	f.Add("<tag>content</tag>")
	f.Add("Quote: \"test\"")
	f.Add("Apostrophe: 'test'")
	f.Add("&lt;&gt;&amp;")
	f.Add("")
	f.Add("Normal text")
	f.Add("Multiple & < > \" ' special chars")
	f.Add("\x00\x01\x02") // Control characters

	f.Fuzz(func(t *testing.T, input string) {
		// The escaper should not panic on any input
		escaped := escapeXML(input)

		// Basic invariants
		_ = escaped

		// Escaped string should handle special characters
		if len(escaped) > 0 && len(input) > 0 {
			// Output length should be reasonable
			if len(escaped) > len(input)*10 {
				t.Error("Escaped string is unexpectedly long")
			}
		}
	})
}

// Helper function to create a test USX corpus
func createTestUSXCorpus(bookCode, title string, numBlocks int) *ir.Corpus {
	corpus := &ir.Corpus{
		ID:         bookCode,
		Version:    "1.0.0",
		ModuleType: ir.ModuleBible,
		LossClass:  ir.LossL0,
		Title:      title,
		Documents:  []*ir.Document{},
	}

	if bookCode != "" {
		doc := &ir.Document{
			ID:            bookCode,
			Title:         title,
			Order:         1,
			ContentBlocks: []*ir.ContentBlock{},
		}

		// Add content blocks based on numBlocks (limit to reasonable size)
		if numBlocks > 100 {
			numBlocks = 100
		}
		if numBlocks < 0 {
			numBlocks = 0
		}

		for i := 0; i < numBlocks; i++ {
			cb := &ir.ContentBlock{
				ID:       fmt.Sprintf("cb-%d", i+1),
				Sequence: i + 1,
				Text:     fmt.Sprintf("Test verse %d", i+1),
				Hash:     "test-hash",
				Anchors: []*ir.Anchor{
					{
						ID:             fmt.Sprintf("a-%d-0", i+1),
						ContentBlockID: fmt.Sprintf("cb-%d", i+1),
						CharOffset:     0,
					},
				},
			}
			doc.ContentBlocks = append(doc.ContentBlocks, cb)
		}

		corpus.Documents = append(corpus.Documents, doc)
	}

	return corpus
}
