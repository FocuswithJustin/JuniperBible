// Package ir provides cross-reference types for Bible text relationships.
package ir

import (
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

// CrossRefType represents the type of cross-reference relationship.
type CrossRefType string

// Cross-reference type constants.
const (
	// CrossRefQuotation indicates a direct quote.
	CrossRefQuotation CrossRefType = "quotation"

	// CrossRefAllusion indicates an indirect reference or allusion.
	CrossRefAllusion CrossRefType = "allusion"

	// CrossRefParallel indicates parallel passages (e.g., Synoptic Gospels).
	CrossRefParallel CrossRefType = "parallel"

	// CrossRefProphecy indicates a prophecy-fulfillment relationship.
	CrossRefProphecy CrossRefType = "prophecy"

	// CrossRefTypology indicates a type-antitype relationship.
	CrossRefTypology CrossRefType = "typology"

	// CrossRefGeneral indicates a general related reference.
	CrossRefGeneral CrossRefType = "general"
)

// CrossReference represents a relationship between two scripture passages.
type CrossReference struct {
	// ID is the unique identifier for this cross-reference.
	ID string `json:"id"`

	// SourceRef is the referring passage.
	SourceRef *Ref `json:"source_ref"`

	// TargetRef is the referenced passage.
	TargetRef *Ref `json:"target_ref"`

	// Type indicates the kind of cross-reference.
	Type CrossRefType `json:"type"`

	// Label is an optional display label (e.g., "cf.", "see also").
	Label string `json:"label,omitempty"`

	// Notes provides optional explanatory notes.
	Notes string `json:"notes,omitempty"`

	// Confidence is the confidence level (0.0-1.0).
	Confidence float64 `json:"confidence,omitempty"`

	// Source indicates where this cross-reference came from (e.g., "TSK", "NASB").
	Source string `json:"source,omitempty"`
}

// CrossRefIndex provides efficient lookup of cross-references.
type CrossRefIndex struct {
	// BySource maps source OSISID to cross-references.
	BySource map[string][]*CrossReference

	// ByTarget maps target OSISID to cross-references.
	ByTarget map[string][]*CrossReference

	// All contains all cross-references.
	All []*CrossReference
}

// NewCrossRefIndex creates a new empty cross-reference index.
func NewCrossRefIndex() *CrossRefIndex {
	return &CrossRefIndex{
		BySource: make(map[string][]*CrossReference),
		ByTarget: make(map[string][]*CrossReference),
		All:      nil,
	}
}

// Add adds a cross-reference to the index.
func (idx *CrossRefIndex) Add(cr *CrossReference) {
	idx.All = append(idx.All, cr)

	if cr.SourceRef != nil && cr.SourceRef.OSISID != "" {
		idx.BySource[cr.SourceRef.OSISID] = append(idx.BySource[cr.SourceRef.OSISID], cr)
	}
	if cr.TargetRef != nil && cr.TargetRef.OSISID != "" {
		idx.ByTarget[cr.TargetRef.OSISID] = append(idx.ByTarget[cr.TargetRef.OSISID], cr)
	}
}

// GetBySource returns all cross-references from a given source reference.
func (idx *CrossRefIndex) GetBySource(osisID string) []*CrossReference {
	return idx.BySource[osisID]
}

// GetByTarget returns all cross-references pointing to a given target reference.
func (idx *CrossRefIndex) GetByTarget(osisID string) []*CrossReference {
	return idx.ByTarget[osisID]
}

// crossRefGrammar is the participle grammar for human-readable references.
// Examples: "Gen 1:1", "Matt 5:3-12", "1John 3:16"
//
//nolint:govet // participle grammar tags are not standard struct tags
type crossRefGrammar struct {
	BookPrefix string `@Int?`
	BookName   string `@Ident`
	Chapter    int    `@Int`
	Colon      string `":"`
	Verse      int    `@Int`
	Range      *int   `( "-" @Int )?`
}

// crossRefLexer defines the lexer for human-readable references.
var crossRefLexer = lexer.MustSimple([]lexer.SimpleRule{
	{Name: "Int", Pattern: `[0-9]+`},
	{Name: "Ident", Pattern: `[A-Za-z]+`},
	{Name: "Punct", Pattern: `[:\-]`},
	{Name: "Whitespace", Pattern: `\s+`},
})

// crossRefParser is the participle parser for human-readable references.
var crossRefParser = participle.MustBuild[crossRefGrammar](
	participle.Lexer(crossRefLexer),
	participle.Elide("Whitespace"),
)

// ParseCrossRefString parses a cross-reference string into references.
// Supports formats like: "Gen 1:1", "Gen 1:1-3", "Matt 5:3-12", "Gen 1:1; Exod 2:3"
func ParseCrossRefString(s string) ([]*Ref, error) {
	var refs []*Ref

	// Remove common prefixes
	s = strings.TrimPrefix(s, "cf. ")
	s = strings.TrimPrefix(s, "cf.")
	s = strings.TrimPrefix(s, "see ")
	s = strings.TrimSpace(s)

	// Split by semicolon for multiple references
	parts := strings.Split(s, ";")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		ref := parseSimpleRef(part)
		if ref != nil {
			refs = append(refs, ref)
		}
	}

	return refs, nil
}

// parseSimpleRef parses a simple reference like "Gen 1:1" or "Matt 5:3-12"
func parseSimpleRef(s string) *Ref {
	parsed, err := crossRefParser.ParseString("", s)
	if err != nil {
		return nil
	}

	book := normalizeBookName(parsed.BookPrefix + parsed.BookName)

	ref := &Ref{
		Book:    book,
		Chapter: parsed.Chapter,
		Verse:   parsed.Verse,
	}

	if parsed.Range != nil {
		ref.VerseEnd = *parsed.Range
	}

	ref.OSISID = ref.String()
	return ref
}

// normalizeBookName converts various book name formats to standard form.
func normalizeBookName(name string) string {
	name = strings.TrimSpace(name)

	// Common abbreviations to standard OSIS book names
	abbrevs := map[string]string{
		"Gen":    "Gen",
		"Exod":   "Exod",
		"Exodus": "Exod",
		"Lev":    "Lev",
		"Num":    "Num",
		"Deut":   "Deut",
		"Matt":   "Matt",
		"Mark":   "Mark",
		"Luke":   "Luke",
		"John":   "John",
		"Acts":   "Acts",
		"Rom":    "Rom",
		"1Cor":   "1Cor",
		"2Cor":   "2Cor",
		"Gal":    "Gal",
		"Eph":    "Eph",
		"Phil":   "Phil",
		"Col":    "Col",
		"1Thess": "1Thess",
		"2Thess": "2Thess",
		"1Tim":   "1Tim",
		"2Tim":   "2Tim",
		"Titus":  "Titus",
		"Phlm":   "Phlm",
		"Heb":    "Heb",
		"Jas":    "Jas",
		"1Pet":   "1Pet",
		"2Pet":   "2Pet",
		"1John":  "1John",
		"2John":  "2John",
		"3John":  "3John",
		"Jude":   "Jude",
		"Rev":    "Rev",
		"Ps":     "Ps",
		"Prov":   "Prov",
		"Eccl":   "Eccl",
		"Song":   "Song",
		"Isa":    "Isa",
		"Jer":    "Jer",
		"Lam":    "Lam",
		"Ezek":   "Ezek",
		"Dan":    "Dan",
		"Hos":    "Hos",
		"Joel":   "Joel",
		"Amos":   "Amos",
		"Obad":   "Obad",
		"Jonah":  "Jonah",
		"Mic":    "Mic",
		"Nah":    "Nah",
		"Hab":    "Hab",
		"Zeph":   "Zeph",
		"Hag":    "Hag",
		"Zech":   "Zech",
		"Mal":    "Mal",
	}

	if standard, ok := abbrevs[name]; ok {
		return standard
	}
	return name
}
