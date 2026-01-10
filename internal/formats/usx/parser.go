// Package usx provides the embedded handler for USX Bible format plugin.
package usx

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/FocuswithJustin/JuniperBible/core/ir"
)

// USX XML types
type USX struct {
	XMLName xml.Name  `xml:"usx"`
	Version string    `xml:"version,attr"`
	Book    *USXBook  `xml:"book"`
	Content []USXNode `xml:",any"`
}

type USXBook struct {
	XMLName xml.Name `xml:"book"`
	Code    string   `xml:"code,attr"`
	Style   string   `xml:"style,attr"`
	Content string   `xml:",chardata"`
}

type USXNode struct {
	XMLName xml.Name
	Attrs   []xml.Attr `xml:",any,attr"`
	Content string     `xml:",chardata"`
	Nodes   []USXNode  `xml:",any"`
}

func parseUSXToIR(data []byte) (*ir.Corpus, error) {
	decoder := xml.NewDecoder(bytes.NewReader(data))

	corpus := &ir.Corpus{
		Version:    "1.0.0",
		ModuleType: ir.ModuleBible,
		LossClass:  ir.LossL0,
		Documents:  []*ir.Document{},
	}

	var doc *ir.Document
	var currentChapter, currentVerse int
	sequence := 0
	var textBuf strings.Builder

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		switch t := token.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "usx":
				for _, attr := range t.Attr {
					if attr.Name.Local == "version" {
						// Store version in corpus (simplified - not using attributes)
					}
				}

			case "book":
				for _, attr := range t.Attr {
					if attr.Name.Local == "code" {
						corpus.ID = attr.Value
						doc = &ir.Document{
							ID:            attr.Value,
							Title:         attr.Value,
							Order:         1,
							ContentBlocks: []*ir.ContentBlock{},
						}
					}
				}

			case "chapter":
				// Flush any pending text
				if textBuf.Len() > 0 && currentVerse > 0 {
					sequence++
					cb := createContentBlock(sequence, textBuf.String(), doc.ID, currentChapter, currentVerse)
					doc.ContentBlocks = append(doc.ContentBlocks, cb)
					textBuf.Reset()
				}

				for _, attr := range t.Attr {
					if attr.Name.Local == "number" {
						currentChapter, _ = strconv.Atoi(attr.Value)
						currentVerse = 0
					}
				}

			case "verse":
				// Flush previous verse
				if textBuf.Len() > 0 && currentVerse > 0 {
					sequence++
					cb := createContentBlock(sequence, textBuf.String(), doc.ID, currentChapter, currentVerse)
					doc.ContentBlocks = append(doc.ContentBlocks, cb)
					textBuf.Reset()
				}

				for _, attr := range t.Attr {
					if attr.Name.Local == "number" {
						currentVerse, _ = strconv.Atoi(attr.Value)
					}
				}

			case "para":
				// Handle paragraph styles - header content captured in text
			}

		case xml.CharData:
			text := strings.TrimSpace(string(t))
			if text != "" && currentVerse > 0 {
				if textBuf.Len() > 0 {
					textBuf.WriteString(" ")
				}
				textBuf.WriteString(text)
			}
		}
	}

	// Flush final verse
	if textBuf.Len() > 0 && currentVerse > 0 && doc != nil {
		sequence++
		cb := createContentBlock(sequence, textBuf.String(), doc.ID, currentChapter, currentVerse)
		doc.ContentBlocks = append(doc.ContentBlocks, cb)
	}

	if doc != nil {
		corpus.Documents = []*ir.Document{doc}
		corpus.Title = doc.Title
	}

	// Compute source hash
	h := sha256.Sum256(data)
	corpus.SourceHash = hex.EncodeToString(h[:])

	return corpus, nil
}

func createContentBlock(sequence int, text, book string, chapter, verse int) *ir.ContentBlock {
	text = strings.TrimSpace(text)

	block := &ir.ContentBlock{
		ID:       fmt.Sprintf("cb-%d", sequence),
		Sequence: sequence,
		Text:     text,
		Anchors: []*ir.Anchor{
			{
				ID:             fmt.Sprintf("a-%d-0", sequence),
				ContentBlockID: fmt.Sprintf("cb-%d", sequence),
				CharOffset:     0,
			},
		},
	}

	block.ComputeHash()
	return block
}

func emitUSXFromIR(corpus *ir.Corpus) string {
	var buf strings.Builder

	version := "3.0"

	buf.WriteString(fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<usx version="%s">
`, version))

	for _, doc := range corpus.Documents {
		buf.WriteString(fmt.Sprintf(`  <book code="%s" style="id">%s</book>
`, doc.ID, doc.Title))

		for _, cb := range doc.ContentBlocks {
			// Simple heuristic: extract chapter/verse from anchors or infer from sequence
			// This is simplified - in real implementation we'd track chapter/verse properly
			if len(cb.Anchors) > 0 {
				// For now, just write as paragraph with verse markers
				// In full implementation, we'd parse the OSIS ID or track chapter/verse
				buf.WriteString(fmt.Sprintf(`  <para style="p">%s</para>
`, escapeXML(cb.Text)))
			}
		}
	}

	buf.WriteString("</usx>\n")
	return buf.String()
}

func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	return s
}
