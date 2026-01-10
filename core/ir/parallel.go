// Package ir provides parallel corpus types for multi-translation alignment.
package ir

import (
	"fmt"
)

// AlignmentLevel represents the granularity of alignment.
type AlignmentLevel string

// Alignment level constants.
const (
	// AlignBook aligns at the book level.
	AlignBook AlignmentLevel = "book"

	// AlignChapter aligns at the chapter level.
	AlignChapter AlignmentLevel = "chapter"

	// AlignVerse aligns at the verse level.
	AlignVerse AlignmentLevel = "verse"

	// AlignToken aligns at the token/word level.
	AlignToken AlignmentLevel = "token"
)

// CorpusRef is a lightweight reference to a corpus.
type CorpusRef struct {
	// ID is the corpus identifier.
	ID string `json:"id"`

	// Language is the BCP-47 language tag.
	Language string `json:"language"`

	// Title is the optional display title.
	Title string `json:"title,omitempty"`
}

// ParallelCorpus represents a collection of aligned translations.
type ParallelCorpus struct {
	// ID is the unique identifier for this parallel corpus.
	ID string `json:"id"`

	// Version is the schema version.
	Version string `json:"version"`

	// BaseCorpus is the primary corpus for alignment (optional).
	BaseCorpus *CorpusRef `json:"base_corpus,omitempty"`

	// Corpora contains references to all aligned corpora.
	Corpora []*CorpusRef `json:"corpora"`

	// Alignments contains the alignment data.
	Alignments []*Alignment `json:"alignments,omitempty"`

	// DefaultAlignment is the default alignment level.
	DefaultAlignment AlignmentLevel `json:"default_alignment"`

	// Metadata contains optional metadata.
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Alignment represents a set of aligned units at a particular level.
type Alignment struct {
	// ID is the unique identifier for this alignment set.
	ID string `json:"id"`

	// Level is the alignment granularity.
	Level AlignmentLevel `json:"level"`

	// Units contains the aligned text units.
	Units []*AlignedUnit `json:"units,omitempty"`
}

// AlignedUnit represents a single aligned text unit (verse, sentence, etc.).
type AlignedUnit struct {
	// ID is the unique identifier for this unit.
	ID string `json:"id"`

	// Ref is the reference (verse, chapter, etc.) for this unit.
	Ref *Ref `json:"ref,omitempty"`

	// Texts maps corpus ID to the text content.
	Texts map[string]string `json:"texts"`

	// Level is the alignment level for this unit.
	Level AlignmentLevel `json:"level"`

	// TokenAlignments contains word-level alignments (optional).
	TokenAlignments []*TokenAlignment `json:"token_alignments,omitempty"`
}

// TokenAlignment represents word-level alignment between languages.
type TokenAlignment struct {
	// ID is the unique identifier for this token alignment.
	ID string `json:"id"`

	// SourceTokens are the token IDs from the source corpus.
	SourceTokens []string `json:"source_tokens"`

	// TargetTokens are the token IDs from the target corpus.
	TargetTokens []string `json:"target_tokens"`

	// Confidence is the alignment confidence (0.0-1.0).
	Confidence float64 `json:"confidence,omitempty"`

	// AlignType describes the alignment type (e.g., "one-to-one", "one-to-many").
	AlignType string `json:"align_type,omitempty"`
}

// InterlinearLine represents a single interlinear display line.
type InterlinearLine struct {
	// Ref is the reference for this line.
	Ref *Ref `json:"ref"`

	// Layers maps layer names to their content.
	Layers map[string]*InterlinearLayer `json:"layers"`
}

// InterlinearLayer represents a single layer in an interlinear display.
type InterlinearLayer struct {
	// CorpusID is the corpus this layer comes from.
	CorpusID string `json:"corpus_id"`

	// Tokens are the tokens in this layer.
	Tokens []string `json:"tokens"`

	// Label is the display label for this layer.
	Label string `json:"label"`
}

// AlignByVerse creates a verse-aligned parallel corpus from multiple corpora.
func AlignByVerse(corpora []*Corpus) (*ParallelCorpus, error) {
	if len(corpora) == 0 {
		return nil, fmt.Errorf("no corpora provided")
	}

	// Create corpus references
	corpusRefs := make([]*CorpusRef, len(corpora))
	for i, c := range corpora {
		corpusRefs[i] = &CorpusRef{
			ID:       c.ID,
			Language: c.Language,
			Title:    c.Title,
		}
	}

	// Use first corpus as base
	baseRef := corpusRefs[0]

	pc := &ParallelCorpus{
		ID:               fmt.Sprintf("parallel-%s", baseRef.ID),
		Version:          "1.0.0",
		BaseCorpus:       baseRef,
		Corpora:          corpusRefs,
		DefaultAlignment: AlignVerse,
	}

	// Create alignment based on verse structure
	alignment := &Alignment{
		ID:    "verse-alignment",
		Level: AlignVerse,
	}

	// For now, just set up the structure
	// Full alignment would iterate through verses and match them
	pc.Alignments = append(pc.Alignments, alignment)

	return pc, nil
}

// GetAlignedVerses returns the aligned units for a given reference.
func (pc *ParallelCorpus) GetAlignedVerses(ref *Ref) []*AlignedUnit {
	var result []*AlignedUnit

	for _, alignment := range pc.Alignments {
		if alignment.Level != AlignVerse {
			continue
		}
		for _, unit := range alignment.Units {
			if unit.Ref != nil && unit.Ref.OSISID == ref.OSISID {
				result = append(result, unit)
			}
		}
	}

	return result
}

// AlignOptions configures token alignment behavior.
type AlignOptions struct {
	// UseStrongs enables Strong's number-based alignment.
	UseStrongs bool `json:"use_strongs"`

	// MinConfidence is the minimum confidence threshold.
	MinConfidence float64 `json:"min_confidence"`

	// AllowUnaligned allows tokens without alignment.
	AllowUnaligned bool `json:"allow_unaligned"`
}

// AlignTokens creates token-level alignments between two content blocks.
func AlignTokens(source, target *ContentBlock, opts *AlignOptions) ([]*TokenAlignment, error) {
	if opts == nil {
		opts = &AlignOptions{
			MinConfidence:  0.5,
			AllowUnaligned: true,
		}
	}

	var alignments []*TokenAlignment

	// Simple 1:1 alignment based on position for now
	// Real implementation would use Strong's numbers or statistical methods
	maxLen := len(source.Tokens)
	if len(target.Tokens) < maxLen {
		maxLen = len(target.Tokens)
	}

	for i := 0; i < maxLen; i++ {
		ta := &TokenAlignment{
			ID:           fmt.Sprintf("ta-%d", i),
			SourceTokens: []string{source.Tokens[i].ID},
			TargetTokens: []string{target.Tokens[i].ID},
			Confidence:   0.5,
			AlignType:    "one-to-one",
		}
		alignments = append(alignments, ta)
	}

	return alignments, nil
}
