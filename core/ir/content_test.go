package ir

import (
	"encoding/json"
	"testing"
)

func TestContentBlockJSON(t *testing.T) {
	cb := &ContentBlock{
		ID:       "cb1",
		Sequence: 0,
		Text:     "In the beginning God created the heaven and the earth.",
	}
	cb.ComputeHash()

	// Marshal to JSON
	data, err := json.Marshal(cb)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	// Unmarshal back
	var decoded ContentBlock
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	// Verify fields
	if decoded.ID != cb.ID {
		t.Errorf("ID = %q, want %q", decoded.ID, cb.ID)
	}
	if decoded.Sequence != cb.Sequence {
		t.Errorf("Sequence = %d, want %d", decoded.Sequence, cb.Sequence)
	}
	if decoded.Text != cb.Text {
		t.Errorf("Text = %q, want %q", decoded.Text, cb.Text)
	}
	if decoded.Hash != cb.Hash {
		t.Errorf("Hash = %q, want %q", decoded.Hash, cb.Hash)
	}
}

func TestContentBlockComputeHash(t *testing.T) {
	cb := &ContentBlock{
		ID:   "cb1",
		Text: "In the beginning God created the heaven and the earth.",
	}

	hash := cb.ComputeHash()

	// Hash should be a 64-character hex string (SHA-256)
	if len(hash) != 64 {
		t.Errorf("hash length = %d, want 64", len(hash))
	}

	// Same text should produce same hash
	cb2 := &ContentBlock{
		ID:   "cb2",
		Text: "In the beginning God created the heaven and the earth.",
	}
	hash2 := cb2.ComputeHash()
	if hash != hash2 {
		t.Errorf("same text produced different hashes: %q vs %q", hash, hash2)
	}

	// Different text should produce different hash
	cb3 := &ContentBlock{
		ID:   "cb3",
		Text: "Different text content.",
	}
	hash3 := cb3.ComputeHash()
	if hash == hash3 {
		t.Errorf("different text produced same hash")
	}
}

func TestContentBlockVerifyHash(t *testing.T) {
	cb := &ContentBlock{
		ID:   "cb1",
		Text: "In the beginning God created the heaven and the earth.",
	}

	// No hash set - should fail
	if cb.VerifyHash() {
		t.Error("VerifyHash returned true with no hash set")
	}

	// Compute hash
	cb.ComputeHash()

	// Valid hash - should pass
	if !cb.VerifyHash() {
		t.Error("VerifyHash returned false with valid hash")
	}

	// Modify text - should fail
	cb.Text = "Modified text"
	if cb.VerifyHash() {
		t.Error("VerifyHash returned true after text modification")
	}
}

func TestContentBlockWithTokens(t *testing.T) {
	cb := &ContentBlock{
		ID:       "cb1",
		Sequence: 0,
		Text:     "In the beginning",
		Tokens: []*Token{
			{ID: "t1", Index: 0, CharStart: 0, CharEnd: 2, Text: "In", Type: TokenWord},
			{ID: "t2", Index: 1, CharStart: 2, CharEnd: 3, Text: " ", Type: TokenWhitespace},
			{ID: "t3", Index: 2, CharStart: 3, CharEnd: 6, Text: "the", Type: TokenWord},
			{ID: "t4", Index: 3, CharStart: 6, CharEnd: 7, Text: " ", Type: TokenWhitespace},
			{ID: "t5", Index: 4, CharStart: 7, CharEnd: 16, Text: "beginning", Type: TokenWord},
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(cb)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	// Unmarshal back
	var decoded ContentBlock
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	// Verify tokens
	if len(decoded.Tokens) != 5 {
		t.Fatalf("len(Tokens) = %d, want 5", len(decoded.Tokens))
	}
	if decoded.Tokens[0].Text != "In" {
		t.Errorf("Tokens[0].Text = %q, want %q", decoded.Tokens[0].Text, "In")
	}
	if decoded.Tokens[4].Text != "beginning" {
		t.Errorf("Tokens[4].Text = %q, want %q", decoded.Tokens[4].Text, "beginning")
	}
}

func TestTokenJSON(t *testing.T) {
	token := &Token{
		ID:         "t1",
		Index:      0,
		CharStart:  0,
		CharEnd:    7,
		Text:       "ελοηιμ",
		Type:       TokenWord,
		Lemma:      "אֱלֹהִים",
		Strongs:    []string{"H430"},
		Morphology: "HNcmpc",
	}

	// Marshal to JSON
	data, err := json.Marshal(token)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	// Unmarshal back
	var decoded Token
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	// Verify fields
	if decoded.ID != token.ID {
		t.Errorf("ID = %q, want %q", decoded.ID, token.ID)
	}
	if decoded.Text != token.Text {
		t.Errorf("Text = %q, want %q", decoded.Text, token.Text)
	}
	if decoded.Lemma != token.Lemma {
		t.Errorf("Lemma = %q, want %q", decoded.Lemma, token.Lemma)
	}
	if len(decoded.Strongs) != 1 || decoded.Strongs[0] != "H430" {
		t.Errorf("Strongs = %v, want [H430]", decoded.Strongs)
	}
	if decoded.Morphology != token.Morphology {
		t.Errorf("Morphology = %q, want %q", decoded.Morphology, token.Morphology)
	}
}

func TestTokenIsWord(t *testing.T) {
	tests := []struct {
		token  *Token
		isWord bool
	}{
		{&Token{Type: TokenWord}, true},
		{&Token{Type: TokenWhitespace}, false},
		{&Token{Type: TokenPunctuation}, false},
	}

	for _, tt := range tests {
		if got := tt.token.IsWord(); got != tt.isWord {
			t.Errorf("Token{Type: %q}.IsWord() = %v, want %v", tt.token.Type, got, tt.isWord)
		}
	}
}

func TestTokenLength(t *testing.T) {
	token := &Token{
		CharStart: 10,
		CharEnd:   20,
	}

	if got := token.Length(); got != 10 {
		t.Errorf("Length() = %d, want 10", got)
	}
}

func TestTokenize(t *testing.T) {
	tests := []struct {
		text     string
		expected []struct {
			text      string
			tokenType TokenType
		}
	}{
		{
			text: "Hello world",
			expected: []struct {
				text      string
				tokenType TokenType
			}{
				{"Hello", TokenWord},
				{" ", TokenWhitespace},
				{"world", TokenWord},
			},
		},
		{
			text: "Hello, world!",
			expected: []struct {
				text      string
				tokenType TokenType
			}{
				{"Hello", TokenWord},
				{",", TokenPunctuation},
				{" ", TokenWhitespace},
				{"world", TokenWord},
				{"!", TokenPunctuation},
			},
		},
		{
			text: "In the beginning",
			expected: []struct {
				text      string
				tokenType TokenType
			}{
				{"In", TokenWord},
				{" ", TokenWhitespace},
				{"the", TokenWord},
				{" ", TokenWhitespace},
				{"beginning", TokenWord},
			},
		},
	}

	for _, tt := range tests {
		tokens := Tokenize(tt.text)
		if len(tokens) != len(tt.expected) {
			t.Errorf("Tokenize(%q) returned %d tokens, want %d", tt.text, len(tokens), len(tt.expected))
			continue
		}

		for i, exp := range tt.expected {
			if tokens[i].Text != exp.text {
				t.Errorf("Tokenize(%q)[%d].Text = %q, want %q", tt.text, i, tokens[i].Text, exp.text)
			}
			if tokens[i].Type != exp.tokenType {
				t.Errorf("Tokenize(%q)[%d].Type = %q, want %q", tt.text, i, tokens[i].Type, exp.tokenType)
			}
		}
	}
}

func TestTokenizeOffsets(t *testing.T) {
	text := "Hello world"
	tokens := Tokenize(text)

	// Verify offsets
	expected := []struct {
		start int
		end   int
	}{
		{0, 5},  // "Hello"
		{5, 6},  // " "
		{6, 11}, // "world"
	}

	for i, exp := range expected {
		if tokens[i].CharStart != exp.start {
			t.Errorf("tokens[%d].CharStart = %d, want %d", i, tokens[i].CharStart, exp.start)
		}
		if tokens[i].CharEnd != exp.end {
			t.Errorf("tokens[%d].CharEnd = %d, want %d", i, tokens[i].CharEnd, exp.end)
		}
	}

	// Verify text can be reconstructed
	var reconstructed string
	for _, tok := range tokens {
		reconstructed += tok.Text
	}
	if reconstructed != text {
		t.Errorf("reconstructed = %q, want %q", reconstructed, text)
	}
}

func TestTokenizeEmpty(t *testing.T) {
	tokens := Tokenize("")
	if len(tokens) != 0 {
		t.Errorf("Tokenize(\"\") returned %d tokens, want 0", len(tokens))
	}
}

func TestTokenizeUnicode(t *testing.T) {
	// Hebrew text
	text := "בְּרֵאשִׁית"
	tokens := Tokenize(text)

	if len(tokens) != 1 {
		t.Errorf("Tokenize(%q) returned %d tokens, want 1", text, len(tokens))
		return
	}

	if tokens[0].Type != TokenWord {
		t.Errorf("tokens[0].Type = %q, want %q", tokens[0].Type, TokenWord)
	}
	if tokens[0].Text != text {
		t.Errorf("tokens[0].Text = %q, want %q", tokens[0].Text, text)
	}
}

func TestContentBlockHashDeterminism(t *testing.T) {
	// Create same content multiple times and verify hash is deterministic
	text := "In the beginning God created the heaven and the earth."

	hashes := make([]string, 10)
	for i := 0; i < 10; i++ {
		cb := &ContentBlock{
			ID:   "test",
			Text: text,
		}
		hashes[i] = cb.ComputeHash()
	}

	// All hashes should be identical
	for i := 1; i < len(hashes); i++ {
		if hashes[i] != hashes[0] {
			t.Errorf("hash[%d] = %q, want %q (non-deterministic hashing)", i, hashes[i], hashes[0])
		}
	}
}
