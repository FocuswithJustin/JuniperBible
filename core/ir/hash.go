package ir

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

// jsonMarshal is a variable to allow testing of marshal errors.
var jsonMarshal = json.Marshal

// HashBytes computes the SHA-256 hash of bytes and returns it as a hex string.
func HashBytes(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

// HashString computes the SHA-256 hash of a string and returns it as a hex string.
func HashString(s string) string {
	return HashBytes([]byte(s))
}

// HashCorpus computes the SHA-256 hash of a Corpus by serializing to JSON.
// This provides a content-addressable hash for the entire corpus.
func HashCorpus(c *Corpus) (string, error) {
	data, err := jsonMarshal(c)
	if err != nil {
		return "", err
	}
	return HashBytes(data), nil
}

// HashDocument computes the SHA-256 hash of a Document by serializing to JSON.
func HashDocument(d *Document) (string, error) {
	data, err := jsonMarshal(d)
	if err != nil {
		return "", err
	}
	return HashBytes(data), nil
}

// HashContentBlock computes the SHA-256 hash of a ContentBlock's text.
// This is used for content-addressing the text content.
func HashContentBlock(cb *ContentBlock) string {
	return HashString(cb.Text)
}

// VerifyContentBlockHash checks if the stored hash matches the computed hash.
func VerifyContentBlockHash(cb *ContentBlock) bool {
	if cb.Hash == "" {
		return false
	}
	return cb.Hash == HashContentBlock(cb)
}

// ComputeAllHashes computes and stores hashes for all content blocks in a corpus.
func ComputeAllHashes(c *Corpus) {
	for _, doc := range c.Documents {
		for _, cb := range doc.ContentBlocks {
			cb.Hash = HashContentBlock(cb)
		}
	}
}

// VerifyAllHashes verifies all content block hashes in a corpus.
// Returns the IDs of any content blocks with invalid hashes.
func VerifyAllHashes(c *Corpus) []string {
	var invalid []string
	for _, doc := range c.Documents {
		for _, cb := range doc.ContentBlocks {
			if !VerifyContentBlockHash(cb) {
				invalid = append(invalid, cb.ID)
			}
		}
	}
	return invalid
}

// HashRef computes a hash of a scripture reference for consistent comparison.
func HashRef(r *Ref) string {
	return HashString(r.String())
}

// HashMappingTable computes the SHA-256 hash of a MappingTable.
func HashMappingTable(mt *MappingTable) (string, error) {
	data, err := jsonMarshal(mt)
	if err != nil {
		return "", err
	}
	return HashBytes(data), nil
}
