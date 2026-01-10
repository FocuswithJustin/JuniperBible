// Package encoding provides shared text encoding and escaping utilities.
package encoding

import (
	"bytes"
	"encoding/xml"
	"strings"
)

// EscapeXML escapes special characters for XML content.
// Uses the standard library's xml.EscapeText for proper escaping.
func EscapeXML(s string) string {
	var buf bytes.Buffer
	xml.EscapeText(&buf, []byte(s))
	return buf.String()
}

// EscapeXMLText escapes only the basic XML entities for text content.
// This is a lighter-weight alternative to EscapeXML.
func EscapeXMLText(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

// EscapeXMLAttr escapes text for use in XML attributes.
// Includes quote escaping in addition to basic XML entities.
func EscapeXMLAttr(s string) string {
	s = EscapeXMLText(s)
	s = strings.ReplaceAll(s, "\"", "&quot;")
	return s
}

// EscapeHTML escapes special characters for HTML content.
// Escapes: & < > "
func EscapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	return s
}

// EscapeLaTeX escapes special characters for LaTeX documents.
// Escapes: \ { } $ % & # _ ^ ~
func EscapeLaTeX(s string) string {
	// Use placeholder for backslash to avoid re-escaping braces in \textbackslash{}
	const placeholder = "\x00BACKSLASH\x00"
	s = strings.ReplaceAll(s, "\\", placeholder)

	// Escape all other special characters
	replacements := []struct {
		old, new string
	}{
		{"{", "\\{"},
		{"}", "\\}"},
		{"$", "\\$"},
		{"%", "\\%"},
		{"&", "\\&"},
		{"#", "\\#"},
		{"_", "\\_"},
		{"^", "\\^{}"},
		{"~", "\\~{}"},
	}

	for _, r := range replacements {
		s = strings.ReplaceAll(s, r.old, r.new)
	}

	// Replace placeholder with final backslash escape
	s = strings.ReplaceAll(s, placeholder, "\\textbackslash{}")
	return s
}

// EscapeManifest escapes text for use in MANIFEST.MF files (J2ME).
// Replaces newlines with spaces to ensure single-line values.
func EscapeManifest(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	return s
}
