// Package rtf provides pure Go RTF parsing and conversion.
package rtf

import (
	"strings"
	"testing"
)

// TestParseValidRTF verifies parsing of well-formed RTF.
func TestParseValidRTF(t *testing.T) {
	rtfData := `{\rtf1\ansi Hello World}`

	doc, err := Parse([]byte(rtfData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if doc == nil {
		t.Fatal("Parse returned nil document")
	}
}

// TestParseEmptyRTF verifies handling of empty RTF.
func TestParseEmptyRTF(t *testing.T) {
	rtfData := `{\rtf1}`

	doc, err := Parse([]byte(rtfData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if doc == nil {
		t.Fatal("Parse returned nil document")
	}
}

// TestParseInvalidRTF verifies error handling for malformed RTF.
func TestParseInvalidRTF(t *testing.T) {
	tests := []struct {
		name string
		rtf  string
	}{
		{"unclosed brace", `{\rtf1 Hello`},
		{"no rtf header", `{Hello World}`},
		{"empty", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse([]byte(tt.rtf))
			if err == nil {
				t.Error("Parse should fail for invalid RTF")
			}
		})
	}
}

// TestToText verifies RTF to plain text conversion.
func TestToText(t *testing.T) {
	rtfData := `{\rtf1\ansi Hello World}`

	doc, err := Parse([]byte(rtfData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	text := doc.ToText()
	if !strings.Contains(text, "Hello World") {
		t.Errorf("ToText should contain 'Hello World': %q", text)
	}
}

// TestToTextWithFormatting verifies text extraction strips formatting.
func TestToTextWithFormatting(t *testing.T) {
	rtfData := `{\rtf1\ansi {\b Bold} and {\i Italic} text}`

	doc, err := Parse([]byte(rtfData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	text := doc.ToText()
	if !strings.Contains(text, "Bold") {
		t.Error("ToText should contain 'Bold'")
	}
	if !strings.Contains(text, "Italic") {
		t.Error("ToText should contain 'Italic'")
	}
}

// TestToTextWithNewlines verifies paragraph handling.
func TestToTextWithNewlines(t *testing.T) {
	rtfData := `{\rtf1\ansi First\par Second\par Third}`

	doc, err := Parse([]byte(rtfData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	text := doc.ToText()
	if !strings.Contains(text, "First") {
		t.Error("ToText should contain 'First'")
	}
	if !strings.Contains(text, "Second") {
		t.Error("ToText should contain 'Second'")
	}
}

// TestToHTML verifies RTF to HTML conversion.
func TestToHTML(t *testing.T) {
	rtfData := `{\rtf1\ansi Hello World}`

	doc, err := Parse([]byte(rtfData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	html := doc.ToHTML()
	if !strings.Contains(html, "Hello World") {
		t.Errorf("ToHTML should contain 'Hello World': %q", html)
	}
	if !strings.Contains(html, "<") {
		t.Error("ToHTML should contain HTML tags")
	}
}

// TestToHTMLWithBold verifies bold text conversion.
func TestToHTMLWithBold(t *testing.T) {
	rtfData := `{\rtf1\ansi {\b Bold text}}`

	doc, err := Parse([]byte(rtfData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	html := doc.ToHTML()
	if !strings.Contains(html, "<b>") || !strings.Contains(html, "</b>") {
		t.Errorf("ToHTML should contain bold tags: %q", html)
	}
}

// TestToHTMLWithItalic verifies italic text conversion.
func TestToHTMLWithItalic(t *testing.T) {
	rtfData := `{\rtf1\ansi {\i Italic text}}`

	doc, err := Parse([]byte(rtfData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	html := doc.ToHTML()
	if !strings.Contains(html, "<i>") || !strings.Contains(html, "</i>") {
		t.Errorf("ToHTML should contain italic tags: %q", html)
	}
}

// TestToHTMLWithParagraphs verifies paragraph conversion.
func TestToHTMLWithParagraphs(t *testing.T) {
	rtfData := `{\rtf1\ansi First paragraph\par Second paragraph}`

	doc, err := Parse([]byte(rtfData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	html := doc.ToHTML()
	if !strings.Contains(html, "<p>") {
		t.Errorf("ToHTML should contain paragraph tags: %q", html)
	}
}

// TestToLaTeX verifies RTF to LaTeX conversion.
func TestToLaTeX(t *testing.T) {
	rtfData := `{\rtf1\ansi Hello World}`

	doc, err := Parse([]byte(rtfData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	latex := doc.ToLaTeX()
	if !strings.Contains(latex, "Hello World") {
		t.Errorf("ToLaTeX should contain 'Hello World': %q", latex)
	}
}

// TestToLaTeXWithBold verifies bold text in LaTeX.
func TestToLaTeXWithBold(t *testing.T) {
	rtfData := `{\rtf1\ansi {\b Bold text}}`

	doc, err := Parse([]byte(rtfData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	latex := doc.ToLaTeX()
	if !strings.Contains(latex, "\\textbf{") {
		t.Errorf("ToLaTeX should contain \\textbf: %q", latex)
	}
}

// TestToLaTeXWithItalic verifies italic text in LaTeX.
func TestToLaTeXWithItalic(t *testing.T) {
	rtfData := `{\rtf1\ansi {\i Italic text}}`

	doc, err := Parse([]byte(rtfData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	latex := doc.ToLaTeX()
	if !strings.Contains(latex, "\\textit{") {
		t.Errorf("ToLaTeX should contain \\textit: %q", latex)
	}
}

// TestUnicodeHandling verifies Unicode character handling.
func TestUnicodeHandling(t *testing.T) {
	// RTF Unicode escape: \u followed by decimal code point
	rtfData := `{\rtf1\ansi Hello \u8212 World}` // em-dash

	doc, err := Parse([]byte(rtfData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	text := doc.ToText()
	if !strings.Contains(text, "Hello") || !strings.Contains(text, "World") {
		t.Errorf("ToText should handle Unicode: %q", text)
	}
}

// TestSpecialCharacters verifies special character handling.
func TestSpecialCharacters(t *testing.T) {
	rtfData := `{\rtf1\ansi \{ braces \} and \\ backslash}`

	doc, err := Parse([]byte(rtfData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	text := doc.ToText()
	if !strings.Contains(text, "{") {
		t.Error("ToText should contain literal brace")
	}
	if !strings.Contains(text, "\\") {
		t.Error("ToText should contain literal backslash")
	}
}

// TestColorTable verifies color table parsing (but not application).
func TestColorTable(t *testing.T) {
	rtfData := `{\rtf1\ansi{\colortbl;\red255\green0\blue0;}Text}`

	doc, err := Parse([]byte(rtfData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	text := doc.ToText()
	if !strings.Contains(text, "Text") {
		t.Errorf("ToText should extract text despite color table: %q", text)
	}
}

// TestFontTable verifies font table parsing (but not application).
func TestFontTable(t *testing.T) {
	rtfData := `{\rtf1\ansi{\fonttbl{\f0 Arial;}}Hello}`

	doc, err := Parse([]byte(rtfData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	text := doc.ToText()
	if !strings.Contains(text, "Hello") {
		t.Errorf("ToText should extract text despite font table: %q", text)
	}
}

// TestNestedGroups verifies nested group handling.
func TestNestedGroups(t *testing.T) {
	rtfData := `{\rtf1\ansi {Outer {Inner} Outer}}`

	doc, err := Parse([]byte(rtfData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	text := doc.ToText()
	if !strings.Contains(text, "Outer") {
		t.Error("ToText should contain 'Outer'")
	}
	if !strings.Contains(text, "Inner") {
		t.Error("ToText should contain 'Inner'")
	}
}

// TestLineBreak verifies line break handling.
func TestLineBreak(t *testing.T) {
	rtfData := `{\rtf1\ansi Line1\line Line2}`

	doc, err := Parse([]byte(rtfData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	text := doc.ToText()
	if !strings.Contains(text, "Line1") || !strings.Contains(text, "Line2") {
		t.Errorf("ToText should contain both lines: %q", text)
	}
}

// TestTab verifies tab handling.
func TestTab(t *testing.T) {
	rtfData := `{\rtf1\ansi Column1\tab Column2}`

	doc, err := Parse([]byte(rtfData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	text := doc.ToText()
	if !strings.Contains(text, "Column1") || !strings.Contains(text, "Column2") {
		t.Errorf("ToText should contain both columns: %q", text)
	}
}

// TestMetadata verifies metadata extraction.
func TestMetadata(t *testing.T) {
	rtfData := `{\rtf1\ansi{\info{\title Test Document}{\author John Doe}}Content}`

	doc, err := Parse([]byte(rtfData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	meta := doc.Metadata()
	if meta.Title != "Test Document" {
		t.Errorf("Title = %q, want %q", meta.Title, "Test Document")
	}
	if meta.Author != "John Doe" {
		t.Errorf("Author = %q, want %q", meta.Author, "John Doe")
	}
}

// TestConvertToBytes verifies byte output methods.
func TestConvertToBytes(t *testing.T) {
	rtfData := `{\rtf1\ansi Hello}`

	doc, err := Parse([]byte(rtfData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	textBytes := doc.ToTextBytes()
	if len(textBytes) == 0 {
		t.Error("ToTextBytes should return non-empty slice")
	}

	htmlBytes := doc.ToHTMLBytes()
	if len(htmlBytes) == 0 {
		t.Error("ToHTMLBytes should return non-empty slice")
	}

	latexBytes := doc.ToLaTeXBytes()
	if len(latexBytes) == 0 {
		t.Error("ToLaTeXBytes should return non-empty slice")
	}
}

// TestNilDocument verifies handling of nil document root.
func TestNilDocument(t *testing.T) {
	doc := &Document{root: nil}

	text := doc.ToText()
	if text != "" {
		t.Errorf("ToText with nil root should return empty string, got %q", text)
	}

	html := doc.ToHTML()
	if html != "" {
		t.Errorf("ToHTML with nil root should return empty string, got %q", html)
	}

	latex := doc.ToLaTeX()
	if latex != "" {
		t.Errorf("ToLaTeX with nil root should return empty string, got %q", latex)
	}

	// extractMetadata with nil root should not panic
	doc.extractMetadata()
	meta := doc.Metadata()
	if meta.Title != "" || meta.Author != "" || meta.Subject != "" {
		t.Error("Metadata should be empty for nil root")
	}
}

// TestMetadataWithSubject verifies subject metadata extraction.
func TestMetadataWithSubject(t *testing.T) {
	rtfData := `{\rtf1\ansi{\info{\subject Test Subject}}Content}`

	doc, err := Parse([]byte(rtfData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	meta := doc.Metadata()
	if meta.Subject != "Test Subject" {
		t.Errorf("Subject = %q, want %q", meta.Subject, "Test Subject")
	}
}

// TestMetadataComplete verifies all metadata fields.
func TestMetadataComplete(t *testing.T) {
	rtfData := `{\rtf1\ansi{\info{\title My Title}{\author Jane Smith}{\subject Important}}Content}`

	doc, err := Parse([]byte(rtfData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	meta := doc.Metadata()
	if meta.Title != "My Title" {
		t.Errorf("Title = %q, want %q", meta.Title, "My Title")
	}
	if meta.Author != "Jane Smith" {
		t.Errorf("Author = %q, want %q", meta.Author, "Jane Smith")
	}
	if meta.Subject != "Important" {
		t.Errorf("Subject = %q, want %q", meta.Subject, "Important")
	}
}

// TestHTMLWithTitle verifies HTML generation with title metadata.
func TestHTMLWithTitle(t *testing.T) {
	rtfData := `{\rtf1\ansi{\info{\title My Page}}Content}`

	doc, err := Parse([]byte(rtfData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	html := doc.ToHTML()
	if !strings.Contains(html, "<title>My Page</title>") {
		t.Errorf("ToHTML should contain title tag: %q", html)
	}
}

// TestHTMLEscaping verifies HTML special character escaping.
func TestHTMLEscaping(t *testing.T) {
	rtfData := `{\rtf1\ansi <tag> & "quoted"}`

	doc, err := Parse([]byte(rtfData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	html := doc.ToHTML()
	if !strings.Contains(html, "&lt;") {
		t.Error("ToHTML should escape < as &lt;")
	}
	if !strings.Contains(html, "&gt;") {
		t.Error("ToHTML should escape > as &gt;")
	}
	if !strings.Contains(html, "&amp;") {
		t.Error("ToHTML should escape & as &amp;")
	}
	if !strings.Contains(html, "&quot;") {
		t.Error("ToHTML should escape \" as &quot;")
	}
}

// TestHTMLBoldToggle verifies bold formatting toggle.
func TestHTMLBoldToggle(t *testing.T) {
	rtfData := `{\rtf1\ansi \b Bold\b0 Normal}`

	doc, err := Parse([]byte(rtfData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	html := doc.ToHTML()
	if !strings.Contains(html, "<b>") {
		t.Error("ToHTML should contain opening bold tag")
	}
	if !strings.Contains(html, "</b>") {
		t.Error("ToHTML should contain closing bold tag")
	}
}

// TestHTMLItalicToggle verifies italic formatting toggle.
func TestHTMLItalicToggle(t *testing.T) {
	rtfData := `{\rtf1\ansi \i Italic\i0 Normal}`

	doc, err := Parse([]byte(rtfData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	html := doc.ToHTML()
	if !strings.Contains(html, "<i>") {
		t.Error("ToHTML should contain opening italic tag")
	}
	if !strings.Contains(html, "</i>") {
		t.Error("ToHTML should contain closing italic tag")
	}
}

// TestHTMLLineBreak verifies line break conversion to <br/>.
func TestHTMLLineBreak(t *testing.T) {
	rtfData := `{\rtf1\ansi Line1\line Line2}`

	doc, err := Parse([]byte(rtfData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	html := doc.ToHTML()
	if !strings.Contains(html, "<br/>") {
		t.Errorf("ToHTML should contain <br/> for line breaks: %q", html)
	}
}

// TestHTMLTab verifies tab conversion to &nbsp;.
func TestHTMLTab(t *testing.T) {
	rtfData := `{\rtf1\ansi Col1\tab Col2}`

	doc, err := Parse([]byte(rtfData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	html := doc.ToHTML()
	if !strings.Contains(html, "&nbsp;") {
		t.Errorf("ToHTML should contain &nbsp; for tabs: %q", html)
	}
}

// TestHTMLUnicode verifies Unicode character rendering in HTML.
func TestHTMLUnicode(t *testing.T) {
	rtfData := `{\rtf1\ansi Hello\u8212 World}` // em-dash

	doc, err := Parse([]byte(rtfData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	html := doc.ToHTML()
	if !strings.Contains(html, "—") {
		t.Errorf("ToHTML should contain em-dash character: %q", html)
	}
}

// TestLaTeXWithTitle verifies LaTeX generation with title.
func TestLaTeXWithTitle(t *testing.T) {
	rtfData := `{\rtf1\ansi{\info{\title My Document}}Content}`

	doc, err := Parse([]byte(rtfData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	latex := doc.ToLaTeX()
	if !strings.Contains(latex, "\\title{My Document}") {
		t.Errorf("ToLaTeX should contain title command: %q", latex)
	}
	if !strings.Contains(latex, "\\maketitle") {
		t.Errorf("ToLaTeX should contain maketitle command: %q", latex)
	}
}

// TestLaTeXWithAuthor verifies LaTeX generation with author.
func TestLaTeXWithAuthor(t *testing.T) {
	rtfData := `{\rtf1\ansi{\info{\author John Doe}}Content}`

	doc, err := Parse([]byte(rtfData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	latex := doc.ToLaTeX()
	if !strings.Contains(latex, "\\author{John Doe}") {
		t.Errorf("ToLaTeX should contain author command: %q", latex)
	}
}

// TestLaTeXWithTitleAndAuthor verifies LaTeX with both title and author.
func TestLaTeXWithTitleAndAuthor(t *testing.T) {
	rtfData := `{\rtf1\ansi{\info{\title Test}{\author Tester}}Content}`

	doc, err := Parse([]byte(rtfData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	latex := doc.ToLaTeX()
	if !strings.Contains(latex, "\\title{Test}") {
		t.Error("ToLaTeX should contain title")
	}
	if !strings.Contains(latex, "\\author{Tester}") {
		t.Error("ToLaTeX should contain author")
	}
	if !strings.Contains(latex, "\\maketitle") {
		t.Error("ToLaTeX should contain maketitle")
	}
}

// TestLaTeXEscaping verifies LaTeX special character escaping.
func TestLaTeXEscaping(t *testing.T) {
	rtfData := `{\rtf1\ansi $ % & # _ \{ \} ~ ^ \\ }`

	doc, err := Parse([]byte(rtfData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	latex := doc.ToLaTeX()
	if !strings.Contains(latex, "\\$") {
		t.Error("ToLaTeX should escape $")
	}
	if !strings.Contains(latex, "\\%") {
		t.Error("ToLaTeX should escape %")
	}
	if !strings.Contains(latex, "\\&") {
		t.Error("ToLaTeX should escape &")
	}
	if !strings.Contains(latex, "\\#") {
		t.Error("ToLaTeX should escape #")
	}
	if !strings.Contains(latex, "\\_") {
		t.Error("ToLaTeX should escape _")
	}
	// Braces are RTF-escaped, producing literal { } which then get LaTeX-escaped
	text := doc.ToText()
	if !strings.Contains(text, "{") || !strings.Contains(text, "}") {
		t.Error("ToText should contain literal braces")
	}
}

// TestLaTeXParagraph verifies paragraph handling in LaTeX.
func TestLaTeXParagraph(t *testing.T) {
	rtfData := `{\rtf1\ansi Para1\par Para2}`

	doc, err := Parse([]byte(rtfData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	latex := doc.ToLaTeX()
	if !strings.Contains(latex, "\n\n") {
		t.Errorf("ToLaTeX should contain paragraph breaks: %q", latex)
	}
}

// TestLaTeXLineBreak verifies line break handling in LaTeX.
func TestLaTeXLineBreak(t *testing.T) {
	rtfData := `{\rtf1\ansi Line1\line Line2}`

	doc, err := Parse([]byte(rtfData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	latex := doc.ToLaTeX()
	if !strings.Contains(latex, "\\\\") {
		t.Errorf("ToLaTeX should contain line breaks: %q", latex)
	}
}

// TestLaTeXTab verifies tab handling in LaTeX.
func TestLaTeXTab(t *testing.T) {
	rtfData := `{\rtf1\ansi Col1\tab Col2}`

	doc, err := Parse([]byte(rtfData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	latex := doc.ToLaTeX()
	if !strings.Contains(latex, "\\quad") {
		t.Errorf("ToLaTeX should contain quad for tabs: %q", latex)
	}
}

// TestLaTeXUnicodePrintable verifies printable Unicode in LaTeX.
func TestLaTeXUnicodePrintable(t *testing.T) {
	rtfData := `{\rtf1\ansi \u65 test}` // 'A'

	doc, err := Parse([]byte(rtfData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	latex := doc.ToLaTeX()
	if !strings.Contains(latex, "A") {
		t.Errorf("ToLaTeX should contain printable Unicode character: %q", latex)
	}
}

// TestLaTeXUnicodeNonPrintable verifies non-printable Unicode in LaTeX.
func TestLaTeXUnicodeNonPrintable(t *testing.T) {
	rtfData := `{\rtf1\ansi \u8212 test}` // em-dash

	doc, err := Parse([]byte(rtfData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	latex := doc.ToLaTeX()
	if !strings.Contains(latex, "\\symbol{") {
		t.Errorf("ToLaTeX should contain symbol command for non-printable Unicode: %q", latex)
	}
}

// TestControlWordWithNegativeParam verifies negative parameter handling.
func TestControlWordWithNegativeParam(t *testing.T) {
	rtfData := `{\rtf1\ansi\li-100 Text}`

	doc, err := Parse([]byte(rtfData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	text := doc.ToText()
	if !strings.Contains(text, "Text") {
		t.Errorf("Should parse control word with negative parameter: %q", text)
	}
}

// TestControlWordWithoutSpace verifies delimiter-less control word.
func TestControlWordWithoutSpace(t *testing.T) {
	rtfData := `{\rtf1\ansi\par Text}`

	doc, err := Parse([]byte(rtfData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	text := doc.ToText()
	if !strings.Contains(text, "Text") {
		t.Errorf("Should parse control word without space delimiter: %q", text)
	}
}

// TestControlSymbol verifies control symbol parsing.
func TestControlSymbol(t *testing.T) {
	rtfData := `{\rtf1\ansi\* Special}`

	doc, err := Parse([]byte(rtfData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	text := doc.ToText()
	if !strings.Contains(text, "Special") {
		t.Errorf("Should parse control symbol: %q", text)
	}
}

// TestParseErrorUnexpectedEndAfterBackslash verifies error on truncated control word.
func TestParseErrorUnexpectedEndAfterBackslash(t *testing.T) {
	rtfData := `{\rtf1\`

	_, err := Parse([]byte(rtfData))
	if err == nil {
		t.Error("Parse should fail for truncated control word")
	}
	if !strings.Contains(err.Error(), "unexpected end") {
		t.Errorf("Error should mention unexpected end: %v", err)
	}
}

// TestInfoGroupNested verifies finding info group nested deeply.
func TestInfoGroupNested(t *testing.T) {
	rtfData := `{\rtf1\ansi{Outer{Inner{\info{\title Nested Title}}}}Content}`

	doc, err := Parse([]byte(rtfData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	meta := doc.Metadata()
	if meta.Title != "Nested Title" {
		t.Errorf("Should find nested info group, got Title = %q", meta.Title)
	}
}

// TestHTMLFormattingGroups verifies HTML generation with formatted groups.
func TestHTMLFormattingGroups(t *testing.T) {
	rtfData := `{\rtf1\ansi Normal {\b Bold group} {\i Italic group}}`

	doc, err := Parse([]byte(rtfData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	html := doc.ToHTML()
	if !strings.Contains(html, "<b>") {
		t.Error("ToHTML should open bold tag for bold group")
	}
	if !strings.Contains(html, "</b>") {
		t.Error("ToHTML should close bold tag for bold group")
	}
	if !strings.Contains(html, "<i>") {
		t.Error("ToHTML should open italic tag for italic group")
	}
	if !strings.Contains(html, "</i>") {
		t.Error("ToHTML should close italic tag for italic group")
	}
}

// TestHTMLSkipSpecialGroups verifies special groups are skipped in HTML.
func TestHTMLSkipSpecialGroups(t *testing.T) {
	rtfData := `{\rtf1\ansi{\fonttbl{\f0 Arial;}}{\colortbl;\red255;}{\stylesheet;}{\pict image}{\object obj}Text}`

	doc, err := Parse([]byte(rtfData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	html := doc.ToHTML()
	if strings.Contains(html, "Arial") {
		t.Error("ToHTML should skip font table")
	}
	if strings.Contains(html, "red255") {
		t.Error("ToHTML should skip color table")
	}
	if strings.Contains(html, "image") {
		t.Error("ToHTML should skip pict group")
	}
	if strings.Contains(html, "obj") {
		t.Error("ToHTML should skip object group")
	}
	if !strings.Contains(html, "Text") {
		t.Error("ToHTML should include regular text")
	}
}

// TestLaTeXSkipSpecialGroups verifies special groups are skipped in LaTeX.
func TestLaTeXSkipSpecialGroups(t *testing.T) {
	rtfData := `{\rtf1\ansi{\fonttbl{\f0 Arial;}}{\colortbl;\red255;}{\stylesheet;}{\pict image}{\object obj}Text}`

	doc, err := Parse([]byte(rtfData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	latex := doc.ToLaTeX()
	if strings.Contains(latex, "Arial") {
		t.Error("ToLaTeX should skip font table")
	}
	if strings.Contains(latex, "red255") {
		t.Error("ToLaTeX should skip color table")
	}
	if strings.Contains(latex, "image") {
		t.Error("ToLaTeX should skip pict group")
	}
	if strings.Contains(latex, "obj") {
		t.Error("ToLaTeX should skip object group")
	}
	if !strings.Contains(latex, "Text") {
		t.Error("ToLaTeX should include regular text")
	}
}

// TestTextSkipSpecialGroups verifies special groups are skipped in plain text.
func TestTextSkipSpecialGroups(t *testing.T) {
	rtfData := `{\rtf1\ansi{\fonttbl{\f0 Arial;}}{\colortbl;\red255;}{\stylesheet;}{\info{\title Skip}}{\pict image}{\object obj}Text}`

	doc, err := Parse([]byte(rtfData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	text := doc.ToText()
	if strings.Contains(text, "Arial") {
		t.Error("ToText should skip font table")
	}
	if strings.Contains(text, "red255") {
		t.Error("ToText should skip color table")
	}
	if strings.Contains(text, "Skip") {
		t.Error("ToText should skip info group")
	}
	if strings.Contains(text, "image") {
		t.Error("ToText should skip pict group")
	}
	if strings.Contains(text, "obj") {
		t.Error("ToText should skip object group")
	}
	if !strings.Contains(text, "Text") {
		t.Error("ToText should include regular text")
	}
}

// TestCarriageReturnNewline verifies CR/LF handling.
func TestCarriageReturnNewline(t *testing.T) {
	rtfData := "{\\rtf1\\ansi Hello\r\nWorld}"

	doc, err := Parse([]byte(rtfData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	text := doc.ToText()
	if !strings.Contains(text, "Hello") {
		t.Error("ToText should contain 'Hello'")
	}
	if !strings.Contains(text, "World") {
		t.Error("ToText should contain 'World'")
	}
}

// TestUnicodeZero verifies handling of \u0.
func TestUnicodeZero(t *testing.T) {
	rtfData := `{\rtf1\ansi Before\u0 After}`

	doc, err := Parse([]byte(rtfData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	text := doc.ToText()
	if !strings.Contains(text, "Before") || !strings.Contains(text, "After") {
		t.Errorf("Should handle \\u0: %q", text)
	}
}

// TestHTMLUnicodeZero verifies HTML handling of \u0.
func TestHTMLUnicodeZero(t *testing.T) {
	rtfData := `{\rtf1\ansi Before\u0 After}`

	doc, err := Parse([]byte(rtfData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	html := doc.ToHTML()
	if !strings.Contains(html, "Before") || !strings.Contains(html, "After") {
		t.Errorf("Should handle \\u0 in HTML: %q", html)
	}
}

// TestLaTeXUnicodeZero verifies LaTeX handling of \u0.
func TestLaTeXUnicodeZero(t *testing.T) {
	rtfData := `{\rtf1\ansi Before\u0 After}`

	doc, err := Parse([]byte(rtfData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	latex := doc.ToLaTeX()
	if !strings.Contains(latex, "Before") || !strings.Contains(latex, "After") {
		t.Errorf("Should handle \\u0 in LaTeX: %q", latex)
	}
}

// TestParseNestedGroupError verifies error handling for invalid nested groups.
func TestParseNestedGroupError(t *testing.T) {
	// Missing closing brace in nested group
	rtfData := `{\rtf1\ansi {Nested without close}`

	_, err := Parse([]byte(rtfData))
	if err == nil {
		t.Error("Parse should fail for unclosed nested group")
	}
	if !strings.Contains(err.Error(), "unclosed group") {
		t.Errorf("Error should mention unclosed group: %v", err)
	}
}

// TestParseControlWordErrorInGroup verifies error when control word fails inside group.
func TestParseControlWordErrorInGroup(t *testing.T) {
	// Control word at end of data (truncated)
	rtfData := []byte(`{\rtf1\ansi Normal text {Group \`)

	_, err := Parse(rtfData)
	if err == nil {
		t.Error("Parse should fail for truncated control word in group")
	}
	if !strings.Contains(err.Error(), "unexpected end") {
		t.Errorf("Error should mention unexpected end: %v", err)
	}
}

// TestHTMLParagraphEdgeCases verifies paragraph handling edge cases.
func TestHTMLParagraphEdgeCases(t *testing.T) {
	// Test paragraph at the end without prior paragraph
	rtfData := `{\rtf1\ansi\par}`

	doc, err := Parse([]byte(rtfData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	html := doc.ToHTML()
	// Should handle \par even when no paragraph was started
	if !strings.Contains(html, "<p>") {
		t.Errorf("ToHTML should create paragraph: %q", html)
	}
}

// TestParseGroupWithNewlines verifies handling of newlines in groups.
func TestParseGroupWithNewlines(t *testing.T) {
	// Newline at group level (not in text)
	rtfData := []byte("{\\rtf1\r}")

	doc, err := Parse(rtfData)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if doc == nil {
		t.Error("Should parse RTF with newline at group level")
	}
}

// TestParseGroupWithLineFeed verifies LF handling in groups.
func TestParseGroupWithLineFeed(t *testing.T) {
	// Line feed at group level
	rtfData := []byte("{\\rtf1\n}")

	doc, err := Parse(rtfData)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if doc == nil {
		t.Error("Should parse RTF with line feed at group level")
	}
}

// TestParseGroupErrorNotAtBrace tests the defensive error check in parseGroup.
func TestParseGroupErrorNotAtBrace(t *testing.T) {
	// Test parseGroup when not positioned at '{'
	parser := &rtfParser{
		data: []byte("not at brace"),
		pos:  0,
	}

	_, err := parser.parseGroup()
	if err == nil {
		t.Error("parseGroup should fail when not at '{'")
	}
	if !strings.Contains(err.Error(), "expected '{'") {
		t.Errorf("Error should mention expected '{': %v", err)
	}
}
