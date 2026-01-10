package ir

// VersificationID represents a versification system identifier.
type VersificationID string

// Common versification system constants.
const (
	VersificationKJV       VersificationID = "KJV"
	VersificationCatholic  VersificationID = "Catholic"
	VersificationLXX       VersificationID = "LXX"
	VersificationVulgate   VersificationID = "Vulgate"
	VersificationEthiopian VersificationID = "Ethiopian"
	VersificationSynodal   VersificationID = "Synodal"
	VersificationMT        VersificationID = "MT" // Masoretic Text
	VersificationNRSV      VersificationID = "NRSV"

	// Phase 16.1: Additional versification systems
	VersificationArmenian  VersificationID = "Armenian"
	VersificationGeorgian  VersificationID = "Georgian"
	VersificationSlavonic  VersificationID = "Slavonic"
	VersificationSyriac    VersificationID = "Syriac"
	VersificationArabic    VersificationID = "Arabic"
	VersificationDSS       VersificationID = "DSS"       // Dead Sea Scrolls
	VersificationSamaritan VersificationID = "Samaritan" // Samaritan Pentateuch
	VersificationBHS       VersificationID = "BHS"       // Biblia Hebraica Stuttgartensia
	VersificationNA28      VersificationID = "NA28"      // Nestle-Aland 28th edition
)

// validVersificationIDs is the set of valid versification IDs.
var validVersificationIDs = map[VersificationID]bool{
	VersificationKJV:       true,
	VersificationCatholic:  true,
	VersificationLXX:       true,
	VersificationVulgate:   true,
	VersificationEthiopian: true,
	VersificationSynodal:   true,
	VersificationMT:        true,
	VersificationNRSV:      true,
	// Phase 16.1: Additional systems
	VersificationArmenian:  true,
	VersificationGeorgian:  true,
	VersificationSlavonic:  true,
	VersificationSyriac:    true,
	VersificationArabic:    true,
	VersificationDSS:       true,
	VersificationSamaritan: true,
	VersificationBHS:       true,
	VersificationNA28:      true,
}

// IsValid returns true if the versification ID is valid.
func (v VersificationID) IsValid() bool {
	return validVersificationIDs[v]
}

// MappingType represents the type of reference mapping.
type MappingType string

// Mapping type constants.
const (
	// MappingExact indicates a 1:1 verse correspondence.
	MappingExact MappingType = "exact"

	// MappingSplit indicates one verse splits into multiple.
	MappingSplit MappingType = "split"

	// MappingMerge indicates multiple verses merge into one.
	MappingMerge MappingType = "merge"

	// MappingMissing indicates a verse exists in source but not target.
	MappingMissing MappingType = "missing"

	// MappingAdded indicates a verse exists in target but not source.
	MappingAdded MappingType = "added"

	// MappingReordered indicates verses are reordered between systems.
	MappingReordered MappingType = "reordered"
)

// RefMapping represents a single reference mapping between versification systems.
type RefMapping struct {
	// From is the source reference.
	From *Ref `json:"from"`

	// To is the target reference (or references for split mappings).
	To *Ref `json:"to"`

	// ToRefs is used when mapping to multiple references (split).
	ToRefs []*Ref `json:"to_refs,omitempty"`

	// Type indicates the mapping type (exact, split, merge, etc.).
	Type MappingType `json:"type"`

	// Note is an optional explanation for the mapping.
	Note string `json:"note,omitempty"`
}

// MappingTable represents versification mappings between two systems.
type MappingTable struct {
	// ID is the unique identifier for this mapping table.
	ID string `json:"id"`

	// FromSystem is the source versification system.
	FromSystem VersificationID `json:"from_system"`

	// ToSystem is the target versification system.
	ToSystem VersificationID `json:"to_system"`

	// Mappings contains individual reference mappings.
	Mappings []*RefMapping `json:"mappings,omitempty"`

	// Hash is the SHA-256 hash for change detection.
	Hash string `json:"hash,omitempty"`
}

// Lookup finds the mapping for a given reference.
func (mt *MappingTable) Lookup(ref *Ref) *RefMapping {
	for _, m := range mt.Mappings {
		if m.From.Book == ref.Book &&
			m.From.Chapter == ref.Chapter &&
			m.From.Verse == ref.Verse {
			return m
		}
	}
	return nil
}

// AddMapping adds a new mapping to the table.
func (mt *MappingTable) AddMapping(from, to *Ref, mappingType MappingType) {
	mt.Mappings = append(mt.Mappings, &RefMapping{
		From: from,
		To:   to,
		Type: mappingType,
	})
}

// MapRef maps a reference from the source system to the target system.
// Returns nil if no mapping exists (assumes identity mapping).
func (mt *MappingTable) MapRef(ref *Ref) *Ref {
	mapping := mt.Lookup(ref)
	if mapping == nil {
		// No explicit mapping - return identity
		return ref
	}
	return mapping.To
}

// Phase 16.1: MappingRegistry manages versification mapping tables.
type MappingRegistry struct {
	tables map[string]*MappingTable // key: "from-to"
}

// NewMappingRegistry creates a new mapping registry.
func NewMappingRegistry() *MappingRegistry {
	return &MappingRegistry{
		tables: make(map[string]*MappingTable),
	}
}

// makeKey creates a lookup key for two versification systems.
func makeKey(from, to VersificationID) string {
	return string(from) + "-" + string(to)
}

// RegisterTable adds a mapping table to the registry.
func (r *MappingRegistry) RegisterTable(table *MappingTable) {
	key := makeKey(table.FromSystem, table.ToSystem)
	r.tables[key] = table
}

// GetTable retrieves a direct mapping table between two systems.
func (r *MappingRegistry) GetTable(from, to VersificationID) *MappingTable {
	key := makeKey(from, to)
	return r.tables[key]
}

// GetChainedMapping finds a mapping path between two systems via intermediates.
// Returns a composite MappingTable that chains the mappings together.
func (r *MappingRegistry) GetChainedMapping(from, to VersificationID) *MappingTable {
	// First try direct mapping
	if direct := r.GetTable(from, to); direct != nil {
		return direct
	}

	// Try to find an intermediate system (one hop)
	for _, table := range r.tables {
		if table.FromSystem == from {
			// Found a table starting from 'from'
			intermediate := table.ToSystem
			if secondHop := r.GetTable(intermediate, to); secondHop != nil {
				// Build chained table
				return r.buildChainedTable(table, secondHop)
			}
		}
	}

	return nil
}

// buildChainedTable creates a composite mapping table from two sequential tables.
func (r *MappingRegistry) buildChainedTable(first, second *MappingTable) *MappingTable {
	chained := &MappingTable{
		ID:         first.ID + "+" + second.ID,
		FromSystem: first.FromSystem,
		ToSystem:   second.ToSystem,
	}

	// For each mapping in the first table, chain through the second
	for _, m := range first.Mappings {
		intermediate := m.To
		if intermediate == nil {
			continue
		}
		final := second.MapRef(intermediate)
		chained.AddMapping(m.From, final, m.Type)
	}

	return chained
}

// MapRefBetweenSystems maps a reference from one versification system to another.
func (r *MappingRegistry) MapRefBetweenSystems(ref *Ref, from, to VersificationID) (*Ref, error) {
	table := r.GetTable(from, to)
	if table == nil {
		table = r.GetChainedMapping(from, to)
	}
	if table == nil {
		// No mapping found - return identity
		return ref, nil
	}
	return table.MapRef(ref), nil
}

// SplitRef returns the target references for a split mapping.
// If the mapping is not a split, returns a slice with just the To ref.
func SplitRef(mapping *RefMapping) []*Ref {
	if mapping.Type == MappingSplit && len(mapping.ToRefs) > 0 {
		return mapping.ToRefs
	}
	if mapping.To != nil {
		return []*Ref{mapping.To}
	}
	return nil
}

// MergeRefs combines multiple references into a single reference.
// Returns the first reference as the canonical merged reference.
func MergeRefs(refs []*Ref) *Ref {
	if len(refs) == 0 {
		return nil
	}
	return refs[0]
}

// ApplyToCorpus applies the versification mapping to an entire corpus.
// Returns a new corpus with mapped references and a loss report.
func (mt *MappingTable) ApplyToCorpus(corpus *Corpus) (*Corpus, *LossReport, error) {
	// Create a copy of the corpus with new versification
	mapped := &Corpus{
		ID:            corpus.ID,
		Version:       corpus.Version,
		Title:         corpus.Title,
		Language:      corpus.Language,
		Versification: string(mt.ToSystem),
		ModuleType:    corpus.ModuleType,
		SourceHash:    corpus.SourceHash,
		LossClass:     corpus.LossClass,
	}

	lossReport := &LossReport{
		SourceFormat: string(mt.FromSystem),
		TargetFormat: string(mt.ToSystem),
		LossClass:    LossL0, // Start optimistic
	}

	// Map each document
	for _, doc := range corpus.Documents {
		mappedDoc := &Document{
			ID:    doc.ID,
			Title: doc.Title,
			Order: doc.Order,
		}

		// Map document's canonical reference if present
		if doc.CanonicalRef != nil {
			mappedDoc.CanonicalRef = mt.MapRef(doc.CanonicalRef)
		}

		// Copy content blocks (they don't have references, those are in annotations)
		for _, block := range doc.ContentBlocks {
			mappedBlock := &ContentBlock{
				ID:       block.ID,
				Sequence: block.Sequence,
				Text:     block.Text,
				Tokens:   block.Tokens,
				Anchors:  block.Anchors,
				Hash:     block.Hash,
			}
			mappedDoc.ContentBlocks = append(mappedDoc.ContentBlocks, mappedBlock)
		}

		// Copy annotations (they reference spans, not verses directly)
		for _, ann := range doc.Annotations {
			mappedAnn := &Annotation{
				ID:         ann.ID,
				SpanID:     ann.SpanID,
				Type:       ann.Type,
				Value:      ann.Value,
				Confidence: ann.Confidence,
			}
			mappedDoc.Annotations = append(mappedDoc.Annotations, mappedAnn)
		}

		mapped.Documents = append(mapped.Documents, mappedDoc)
	}

	// Copy and add mapped mapping tables
	mapped.MappingTables = append(mapped.MappingTables, mt)

	return mapped, lossReport, nil
}
