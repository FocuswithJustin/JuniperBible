package ir

import (
	"crypto/sha256"
	"encoding/hex"
)

// content.go - Content block utility functions
// Note: Type definitions are in types.go

// ComputeHash calculates and stores the SHA-256 hash of the Text field.
func (cb *ContentBlock) ComputeHash() string {
	h := sha256.Sum256([]byte(cb.Text))
	cb.Hash = hex.EncodeToString(h[:])
	return cb.Hash
}

// VerifyHash returns true if the stored hash matches the computed hash.
func (cb *ContentBlock) VerifyHash() bool {
	if cb.Hash == "" {
		return false
	}
	h := sha256.Sum256([]byte(cb.Text))
	return cb.Hash == hex.EncodeToString(h[:])
}

// Tokenize breaks text into tokens. This is a simple implementation
// that handles common English/Western text patterns.
func Tokenize(text string) []*Token {
	var tokens []*Token
	var tokenStart int
	var tokenText []byte
	var currentType TokenType
	index := 0

	finishToken := func(end int) {
		if len(tokenText) > 0 {
			tokens = append(tokens, &Token{
				ID:        "",
				Index:     index,
				CharStart: tokenStart,
				CharEnd:   end,
				Text:      string(tokenText),
				Type:      currentType,
			})
			index++
			tokenText = nil
		}
	}

	for i := 0; i < len(text); i++ {
		c := text[i]
		var newType TokenType

		switch {
		case c == ' ' || c == '\t' || c == '\n' || c == '\r':
			newType = TokenWhitespace
		case (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '\'' || c >= 0x80:
			// Letters, numbers, apostrophe, and non-ASCII (for Unicode words)
			newType = TokenWord
		default:
			newType = TokenPunctuation
		}

		if len(tokenText) == 0 {
			tokenStart = i
			currentType = newType
		} else if newType != currentType {
			finishToken(i)
			tokenStart = i
			currentType = newType
		}

		tokenText = append(tokenText, c)
	}

	finishToken(len(text))
	return tokens
}
