package ir

import (
	"errors"
	"testing"
)

func TestHashBytes(t *testing.T) {
	// Test basic hashing
	data := []byte("In the beginning God created the heaven and the earth.")
	hash := HashBytes(data)

	// Should be 64 hex characters (SHA-256)
	if len(hash) != 64 {
		t.Errorf("hash length = %d, want 64", len(hash))
	}

	// Same input should produce same hash
	hash2 := HashBytes(data)
	if hash != hash2 {
		t.Errorf("same data produced different hashes: %q vs %q", hash, hash2)
	}

	// Different input should produce different hash
	hash3 := HashBytes([]byte("Different content"))
	if hash == hash3 {
		t.Error("different data produced same hash")
	}
}

func TestHashString(t *testing.T) {
	text := "In the beginning God created the heaven and the earth."
	hash := HashString(text)

	// Should be 64 hex characters
	if len(hash) != 64 {
		t.Errorf("hash length = %d, want 64", len(hash))
	}

	// HashString and HashBytes should produce same result
	hashBytes := HashBytes([]byte(text))
	if hash != hashBytes {
		t.Errorf("HashString and HashBytes differ: %q vs %q", hash, hashBytes)
	}
}

func TestHashCorpus(t *testing.T) {
	corpus := &Corpus{
		ID:         "KJV",
		Version:    "1.0.0",
		ModuleType: ModuleBible,
		Title:      "King James Version",
	}

	hash, err := HashCorpus(corpus)
	if err != nil {
		t.Fatalf("HashCorpus failed: %v", err)
	}

	// Should be 64 hex characters
	if len(hash) != 64 {
		t.Errorf("hash length = %d, want 64", len(hash))
	}

	// Same corpus should produce same hash
	hash2, err := HashCorpus(corpus)
	if err != nil {
		t.Fatalf("HashCorpus failed: %v", err)
	}
	if hash != hash2 {
		t.Errorf("same corpus produced different hashes")
	}

	// Modified corpus should produce different hash
	corpus.Title = "Modified Title"
	hash3, err := HashCorpus(corpus)
	if err != nil {
		t.Fatalf("HashCorpus failed: %v", err)
	}
	if hash == hash3 {
		t.Error("modified corpus produced same hash")
	}
}

func TestHashDocument(t *testing.T) {
	doc := &Document{
		ID:    "Gen",
		Title: "Genesis",
		Order: 1,
	}

	hash, err := HashDocument(doc)
	if err != nil {
		t.Fatalf("HashDocument failed: %v", err)
	}

	// Should be 64 hex characters
	if len(hash) != 64 {
		t.Errorf("hash length = %d, want 64", len(hash))
	}

	// Same document should produce same hash
	hash2, err := HashDocument(doc)
	if err != nil {
		t.Fatalf("HashDocument failed: %v", err)
	}
	if hash != hash2 {
		t.Errorf("same document produced different hashes")
	}
}

func TestHashContentBlock(t *testing.T) {
	cb := &ContentBlock{
		ID:   "cb1",
		Text: "In the beginning God created the heaven and the earth.",
	}

	hash := HashContentBlock(cb)

	// Should be 64 hex characters
	if len(hash) != 64 {
		t.Errorf("hash length = %d, want 64", len(hash))
	}

	// Hash should match the Text hash
	textHash := HashString(cb.Text)
	if hash != textHash {
		t.Errorf("ContentBlock hash differs from text hash")
	}
}

func TestHashDeterminism(t *testing.T) {
	// Create same corpus multiple times and verify deterministic hashing
	var firstHash string
	for i := 0; i < 10; i++ {
		corpus := &Corpus{
			ID:         "test",
			Version:    "1.0.0",
			ModuleType: ModuleBible,
			Documents: []*Document{
				{
					ID:    "Gen",
					Title: "Genesis",
					Order: 1,
					ContentBlocks: []*ContentBlock{
						{
							ID:   "cb1",
							Text: "In the beginning",
						},
					},
				},
			},
		}

		hash, err := HashCorpus(corpus)
		if err != nil {
			t.Fatalf("HashCorpus failed: %v", err)
		}

		// All iterations should produce the same hash
		if i == 0 {
			firstHash = hash
		} else if hash != firstHash {
			t.Errorf("iteration %d: hash mismatch (non-deterministic)", i)
		}
	}
}

func TestHashEmpty(t *testing.T) {
	// Empty string
	hash := HashString("")
	if len(hash) != 64 {
		t.Errorf("empty string hash length = %d, want 64", len(hash))
	}

	// Empty bytes
	hashBytes := HashBytes(nil)
	if len(hashBytes) != 64 {
		t.Errorf("nil bytes hash length = %d, want 64", len(hashBytes))
	}

	// Empty bytes and empty string should be same
	hashEmptyBytes := HashBytes([]byte{})
	if hashEmptyBytes != hash {
		t.Error("empty []byte and empty string produce different hashes")
	}
}

func TestVerifyContentBlockHash(t *testing.T) {
	cb := &ContentBlock{
		ID:   "cb1",
		Text: "In the beginning God created the heaven and the earth.",
	}

	// Compute and set hash
	cb.Hash = HashContentBlock(cb)

	// Verify should succeed
	if !VerifyContentBlockHash(cb) {
		t.Error("VerifyContentBlockHash failed for valid hash")
	}

	// Modify text - verify should fail
	cb.Text = "Modified text"
	if VerifyContentBlockHash(cb) {
		t.Error("VerifyContentBlockHash succeeded for invalid hash")
	}

	// No hash set - verify should fail
	cb.Hash = ""
	if VerifyContentBlockHash(cb) {
		t.Error("VerifyContentBlockHash succeeded with no hash")
	}
}

func TestComputeAllHashes(t *testing.T) {
	corpus := &Corpus{
		ID:         "test",
		Version:    "1.0.0",
		ModuleType: ModuleBible,
		Documents: []*Document{
			{
				ID:    "Gen",
				Title: "Genesis",
				Order: 1,
				ContentBlocks: []*ContentBlock{
					{ID: "cb1", Text: "In the beginning"},
					{ID: "cb2", Text: "God created"},
				},
			},
		},
	}

	// Compute all hashes
	ComputeAllHashes(corpus)

	// Verify all content blocks have hashes
	for _, doc := range corpus.Documents {
		for _, cb := range doc.ContentBlocks {
			if cb.Hash == "" {
				t.Errorf("ContentBlock %s has no hash", cb.ID)
			}
			if !VerifyContentBlockHash(cb) {
				t.Errorf("ContentBlock %s has invalid hash", cb.ID)
			}
		}
	}
}

// TestVerifyAllHashes tests verifying all hashes in a corpus.
func TestVerifyAllHashes(t *testing.T) {
	corpus := &Corpus{
		ID:         "test",
		Version:    "1.0.0",
		ModuleType: ModuleBible,
		Documents: []*Document{
			{
				ID:    "Gen",
				Title: "Genesis",
				Order: 1,
				ContentBlocks: []*ContentBlock{
					{ID: "cb1", Text: "In the beginning"},
					{ID: "cb2", Text: "God created"},
				},
			},
		},
	}

	// Compute all hashes
	ComputeAllHashes(corpus)

	// Verify all - should pass
	invalid := VerifyAllHashes(corpus)
	if len(invalid) != 0 {
		t.Errorf("VerifyAllHashes returned invalid IDs for valid corpus: %v", invalid)
	}

	// Corrupt one hash
	corpus.Documents[0].ContentBlocks[0].Hash = "corrupted"

	// Verify - should detect invalid
	invalid = VerifyAllHashes(corpus)
	if len(invalid) != 1 {
		t.Errorf("VerifyAllHashes returned %d invalid, want 1", len(invalid))
	}
	if len(invalid) > 0 && invalid[0] != "cb1" {
		t.Errorf("VerifyAllHashes returned %q, want %q", invalid[0], "cb1")
	}
}

// TestVerifyAllHashesEmpty tests verifying hashes when none are set.
func TestVerifyAllHashesEmpty(t *testing.T) {
	corpus := &Corpus{
		ID:         "test",
		Version:    "1.0.0",
		ModuleType: ModuleBible,
		Documents: []*Document{
			{
				ID:    "Gen",
				Title: "Genesis",
				Order: 1,
				ContentBlocks: []*ContentBlock{
					{ID: "cb1", Text: "In the beginning"},
					{ID: "cb2", Text: "God created"},
				},
			},
		},
	}

	// Don't compute hashes - verify should report all as invalid
	invalid := VerifyAllHashes(corpus)
	if len(invalid) != 2 {
		t.Errorf("VerifyAllHashes returned %d invalid, want 2", len(invalid))
	}
}

// TestHashRef tests hashing a scripture reference.
func TestHashRef(t *testing.T) {
	ref := &Ref{
		Book:    "Gen",
		Chapter: 1,
		Verse:   1,
	}

	hash := HashRef(ref)

	// Should be 64 hex characters
	if len(hash) != 64 {
		t.Errorf("hash length = %d, want 64", len(hash))
	}

	// Same ref should produce same hash
	ref2 := &Ref{
		Book:    "Gen",
		Chapter: 1,
		Verse:   1,
	}
	hash2 := HashRef(ref2)
	if hash != hash2 {
		t.Errorf("same ref produced different hashes: %q vs %q", hash, hash2)
	}

	// Different ref should produce different hash
	ref3 := &Ref{
		Book:    "Gen",
		Chapter: 1,
		Verse:   2,
	}
	hash3 := HashRef(ref3)
	if hash == hash3 {
		t.Error("different refs produced same hash")
	}
}

// TestHashMappingTable tests hashing a mapping table.
func TestHashMappingTable(t *testing.T) {
	mt := &MappingTable{
		ID:         "mt1",
		FromSystem: VersificationKJV,
		ToSystem:   VersificationLXX,
		Mappings: []*RefMapping{
			{
				From: &Ref{Book: "Gen", Chapter: 1, Verse: 1},
				To:   &Ref{Book: "Gen", Chapter: 1, Verse: 1},
			},
		},
	}

	hash, err := HashMappingTable(mt)
	if err != nil {
		t.Fatalf("HashMappingTable failed: %v", err)
	}

	// Should be 64 hex characters
	if len(hash) != 64 {
		t.Errorf("hash length = %d, want 64", len(hash))
	}

	// Same table should produce same hash
	hash2, err := HashMappingTable(mt)
	if err != nil {
		t.Fatalf("HashMappingTable failed: %v", err)
	}
	if hash != hash2 {
		t.Errorf("same table produced different hashes")
	}

	// Modified table should produce different hash
	mt.Mappings[0].To.Verse = 2
	hash3, err := HashMappingTable(mt)
	if err != nil {
		t.Fatalf("HashMappingTable failed: %v", err)
	}
	if hash == hash3 {
		t.Error("modified table produced same hash")
	}
}

// TestHashCorpusMarshalError tests HashCorpus when json.Marshal fails.
// This can only happen if the Corpus struct contains unmarshalable types (channels, funcs).
// We use dependency injection to verify the error path works.
func TestHashCorpusMarshalError(t *testing.T) {
	orig := jsonMarshal
	defer func() { jsonMarshal = orig }()

	jsonMarshal = func(v any) ([]byte, error) {
		return nil, errors.New("injected marshal error")
	}

	corpus := &Corpus{ID: "test"}
	_, err := HashCorpus(corpus)
	if err == nil {
		t.Error("expected error when marshal fails")
	}
}

// TestHashDocumentMarshalError tests HashDocument when json.Marshal fails.
func TestHashDocumentMarshalError(t *testing.T) {
	orig := jsonMarshal
	defer func() { jsonMarshal = orig }()

	jsonMarshal = func(v any) ([]byte, error) {
		return nil, errors.New("injected marshal error")
	}

	doc := &Document{ID: "test"}
	_, err := HashDocument(doc)
	if err == nil {
		t.Error("expected error when marshal fails")
	}
}

// TestHashMappingTableMarshalError tests HashMappingTable when json.Marshal fails.
func TestHashMappingTableMarshalError(t *testing.T) {
	orig := jsonMarshal
	defer func() { jsonMarshal = orig }()

	jsonMarshal = func(v any) ([]byte, error) {
		return nil, errors.New("injected marshal error")
	}

	mt := &MappingTable{ID: "test"}
	_, err := HashMappingTable(mt)
	if err == nil {
		t.Error("expected error when marshal fails")
	}
}
