// Package rtf provides pure Go RTF parsing and conversion.
// This replaces external unrtf dependency with native Go implementation.
package rtf

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/FocuswithJustin/JuniperBible/core/encoding"
)

// Document represents a parsed RTF document.
type Document struct {
	root     *Group
	metadata DocumentMetadata
}

// DocumentMetadata contains RTF document metadata.
type DocumentMetadata struct {
	Title   string
	Author  string
	Subject string
	Created string
}

// Group represents an RTF group (content within braces).
type Group struct {
	children []interface{} // can be *Group, ControlWord, or string
}

// ControlWord represents an RTF control word.
type ControlWord struct {
	word  string
	param int
	has   bool
}

// Parse parses RTF data and returns a Document.
func Parse(data []byte) (*Document, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty RTF data")
	}

	// Check for RTF header
	if !bytes.HasPrefix(data, []byte("{\\rtf")) {
		return nil, fmt.Errorf("not a valid RTF document: missing \\rtf header")
	}

	parser := &rtfParser{data: data, pos: 0}
	root, err := parser.parseGroup()
	if err != nil {
		return nil, err
	}

	doc := &Document{root: root}
	doc.extractMetadata()

	return doc, nil
}

type rtfParser struct {
	data []byte
	pos  int
}

func (p *rtfParser) parseGroup() (*Group, error) {
	if p.pos >= len(p.data) || p.data[p.pos] != '{' {
		return nil, fmt.Errorf("expected '{' at position %d", p.pos)
	}
	p.pos++ // consume '{'

	group := &Group{}

	for p.pos < len(p.data) {
		ch := p.data[p.pos]

		switch ch {
		case '}':
			p.pos++ // consume '}'
			return group, nil

		case '{':
			// Nested group
			nested, err := p.parseGroup()
			if err != nil {
				return nil, err
			}
			group.children = append(group.children, nested)

		case '\\':
			// Control word or symbol
			cw, err := p.parseControlWord()
			if err != nil {
				return nil, err
			}
			group.children = append(group.children, cw)

		case '\r', '\n':
			p.pos++

		default:
			// Text content
			text := p.parseText()
			if text != "" {
				group.children = append(group.children, text)
			}
		}
	}

	return nil, fmt.Errorf("unclosed group")
}

func (p *rtfParser) parseControlWord() (ControlWord, error) {
	p.pos++ // consume '\'
	if p.pos >= len(p.data) {
		return ControlWord{}, fmt.Errorf("unexpected end after backslash")
	}

	ch := p.data[p.pos]

	// Special characters: \{, \}, \\
	if ch == '{' || ch == '}' || ch == '\\' {
		p.pos++
		return ControlWord{word: string(ch)}, nil
	}

	// Control word
	if isLetter(ch) {
		start := p.pos
		for p.pos < len(p.data) && isLetter(p.data[p.pos]) {
			p.pos++
		}
		word := string(p.data[start:p.pos])

		// Check for numeric parameter
		var param int
		var hasParam bool
		if p.pos < len(p.data) && (p.data[p.pos] == '-' || isDigit(p.data[p.pos])) {
			numStart := p.pos
			if p.data[p.pos] == '-' {
				p.pos++
			}
			for p.pos < len(p.data) && isDigit(p.data[p.pos]) {
				p.pos++
			}
			param, _ = strconv.Atoi(string(p.data[numStart:p.pos]))
			hasParam = true
		}

		// Skip delimiter space
		if p.pos < len(p.data) && p.data[p.pos] == ' ' {
			p.pos++
		}

		return ControlWord{word: word, param: param, has: hasParam}, nil
	}

	// Control symbol (single non-letter character after \)
	p.pos++
	return ControlWord{word: string(ch)}, nil
}

func (p *rtfParser) parseText() string {
	var buf bytes.Buffer
	for p.pos < len(p.data) {
		ch := p.data[p.pos]
		if ch == '{' || ch == '}' || ch == '\\' {
			break
		}
		if ch != '\r' && ch != '\n' {
			buf.WriteByte(ch)
		}
		p.pos++
	}
	return buf.String()
}

func isLetter(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}

func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}

// extractMetadata extracts document metadata from info group.
func (doc *Document) extractMetadata() {
	if doc.root == nil {
		return
	}

	for _, child := range doc.root.children {
		if group, ok := child.(*Group); ok {
			doc.findInfoGroup(group)
		}
	}
}

func (doc *Document) findInfoGroup(group *Group) {
	for _, child := range group.children {
		if cw, ok := child.(ControlWord); ok {
			if cw.word == "info" {
				doc.parseInfoGroup(group)
				return
			}
		}
		if nested, ok := child.(*Group); ok {
			// Check if this is an info group
			for _, c := range nested.children {
				if cw, ok := c.(ControlWord); ok && cw.word == "info" {
					doc.parseInfoGroup(nested)
					return
				}
			}
			doc.findInfoGroup(nested)
		}
	}
}

func (doc *Document) parseInfoGroup(group *Group) {
	for _, child := range group.children {
		if nested, ok := child.(*Group); ok {
			var fieldName string
			var fieldValue strings.Builder

			for _, c := range nested.children {
				if cw, ok := c.(ControlWord); ok {
					switch cw.word {
					case "title":
						fieldName = "title"
					case "author":
						fieldName = "author"
					case "subject":
						fieldName = "subject"
					}
				}
				if text, ok := c.(string); ok {
					fieldValue.WriteString(text)
				}
			}

			switch fieldName {
			case "title":
				doc.metadata.Title = strings.TrimSpace(fieldValue.String())
			case "author":
				doc.metadata.Author = strings.TrimSpace(fieldValue.String())
			case "subject":
				doc.metadata.Subject = strings.TrimSpace(fieldValue.String())
			}
		}
	}
}

// ToText converts the document to plain text.
func (doc *Document) ToText() string {
	if doc.root == nil {
		return ""
	}
	var buf strings.Builder
	doc.extractText(doc.root, &buf)
	return strings.TrimSpace(buf.String())
}

// ToTextBytes converts the document to plain text bytes.
func (doc *Document) ToTextBytes() []byte {
	return []byte(doc.ToText())
}

func (doc *Document) extractText(group *Group, buf *strings.Builder) {
	for _, child := range group.children {
		switch v := child.(type) {
		case string:
			buf.WriteString(v)
		case ControlWord:
			switch v.word {
			case "par", "line":
				buf.WriteString("\n")
			case "tab":
				buf.WriteString("\t")
			case "{":
				buf.WriteString("{")
			case "}":
				buf.WriteString("}")
			case "\\":
				buf.WriteString("\\")
			case "u":
				// Unicode escape
				if v.has && v.param > 0 {
					buf.WriteRune(rune(v.param))
				}
			}
		case *Group:
			// Skip special groups
			isSpecialGroup := false
			for _, c := range v.children {
				if cw, ok := c.(ControlWord); ok {
					switch cw.word {
					case "fonttbl", "colortbl", "stylesheet", "info", "pict", "object":
						isSpecialGroup = true
					}
				}
			}
			if !isSpecialGroup {
				doc.extractText(v, buf)
			}
		}
	}
}

// ToHTML converts the document to HTML.
func (doc *Document) ToHTML() string {
	if doc.root == nil {
		return ""
	}
	var buf strings.Builder
	buf.WriteString("<!DOCTYPE html>\n<html>\n<head>\n")
	if doc.metadata.Title != "" {
		buf.WriteString("<title>")
		buf.WriteString(encoding.EscapeHTML(doc.metadata.Title))
		buf.WriteString("</title>\n")
	}
	buf.WriteString("</head>\n<body>\n")

	doc.extractHTML(doc.root, &buf, false, false)

	buf.WriteString("\n</body>\n</html>")
	return buf.String()
}

// ToHTMLBytes converts the document to HTML bytes.
func (doc *Document) ToHTMLBytes() []byte {
	return []byte(doc.ToHTML())
}

func (doc *Document) extractHTML(group *Group, buf *strings.Builder, inBold, inItalic bool) {
	localBold := inBold
	localItalic := inItalic
	inParagraph := false

	for _, child := range group.children {
		switch v := child.(type) {
		case string:
			if !inParagraph && strings.TrimSpace(v) != "" {
				buf.WriteString("<p>")
				inParagraph = true
			}
			buf.WriteString(encoding.EscapeHTML(v))

		case ControlWord:
			switch v.word {
			case "b":
				if !localBold && (v.param != 0 || !v.has) {
					buf.WriteString("<b>")
					localBold = true
				} else if localBold && v.param == 0 {
					buf.WriteString("</b>")
					localBold = false
				}
			case "i":
				if !localItalic && (v.param != 0 || !v.has) {
					buf.WriteString("<i>")
					localItalic = true
				} else if localItalic && v.param == 0 {
					buf.WriteString("</i>")
					localItalic = false
				}
			case "par":
				if inParagraph {
					buf.WriteString("</p>\n<p>")
				} else {
					buf.WriteString("<p>")
					inParagraph = true
				}
			case "line":
				buf.WriteString("<br/>")
			case "tab":
				buf.WriteString("&nbsp;&nbsp;&nbsp;&nbsp;")
			case "u":
				if v.has && v.param > 0 {
					buf.WriteRune(rune(v.param))
				}
			}

		case *Group:
			// Check for special formatting groups
			groupBold := localBold
			groupItalic := localItalic

			for _, c := range v.children {
				if cw, ok := c.(ControlWord); ok {
					switch cw.word {
					case "b":
						if cw.param != 0 || !cw.has {
							groupBold = true
						}
					case "i":
						if cw.param != 0 || !cw.has {
							groupItalic = true
						}
					case "fonttbl", "colortbl", "stylesheet", "info", "pict", "object":
						goto skipGroup
					}
				}
			}

			if groupBold && !localBold {
				buf.WriteString("<b>")
			}
			if groupItalic && !localItalic {
				buf.WriteString("<i>")
			}

			doc.extractHTML(v, buf, groupBold, groupItalic)

			if groupItalic && !localItalic {
				buf.WriteString("</i>")
			}
			if groupBold && !localBold {
				buf.WriteString("</b>")
			}
			continue

		skipGroup:
		}
	}

	if inParagraph {
		buf.WriteString("</p>")
	}
}

// ToLaTeX converts the document to LaTeX.
func (doc *Document) ToLaTeX() string {
	if doc.root == nil {
		return ""
	}
	var buf strings.Builder
	buf.WriteString("\\documentclass{article}\n")
	buf.WriteString("\\usepackage[utf8]{inputenc}\n")
	if doc.metadata.Title != "" {
		buf.WriteString("\\title{")
		buf.WriteString(encoding.EscapeLaTeX(doc.metadata.Title))
		buf.WriteString("}\n")
	}
	if doc.metadata.Author != "" {
		buf.WriteString("\\author{")
		buf.WriteString(encoding.EscapeLaTeX(doc.metadata.Author))
		buf.WriteString("}\n")
	}
	buf.WriteString("\\begin{document}\n")
	if doc.metadata.Title != "" {
		buf.WriteString("\\maketitle\n")
	}

	doc.extractLaTeX(doc.root, &buf, false, false)

	buf.WriteString("\n\\end{document}\n")
	return buf.String()
}

// ToLaTeXBytes converts the document to LaTeX bytes.
func (doc *Document) ToLaTeXBytes() []byte {
	return []byte(doc.ToLaTeX())
}

func (doc *Document) extractLaTeX(group *Group, buf *strings.Builder, inBold, inItalic bool) {
	for _, child := range group.children {
		switch v := child.(type) {
		case string:
			buf.WriteString(encoding.EscapeLaTeX(v))

		case ControlWord:
			switch v.word {
			case "par":
				buf.WriteString("\n\n")
			case "line":
				buf.WriteString("\\\\\n")
			case "tab":
				buf.WriteString("\\quad ")
			case "u":
				if v.has && v.param > 0 {
					r := rune(v.param)
					if r < 128 && unicode.IsPrint(r) {
						buf.WriteRune(r)
					} else {
						buf.WriteString(fmt.Sprintf("\\symbol{%d}", v.param))
					}
				}
			}

		case *Group:
			// Check for formatting
			groupBold := false
			groupItalic := false

			for _, c := range v.children {
				if cw, ok := c.(ControlWord); ok {
					switch cw.word {
					case "b":
						if cw.param != 0 || !cw.has {
							groupBold = true
						}
					case "i":
						if cw.param != 0 || !cw.has {
							groupItalic = true
						}
					case "fonttbl", "colortbl", "stylesheet", "info", "pict", "object":
						goto skipGroup
					}
				}
			}

			if groupBold {
				buf.WriteString("\\textbf{")
			}
			if groupItalic {
				buf.WriteString("\\textit{")
			}

			doc.extractLaTeX(v, buf, groupBold, groupItalic)

			if groupItalic {
				buf.WriteString("}")
			}
			if groupBold {
				buf.WriteString("}")
			}
			continue

		skipGroup:
		}
	}
}

// Metadata returns the document metadata.
func (doc *Document) Metadata() DocumentMetadata {
	return doc.metadata
}
