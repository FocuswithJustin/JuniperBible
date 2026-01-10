package ir

import (
	"encoding/json"
	"testing"
)

func TestModuleTypeConstants(t *testing.T) {
	tests := []struct {
		mt   ModuleType
		want string
	}{
		{ModuleBible, "BIBLE"},
		{ModuleCommentary, "COMMENTARY"},
		{ModuleDictionary, "DICTIONARY"},
		{ModuleGenBook, "GENBOOK"},
		{ModuleDevotional, "DEVOTIONAL"},
	}

	for _, tt := range tests {
		if string(tt.mt) != tt.want {
			t.Errorf("ModuleType = %q, want %q", tt.mt, tt.want)
		}
	}
}

func TestModuleTypeValidation(t *testing.T) {
	tests := []struct {
		mt    ModuleType
		valid bool
	}{
		{ModuleBible, true},
		{ModuleCommentary, true},
		{ModuleDictionary, true},
		{ModuleGenBook, true},
		{ModuleDevotional, true},
		{ModuleType("INVALID"), false},
		{ModuleType(""), false},
	}

	for _, tt := range tests {
		if got := tt.mt.IsValid(); got != tt.valid {
			t.Errorf("ModuleType(%q).IsValid() = %v, want %v", tt.mt, got, tt.valid)
		}
	}
}

func TestCorpusJSON(t *testing.T) {
	corpus := &Corpus{
		ID:            "KJV",
		Version:       "1.0.0",
		ModuleType:    ModuleBible,
		Versification: "KJV",
		Language:      "en",
		Title:         "King James Version",
		SourceHash:    "abc123",
		LossClass:     LossL0,
	}

	// Marshal to JSON
	data, err := json.Marshal(corpus)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	// Unmarshal back
	var decoded Corpus
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	// Verify fields
	if decoded.ID != corpus.ID {
		t.Errorf("ID = %q, want %q", decoded.ID, corpus.ID)
	}
	if decoded.Version != corpus.Version {
		t.Errorf("Version = %q, want %q", decoded.Version, corpus.Version)
	}
	if decoded.ModuleType != corpus.ModuleType {
		t.Errorf("ModuleType = %q, want %q", decoded.ModuleType, corpus.ModuleType)
	}
	if decoded.Versification != corpus.Versification {
		t.Errorf("Versification = %q, want %q", decoded.Versification, corpus.Versification)
	}
	if decoded.Language != corpus.Language {
		t.Errorf("Language = %q, want %q", decoded.Language, corpus.Language)
	}
	if decoded.Title != corpus.Title {
		t.Errorf("Title = %q, want %q", decoded.Title, corpus.Title)
	}
	if decoded.SourceHash != corpus.SourceHash {
		t.Errorf("SourceHash = %q, want %q", decoded.SourceHash, corpus.SourceHash)
	}
	if decoded.LossClass != corpus.LossClass {
		t.Errorf("LossClass = %q, want %q", decoded.LossClass, corpus.LossClass)
	}
}

func TestCorpusWithDocuments(t *testing.T) {
	corpus := &Corpus{
		ID:         "KJV",
		Version:    "1.0.0",
		ModuleType: ModuleBible,
		Documents: []*Document{
			{
				ID:    "Gen",
				Title: "Genesis",
				Order: 1,
			},
			{
				ID:    "Exod",
				Title: "Exodus",
				Order: 2,
			},
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(corpus)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	// Unmarshal back
	var decoded Corpus
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	// Verify documents
	if len(decoded.Documents) != 2 {
		t.Fatalf("len(Documents) = %d, want 2", len(decoded.Documents))
	}
	if decoded.Documents[0].ID != "Gen" {
		t.Errorf("Documents[0].ID = %q, want %q", decoded.Documents[0].ID, "Gen")
	}
	if decoded.Documents[1].ID != "Exod" {
		t.Errorf("Documents[1].ID = %q, want %q", decoded.Documents[1].ID, "Exod")
	}
}

func TestDocumentJSON(t *testing.T) {
	doc := &Document{
		ID:    "Gen",
		Title: "Genesis",
		Order: 1,
		CanonicalRef: &Ref{
			Book:   "Gen",
			OSISID: "Gen",
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(doc)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	// Unmarshal back
	var decoded Document
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	// Verify fields
	if decoded.ID != doc.ID {
		t.Errorf("ID = %q, want %q", decoded.ID, doc.ID)
	}
	if decoded.Title != doc.Title {
		t.Errorf("Title = %q, want %q", decoded.Title, doc.Title)
	}
	if decoded.Order != doc.Order {
		t.Errorf("Order = %d, want %d", decoded.Order, doc.Order)
	}
	if decoded.CanonicalRef == nil {
		t.Fatal("CanonicalRef is nil")
	}
	if decoded.CanonicalRef.Book != doc.CanonicalRef.Book {
		t.Errorf("CanonicalRef.Book = %q, want %q", decoded.CanonicalRef.Book, doc.CanonicalRef.Book)
	}
}

func TestCorpusJSONFieldNames(t *testing.T) {
	corpus := &Corpus{
		ID:            "test",
		Version:       "1.0.0",
		ModuleType:    ModuleBible,
		Versification: "KJV",
		Language:      "en",
		Title:         "Test",
		SourceHash:    "hash",
		LossClass:     LossL0,
	}

	data, err := json.Marshal(corpus)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	// Parse as generic map to check field names
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	expectedFields := []string{
		"id", "version", "module_type", "versification",
		"language", "title", "source_hash", "loss_class",
	}

	for _, field := range expectedFields {
		if _, ok := m[field]; !ok {
			t.Errorf("missing expected JSON field: %s", field)
		}
	}
}

func TestDocumentWithContentBlocks(t *testing.T) {
	doc := &Document{
		ID:    "Gen",
		Title: "Genesis",
		Order: 1,
		ContentBlocks: []*ContentBlock{
			{
				ID:       "cb1",
				Sequence: 0,
				Text:     "In the beginning God created the heaven and the earth.",
			},
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(doc)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	// Unmarshal back
	var decoded Document
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	// Verify content blocks
	if len(decoded.ContentBlocks) != 1 {
		t.Fatalf("len(ContentBlocks) = %d, want 1", len(decoded.ContentBlocks))
	}
	if decoded.ContentBlocks[0].Text != doc.ContentBlocks[0].Text {
		t.Errorf("ContentBlocks[0].Text = %q, want %q",
			decoded.ContentBlocks[0].Text, doc.ContentBlocks[0].Text)
	}
}

func TestDocumentWithAnnotations(t *testing.T) {
	doc := &Document{
		ID:    "Gen",
		Title: "Genesis",
		Order: 1,
		Annotations: []*Annotation{
			{
				ID:     "ann1",
				SpanID: "span1",
				Type:   AnnotationStrongs,
				Value:  "H1234",
			},
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(doc)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	// Unmarshal back
	var decoded Document
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	// Verify annotations
	if len(decoded.Annotations) != 1 {
		t.Fatalf("len(Annotations) = %d, want 1", len(decoded.Annotations))
	}
	if decoded.Annotations[0].Type != AnnotationStrongs {
		t.Errorf("Annotations[0].Type = %q, want %q",
			decoded.Annotations[0].Type, AnnotationStrongs)
	}
}

func TestCorpusWithMappingTables(t *testing.T) {
	corpus := &Corpus{
		ID:         "test",
		Version:    "1.0.0",
		ModuleType: ModuleBible,
		MappingTables: []*MappingTable{
			{
				ID:         "kjv-to-lxx",
				FromSystem: "KJV",
				ToSystem:   "LXX",
			},
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(corpus)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	// Unmarshal back
	var decoded Corpus
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	// Verify mapping tables
	if len(decoded.MappingTables) != 1 {
		t.Fatalf("len(MappingTables) = %d, want 1", len(decoded.MappingTables))
	}
	if decoded.MappingTables[0].ID != "kjv-to-lxx" {
		t.Errorf("MappingTables[0].ID = %q, want %q",
			decoded.MappingTables[0].ID, "kjv-to-lxx")
	}
}

func TestEmptyCorpus(t *testing.T) {
	corpus := &Corpus{}

	data, err := json.Marshal(corpus)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	var decoded Corpus
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	// Empty corpus should have nil slices
	if decoded.Documents != nil && len(decoded.Documents) != 0 {
		t.Errorf("Documents should be nil or empty")
	}
	if decoded.MappingTables != nil && len(decoded.MappingTables) != 0 {
		t.Errorf("MappingTables should be nil or empty")
	}
}
