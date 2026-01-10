package ir

// types.go - Consolidated IR schema type definitions
// This file contains all core IR (Intermediate Representation) types used throughout Mimicry.
// All format handlers should import these types from core/ir rather than defining their own.

// ModuleType represents the type of Bible module.
type ModuleType string

// Module type constants.
const (
	ModuleBible      ModuleType = "BIBLE"
	ModuleCommentary ModuleType = "COMMENTARY"
	ModuleDictionary ModuleType = "DICTIONARY"
	ModuleGenBook    ModuleType = "GENBOOK"
	ModuleDevotional ModuleType = "DEVOTIONAL"
)

// validModuleTypes is the set of valid module types.
var validModuleTypes = map[ModuleType]bool{
	ModuleBible:      true,
	ModuleCommentary: true,
	ModuleDictionary: true,
	ModuleGenBook:    true,
	ModuleDevotional: true,
}

// IsValid returns true if the module type is valid.
func (m ModuleType) IsValid() bool {
	return validModuleTypes[m]
}

// Corpus is the top-level container for a complete Bible module.
// It contains all documents (books), versification mappings, and metadata.
type Corpus struct {
	// ID is the unique identifier for this corpus (e.g., "KJV", "ESV").
	ID string `json:"id"`

	// Version is the IR schema version (e.g., "1.0.0").
	Version string `json:"version"`

	// ModuleType indicates the type of content (BIBLE, COMMENTARY, etc.).
	ModuleType ModuleType `json:"module_type"`

	// Versification is the versification system (e.g., "KJV", "Catholic", "LXX").
	Versification string `json:"versification,omitempty"`

	// Language is the BCP-47 language tag (e.g., "en", "he", "grc").
	Language string `json:"language,omitempty"`

	// Title is the human-readable title of the corpus.
	Title string `json:"title,omitempty"`

	// Description is an optional description of the corpus.
	Description string `json:"description,omitempty"`

	// Publisher is the publisher information (optional).
	Publisher string `json:"publisher,omitempty"`

	// Rights contains copyright and licensing information (optional).
	Rights string `json:"rights,omitempty"`

	// SourceFormat indicates the original format (e.g., "USFM", "JSON", "HTML").
	SourceFormat string `json:"source_format,omitempty"`

	// Documents contains all books, articles, or entries in this corpus.
	Documents []*Document `json:"documents,omitempty"`

	// MappingTables contains versification mappings between systems.
	MappingTables []*MappingTable `json:"mapping_tables,omitempty"`

	// SourceHash is the SHA-256 hash of the source artifact.
	SourceHash string `json:"source_hash,omitempty"`

	// LossClass indicates the fidelity of the extraction.
	LossClass LossClass `json:"loss_class,omitempty"`

	// CrossReferences contains cross-reference relationships.
	CrossReferences []*CrossReference `json:"cross_references,omitempty"`

	// Attributes contains additional metadata as key-value pairs.
	Attributes map[string]string `json:"attributes,omitempty"`
}

// AddCrossReference adds a cross-reference to the corpus.
func (c *Corpus) AddCrossReference(cr *CrossReference) {
	c.CrossReferences = append(c.CrossReferences, cr)
}

// BuildCrossRefIndex builds an index for efficient cross-reference lookup.
func (c *Corpus) BuildCrossRefIndex() *CrossRefIndex {
	index := NewCrossRefIndex()
	for _, cr := range c.CrossReferences {
		index.Add(cr)
	}
	return index
}

// Document represents a single book, article, or dictionary entry within a corpus.
type Document struct {
	// ID is the document identifier (e.g., "Gen", "Matt", "1John").
	ID string `json:"id"`

	// CanonicalRef is the primary scripture reference for this document.
	CanonicalRef *Ref `json:"canonical_ref,omitempty"`

	// Title is the human-readable title (e.g., "Genesis", "Matthew").
	Title string `json:"title,omitempty"`

	// Order is the position within the corpus (1-indexed).
	Order int `json:"order"`

	// ContentBlocks contains the text content of this document.
	ContentBlocks []*ContentBlock `json:"content_blocks,omitempty"`

	// Annotations contains stand-off annotations for this document.
	Annotations []*Annotation `json:"annotations,omitempty"`

	// Attributes contains additional document metadata.
	Attributes map[string]string `json:"attributes,omitempty"`
}

// ContentBlock represents a contiguous text unit (typically a paragraph or section).
type ContentBlock struct {
	// ID is the unique identifier within the document.
	ID string `json:"id"`

	// Sequence is the order within the document (0-indexed).
	Sequence int `json:"sequence"`

	// Text is the raw UTF-8 text content.
	Text string `json:"text"`

	// Tokens contains word-level breakdown for linguistic annotation.
	Tokens []*Token `json:"tokens,omitempty"`

	// Anchors contains position markers for stand-off markup.
	Anchors []*Anchor `json:"anchors,omitempty"`

	// Hash is the SHA-256 hash of the Text field.
	Hash string `json:"hash,omitempty"`

	// Attributes contains additional content block metadata.
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

// TokenType represents the type of a token.
type TokenType string

// Token type constants.
const (
	TokenWord        TokenType = "word"
	TokenWhitespace  TokenType = "whitespace"
	TokenPunctuation TokenType = "punctuation"
)

// Token represents a word or whitespace unit for linguistic annotation.
type Token struct {
	// ID is the unique identifier within the content block.
	ID string `json:"id"`

	// Index is the position in the token sequence (0-indexed).
	Index int `json:"index"`

	// CharStart is the UTF-8 byte offset where the token starts.
	CharStart int `json:"char_start"`

	// CharEnd is the UTF-8 byte offset where the token ends.
	CharEnd int `json:"char_end"`

	// Text is the token text.
	Text string `json:"text"`

	// Type is the token type (word, whitespace, punctuation).
	Type TokenType `json:"type"`

	// Lemma is the dictionary form (optional).
	Lemma string `json:"lemma,omitempty"`

	// Strongs contains Strong's numbers (e.g., ["H1234"]).
	Strongs []string `json:"strongs,omitempty"`

	// Morphology is the morphological code (optional).
	Morphology string `json:"morphology,omitempty"`
}

// IsWord returns true if this token is a word token.
func (t *Token) IsWord() bool {
	return t.Type == TokenWord
}

// Length returns the length of the token in bytes.
func (t *Token) Length() int {
	return t.CharEnd - t.CharStart
}

// Anchor represents a position marker within a content block for stand-off markup.
// Anchors allow spans to reference specific positions without inline markup.
type Anchor struct {
	// ID is the unique identifier within the content block.
	ID string `json:"id"`

	// ContentBlockID is the parent content block.
	ContentBlockID string `json:"content_block_id,omitempty"`

	// CharOffset is the UTF-8 byte offset within the content block.
	CharOffset int `json:"char_offset,omitempty"`

	// Position is an alternative way to specify position (used by some formats).
	Position int `json:"position,omitempty"`

	// TokenIndex is the token position (optional, -1 if not set).
	TokenIndex int `json:"token_index,omitempty"`

	// Hash is the hash of the ContentBlock at this anchor point.
	Hash string `json:"hash,omitempty"`

	// Spans contains spans starting at this anchor (used by some formats).
	Spans []*Span `json:"spans,omitempty"`
}

// SpanType represents the type of a span.
type SpanType string

// Span type constants.
const (
	SpanVerse      SpanType = "VERSE"
	SpanChapter    SpanType = "CHAPTER"
	SpanParagraph  SpanType = "PARAGRAPH"
	SpanPoetryLine SpanType = "POETRY_LINE"
	SpanQuotation  SpanType = "QUOTATION"
	SpanRedLetter  SpanType = "RED_LETTER"
	SpanNote       SpanType = "NOTE"
	SpanCrossRef   SpanType = "CROSS_REF"
	SpanSection    SpanType = "SECTION"
	SpanTitle      SpanType = "TITLE"
	SpanDivine     SpanType = "DIVINE_NAME"
	SpanEmphasis   SpanType = "EMPHASIS"
	SpanForeign    SpanType = "FOREIGN"
	SpanSelah      SpanType = "SELAH"
)

// validSpanTypes is the set of valid span types.
var validSpanTypes = map[SpanType]bool{
	SpanVerse:      true,
	SpanChapter:    true,
	SpanParagraph:  true,
	SpanPoetryLine: true,
	SpanQuotation:  true,
	SpanRedLetter:  true,
	SpanNote:       true,
	SpanCrossRef:   true,
	SpanSection:    true,
	SpanTitle:      true,
	SpanDivine:     true,
	SpanEmphasis:   true,
	SpanForeign:    true,
	SpanSelah:      true,
}

// IsValid returns true if the span type is valid.
func (s SpanType) IsValid() bool {
	return validSpanTypes[s]
}

// Span represents a region between two anchors. Spans can overlap freely,
// which is essential for representing structures like verses that span
// poetry lines or quotations that cross chapter boundaries.
type Span struct {
	// ID is the unique identifier within the document.
	ID string `json:"id"`

	// Type indicates the kind of span (VERSE, CHAPTER, POETRY_LINE, etc.).
	Type SpanType `json:"type"`

	// StartAnchorID is the opening anchor.
	StartAnchorID string `json:"start_anchor_id"`

	// EndAnchorID is the closing anchor.
	EndAnchorID string `json:"end_anchor_id,omitempty"`

	// Ref is the scripture reference (optional, for verse/chapter spans).
	Ref *Ref `json:"ref,omitempty"`

	// Attributes contains type-specific metadata.
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

// SetAttribute sets an attribute value.
func (s *Span) SetAttribute(key string, value interface{}) {
	if s.Attributes == nil {
		s.Attributes = make(map[string]interface{})
	}
	s.Attributes[key] = value
}

// GetAttribute gets an attribute value.
func (s *Span) GetAttribute(key string) (interface{}, bool) {
	if s.Attributes == nil {
		return nil, false
	}
	v, ok := s.Attributes[key]
	return v, ok
}

// AnnotationType represents the type of an annotation.
type AnnotationType string

// Annotation type constants.
const (
	AnnotationStrongs    AnnotationType = "STRONGS"
	AnnotationMorphology AnnotationType = "MORPHOLOGY"
	AnnotationFootnote   AnnotationType = "FOOTNOTE"
	AnnotationCrossRef   AnnotationType = "CROSS_REF"
	AnnotationGloss      AnnotationType = "GLOSS"
	AnnotationSource     AnnotationType = "SOURCE"
	AnnotationAlternate  AnnotationType = "ALTERNATE"
	AnnotationVariant    AnnotationType = "VARIANT"
)

// validAnnotationTypes is the set of valid annotation types.
var validAnnotationTypes = map[AnnotationType]bool{
	AnnotationStrongs:    true,
	AnnotationMorphology: true,
	AnnotationFootnote:   true,
	AnnotationCrossRef:   true,
	AnnotationGloss:      true,
	AnnotationSource:     true,
	AnnotationAlternate:  true,
	AnnotationVariant:    true,
}

// IsValid returns true if the annotation type is valid.
func (a AnnotationType) IsValid() bool {
	return validAnnotationTypes[a]
}

// Annotation represents data attached to a span.
type Annotation struct {
	// ID is the unique identifier within the document.
	ID string `json:"id"`

	// SpanID is the parent span this annotation is attached to.
	SpanID string `json:"span_id"`

	// Type indicates the kind of annotation (STRONGS, MORPHOLOGY, etc.).
	Type AnnotationType `json:"type"`

	// Value is the annotation data (type-specific).
	Value interface{} `json:"value"`

	// Confidence is the confidence level (0.0-1.0, optional).
	Confidence float64 `json:"confidence,omitempty"`

	// Source is the attribution for this annotation (optional).
	Source string `json:"source,omitempty"`
}

// Ref is defined in ref.go and represents a canonical scripture reference.
// It is re-exported here for documentation purposes but the actual definition
// is in ref.go to maintain the existing parser functionality.
