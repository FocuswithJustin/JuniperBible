package ipc

// IR Types shared across all format plugins.
// These mirror the core/ir package types but are used in plugin JSON serialization.

// Corpus represents a complete text collection (Bible, commentary, dictionary, etc).
type Corpus struct {
	ID            string            `json:"id"`
	Version       string            `json:"version"`
	ModuleType    string            `json:"module_type"`
	Versification string            `json:"versification,omitempty"`
	Language      string            `json:"language,omitempty"`
	Title         string            `json:"title,omitempty"`
	Description   string            `json:"description,omitempty"`
	Publisher     string            `json:"publisher,omitempty"`
	Rights        string            `json:"rights,omitempty"`
	SourceFormat  string            `json:"source_format,omitempty"`
	Documents     []*Document       `json:"documents,omitempty"`
	SourceHash    string            `json:"source_hash,omitempty"`
	LossClass     string            `json:"loss_class,omitempty"`
	Attributes    map[string]string `json:"attributes,omitempty"`
}

// Document represents a single document within a corpus (e.g., a Bible book).
type Document struct {
	ID            string            `json:"id"`
	Title         string            `json:"title,omitempty"`
	Order         int               `json:"order"`
	ContentBlocks []*ContentBlock   `json:"content_blocks,omitempty"`
	Attributes    map[string]string `json:"attributes,omitempty"`
}

// ContentBlock represents a unit of content with stand-off markup.
type ContentBlock struct {
	ID         string                 `json:"id"`
	Sequence   int                    `json:"sequence"`
	Text       string                 `json:"text"`
	Tokens     []*Token               `json:"tokens,omitempty"`
	Anchors    []*Anchor              `json:"anchors,omitempty"`
	Hash       string                 `json:"hash,omitempty"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

// Token represents a tokenized word or morpheme.
type Token struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Text     string `json:"text"`
	StartPos int    `json:"start_pos"`
	EndPos   int    `json:"end_pos"`
}

// Anchor represents a position in the text where spans can attach.
type Anchor struct {
	ID       string  `json:"id"`
	Position int     `json:"position"`
	Spans    []*Span `json:"spans,omitempty"`
}

// Span represents markup that spans from one anchor to another.
type Span struct {
	ID            string                 `json:"id"`
	Type          string                 `json:"type"`
	StartAnchorID string                 `json:"start_anchor_id"`
	EndAnchorID   string                 `json:"end_anchor_id,omitempty"`
	Ref           *Ref                   `json:"ref,omitempty"`
	Attributes    map[string]interface{} `json:"attributes,omitempty"`
}

// Ref represents a biblical or textual reference.
type Ref struct {
	Book     string `json:"book"`
	Chapter  int    `json:"chapter,omitempty"`
	Verse    int    `json:"verse,omitempty"`
	VerseEnd int    `json:"verse_end,omitempty"`
	SubVerse string `json:"sub_verse,omitempty"`
	OSISID   string `json:"osis_id,omitempty"`
}

// ParallelCorpus represents multiple aligned corpora (e.g., parallel translations).
type ParallelCorpus struct {
	ID               string                 `json:"id"`
	Version          string                 `json:"version"`
	BaseCorpus       *CorpusRef             `json:"base_corpus,omitempty"`
	Corpora          []*CorpusRef           `json:"corpora"`
	Alignments       []*Alignment           `json:"alignments,omitempty"`
	DefaultAlignment string                 `json:"default_alignment"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// CorpusRef references a corpus in a parallel corpus.
type CorpusRef struct {
	ID       string `json:"id"`
	Language string `json:"language"`
	Title    string `json:"title,omitempty"`
}

// Alignment represents alignment between corpora.
type Alignment struct {
	ID    string         `json:"id"`
	Level string         `json:"level"`
	Units []*AlignedUnit `json:"units,omitempty"`
}

// AlignedUnit represents a single aligned unit across translations.
type AlignedUnit struct {
	ID              string            `json:"id"`
	Ref             *Ref              `json:"ref,omitempty"`
	Texts           map[string]string `json:"texts"`
	Level           string            `json:"level"`
	TokenAlignments []*TokenAlignment `json:"token_alignments,omitempty"`
}

// TokenAlignment represents word-level alignment.
type TokenAlignment struct {
	ID           string   `json:"id"`
	SourceTokens []string `json:"source_tokens"`
	TargetTokens []string `json:"target_tokens"`
	Confidence   float64  `json:"confidence,omitempty"`
	AlignType    string   `json:"align_type,omitempty"`
}

// InterlinearLine represents a line of interlinear text with multiple layers.
type InterlinearLine struct {
	Ref    *Ref                         `json:"ref"`
	Layers map[string]*InterlinearLayer `json:"layers"`
}

// InterlinearLayer represents one layer of interlinear text (e.g., Greek, English, gloss).
type InterlinearLayer struct {
	CorpusID string   `json:"corpus_id"`
	Tokens   []string `json:"tokens"`
	Label    string   `json:"label"`
}
