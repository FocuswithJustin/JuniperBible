// Package usfm provides the embedded handler for USFM Bible format plugin.
package usfm

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"

	"github.com/FocuswithJustin/JuniperBible/core/ir"
)

// USFM parsing helpers
var (
	markerRegex   = regexp.MustCompile(`\\([a-zA-Z0-9]+)\*?(?:\s|$)`)
	verseNumRegex = regexp.MustCompile(`^(\d+)(?:-(\d+))?`)
	chapterRegex  = regexp.MustCompile(`^(\d+)`)
)

// Common USFM book IDs
var bookNames = map[string]string{
	"GEN": "Genesis", "EXO": "Exodus", "LEV": "Leviticus", "NUM": "Numbers",
	"DEU": "Deuteronomy", "JOS": "Joshua", "JDG": "Judges", "RUT": "Ruth",
	"1SA": "1 Samuel", "2SA": "2 Samuel", "1KI": "1 Kings", "2KI": "2 Kings",
	"1CH": "1 Chronicles", "2CH": "2 Chronicles", "EZR": "Ezra", "NEH": "Nehemiah",
	"EST": "Esther", "JOB": "Job", "PSA": "Psalms", "PRO": "Proverbs",
	"ECC": "Ecclesiastes", "SNG": "Song of Solomon", "ISA": "Isaiah", "JER": "Jeremiah",
	"LAM": "Lamentations", "EZK": "Ezekiel", "DAN": "Daniel", "HOS": "Hosea",
	"JOL": "Joel", "AMO": "Amos", "OBA": "Obadiah", "JON": "Jonah",
	"MIC": "Micah", "NAM": "Nahum", "HAB": "Habakkuk", "ZEP": "Zephaniah",
	"HAG": "Haggai", "ZEC": "Zechariah", "MAL": "Malachi",
	"MAT": "Matthew", "MRK": "Mark", "LUK": "Luke", "JHN": "John",
	"ACT": "Acts", "ROM": "Romans", "1CO": "1 Corinthians", "2CO": "2 Corinthians",
	"GAL": "Galatians", "EPH": "Ephesians", "PHP": "Philippians", "COL": "Colossians",
	"1TH": "1 Thessalonians", "2TH": "2 Thessalonians", "1TI": "1 Timothy", "2TI": "2 Timothy",
	"TIT": "Titus", "PHM": "Philemon", "HEB": "Hebrews", "JAS": "James",
	"1PE": "1 Peter", "2PE": "2 Peter", "1JN": "1 John", "2JN": "2 John",
	"3JN": "3 John", "JUD": "Jude", "REV": "Revelation",
}

// parseUSFMToIR converts USFM text to IR Corpus
func parseUSFMToIR(data []byte) (*ir.Corpus, error) {
	corpus := &ir.Corpus{
		Version:    "1.0.0",
		ModuleType: ir.ModuleBible,
		LossClass:  ir.LossL0,
		Documents:  []*ir.Document{},
	}

	var currentDoc *ir.Document
	var blockSeq int

	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Parse markers
		if strings.HasPrefix(trimmed, "\\") {
			parts := strings.SplitN(trimmed, " ", 2)
			marker := strings.TrimPrefix(parts[0], "\\")
			var value string
			if len(parts) > 1 {
				value = parts[1]
			}

			switch marker {
			case "id":
				// Book ID
				idParts := strings.Fields(value)
				if len(idParts) > 0 {
					bookID := strings.ToUpper(idParts[0])
					corpus.ID = bookID
					currentDoc = &ir.Document{
						ID:            bookID,
						Order:         len(corpus.Documents) + 1,
						ContentBlocks: []*ir.ContentBlock{},
					}
					if name, ok := bookNames[bookID]; ok {
						currentDoc.Title = name
					}
					corpus.Documents = append(corpus.Documents, currentDoc)
				}

			case "h", "toc1", "toc2", "toc3":
				// Header/TOC entries - stored in document attributes (not used in simplified version)
				if currentDoc != nil && value != "" {
					if marker == "h" && currentDoc.Title == "" {
						currentDoc.Title = value
					}
				}

			case "mt", "mt1", "mt2", "mt3":
				// Main title
				if corpus.Title == "" && value != "" {
					corpus.Title = value
				}

			case "c":
				// Chapter marker (parsing simplified for now)
				_ = value

			case "v":
				// Verse
				if currentDoc != nil {
					verseText := value

					// Parse verse number
					if matches := verseNumRegex.FindStringSubmatch(value); len(matches) > 0 {
						// Extract text after verse number (verse number parsing simplified for now)
						verseText = strings.TrimSpace(value[len(matches[0]):])
					}

					if verseText != "" {
						blockSeq++

						block := &ir.ContentBlock{
							ID:       fmt.Sprintf("cb-%d", blockSeq),
							Sequence: blockSeq,
							Text:     verseText,
							Anchors: []*ir.Anchor{
								{
									ID:             fmt.Sprintf("a-%d-0", blockSeq),
									ContentBlockID: fmt.Sprintf("cb-%d", blockSeq),
									CharOffset:     0,
								},
							},
						}

						// Compute hash
						block.ComputeHash()

						currentDoc.ContentBlocks = append(currentDoc.ContentBlocks, block)
					}
				}

			case "p", "m", "pi", "mi", "nb":
				// Paragraph markers - may contain text
				if currentDoc != nil && value != "" {
					blockSeq++
					block := &ir.ContentBlock{
						ID:       fmt.Sprintf("cb-%d", blockSeq),
						Sequence: blockSeq,
						Text:     value,
					}
					block.ComputeHash()
					currentDoc.ContentBlocks = append(currentDoc.ContentBlocks, block)
				}

			case "q", "q1", "q2", "q3", "qr", "qc", "qm":
				// Poetry markers
				if currentDoc != nil && value != "" {
					blockSeq++
					block := &ir.ContentBlock{
						ID:       fmt.Sprintf("cb-%d", blockSeq),
						Sequence: blockSeq,
						Text:     value,
					}
					block.ComputeHash()
					currentDoc.ContentBlocks = append(currentDoc.ContentBlocks, block)
				}
			}
		}
	}

	// Compute source hash
	h := sha256.Sum256(data)
	corpus.SourceHash = hex.EncodeToString(h[:])

	return corpus, nil
}

// emitUSFMFromIR converts IR Corpus back to USFM text
func emitUSFMFromIR(corpus *ir.Corpus) ([]byte, error) {
	var buf bytes.Buffer

	for _, doc := range corpus.Documents {
		// Write book ID
		buf.WriteString(fmt.Sprintf("\\id %s\n", doc.ID))

		// Write header
		if doc.Title != "" {
			buf.WriteString(fmt.Sprintf("\\h %s\n", doc.Title))
		}

		for _, block := range doc.ContentBlocks {
			// Check for verse spans to determine chapter/verse
			hasVerse := false
			for _, anchor := range block.Anchors {
				// Simple heuristic: if we have anchors at start, assume verse content
				if anchor.CharOffset == 0 {
					// Try to extract chapter/verse from block ID or infer from content
					// For now, just mark as having verse
					hasVerse = true
					break
				}
			}

			if hasVerse {
				// Try to infer chapter/verse from text or use simple increment
				// This is simplified - in real implementation we'd need better tracking
				buf.WriteString(fmt.Sprintf("\\v %s\n", block.Text))
			} else {
				// Non-verse block
				buf.WriteString(fmt.Sprintf("\\p %s\n", block.Text))
			}
		}
	}

	return buf.Bytes(), nil
}
