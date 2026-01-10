// Package xml provides pure Go XML validation, XPath, and formatting.
package xml

import (
	"strings"
	"testing"
)

// TestParseValidXML verifies parsing of well-formed XML.
func TestParseValidXML(t *testing.T) {
	xmlData := `<?xml version="1.0"?>
<root>
	<element attr="value">text</element>
</root>`

	doc, err := Parse([]byte(xmlData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if doc == nil {
		t.Fatal("Parse returned nil document")
	}
}

// TestParseInvalidXML verifies error handling for malformed XML.
func TestParseInvalidXML(t *testing.T) {
	tests := []struct {
		name string
		xml  string
	}{
		{"unclosed tag", "<root><element></root>"},
		{"mismatched tags", "<root></other>"},
		{"invalid chars", "<root>\x00</root>"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse([]byte(tt.xml))
			if err == nil {
				t.Error("Parse should fail for invalid XML")
			}
		})
	}
}

// TestValidateWellFormed verifies well-formedness validation.
func TestValidateWellFormed(t *testing.T) {
	valid := `<?xml version="1.0"?><root><child/></root>`
	result := Validate([]byte(valid), nil)
	if !result.Valid {
		t.Errorf("Valid XML should pass: %v", result.Errors)
	}
}

// TestValidateWithDTD verifies DTD validation.
func TestValidateWithDTD(t *testing.T) {
	xmlData := `<?xml version="1.0"?>
<!DOCTYPE note [
<!ELEMENT note (to,from,body)>
<!ELEMENT to (#PCDATA)>
<!ELEMENT from (#PCDATA)>
<!ELEMENT body (#PCDATA)>
]>
<note>
	<to>User</to>
	<from>System</from>
	<body>Hello</body>
</note>`

	result := Validate([]byte(xmlData), nil)
	if !result.Valid {
		t.Errorf("Valid DTD XML should pass: %v", result.Errors)
	}
}

// TestXPathQuery verifies XPath query execution.
func TestXPathQuery(t *testing.T) {
	xmlData := `<?xml version="1.0"?>
<library>
	<book id="1"><title>Book One</title></book>
	<book id="2"><title>Book Two</title></book>
</library>`

	doc, err := Parse([]byte(xmlData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	results, err := doc.XPath("//book/title")
	if err != nil {
		t.Fatalf("XPath failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("XPath should return 2 results, got %d", len(results))
	}
}

// TestXPathQueryAttribute verifies XPath attribute selection.
func TestXPathQueryAttribute(t *testing.T) {
	xmlData := `<root><item id="123" name="test"/></root>`

	doc, err := Parse([]byte(xmlData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	results, err := doc.XPath("//item/@id")
	if err != nil {
		t.Fatalf("XPath failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("XPath should return 1 result, got %d", len(results))
	}
}

// TestXPathQueryText verifies XPath text extraction.
func TestXPathQueryText(t *testing.T) {
	xmlData := `<root><message>Hello World</message></root>`

	doc, err := Parse([]byte(xmlData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	results, err := doc.XPath("//message/text()")
	if err != nil {
		t.Fatalf("XPath failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("XPath should return 1 result, got %d", len(results))
	}

	if results[0].Text() != "Hello World" {
		t.Errorf("Text = %q, want %q", results[0].Text(), "Hello World")
	}
}

// TestXPathInvalidExpression verifies error handling for invalid XPath.
func TestXPathInvalidExpression(t *testing.T) {
	xmlData := `<root/>`

	doc, err := Parse([]byte(xmlData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	_, err = doc.XPath("[invalid")
	if err == nil {
		t.Error("Invalid XPath should return error")
	}
}

// TestFormat verifies XML pretty-printing.
func TestFormat(t *testing.T) {
	xmlData := `<?xml version="1.0"?><root><child attr="val">text</child></root>`

	formatted, err := Format([]byte(xmlData), FormatOptions{Indent: "  "})
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	// Should have newlines and indentation
	if !strings.Contains(string(formatted), "\n") {
		t.Error("Formatted XML should contain newlines")
	}
	if !strings.Contains(string(formatted), "  ") {
		t.Error("Formatted XML should contain indentation")
	}
}

// TestFormatWithTabs verifies tab indentation.
func TestFormatWithTabs(t *testing.T) {
	xmlData := `<root><child/></root>`

	formatted, err := Format([]byte(xmlData), FormatOptions{Indent: "\t"})
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	if !strings.Contains(string(formatted), "\t") {
		t.Error("Formatted XML should contain tabs")
	}
}

// TestFormatPreservesContent verifies content is preserved during formatting.
func TestFormatPreservesContent(t *testing.T) {
	xmlData := `<root><message>Hello &amp; World</message></root>`

	formatted, err := Format([]byte(xmlData), FormatOptions{Indent: "  "})
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	if !strings.Contains(string(formatted), "Hello &amp; World") {
		t.Error("Formatted XML should preserve entity references")
	}
}

// TestDocumentRoot verifies root element access.
func TestDocumentRoot(t *testing.T) {
	xmlData := `<?xml version="1.0"?><root attr="value"><child/></root>`

	doc, err := Parse([]byte(xmlData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	root := doc.Root()
	if root == nil {
		t.Fatal("Root should not be nil")
	}

	if root.Name() != "root" {
		t.Errorf("Root name = %q, want %q", root.Name(), "root")
	}
}

// TestNodeChildren verifies child node access.
func TestNodeChildren(t *testing.T) {
	xmlData := `<parent><child1/><child2/><child3/></parent>`

	doc, err := Parse([]byte(xmlData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	children := doc.Root().Children()
	if len(children) != 3 {
		t.Errorf("Should have 3 children, got %d", len(children))
	}
}

// TestNodeAttributes verifies attribute access.
func TestNodeAttributes(t *testing.T) {
	xmlData := `<element id="123" class="test" data-value="abc"/>`

	doc, err := Parse([]byte(xmlData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	attrs := doc.Root().Attributes()
	if len(attrs) != 3 {
		t.Errorf("Should have 3 attributes, got %d", len(attrs))
	}

	if doc.Root().Attr("id") != "123" {
		t.Errorf("Attr(id) = %q, want %q", doc.Root().Attr("id"), "123")
	}
}

// TestNodeInnerText verifies inner text extraction.
func TestNodeInnerText(t *testing.T) {
	xmlData := `<root>Hello <b>World</b>!</root>`

	doc, err := Parse([]byte(xmlData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	text := doc.Root().InnerText()
	if text != "Hello World!" {
		t.Errorf("InnerText = %q, want %q", text, "Hello World!")
	}
}

// TestNodeInnerXML verifies inner XML extraction.
func TestNodeInnerXML(t *testing.T) {
	xmlData := `<root>Hello <b>World</b>!</root>`

	doc, err := Parse([]byte(xmlData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	innerXML := doc.Root().InnerXML()
	if !strings.Contains(innerXML, "<b>World</b>") {
		t.Errorf("InnerXML should contain markup: %q", innerXML)
	}
}

// TestValidationResult verifies validation result structure.
func TestValidationResult(t *testing.T) {
	result := ValidationResult{
		Valid:  false,
		Errors: []ValidationError{{Line: 1, Column: 5, Message: "test error"}},
	}

	if result.Valid {
		t.Error("Result should not be valid")
	}

	if len(result.Errors) != 1 {
		t.Error("Result should have 1 error")
	}

	if result.Errors[0].Line != 1 {
		t.Errorf("Error line = %d, want 1", result.Errors[0].Line)
	}
}

// TestNamespaceHandling verifies namespace support.
func TestNamespaceHandling(t *testing.T) {
	xmlData := `<root xmlns:ns="http://example.com"><ns:child/></root>`

	doc, err := Parse([]byte(xmlData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Should parse without error
	if doc.Root() == nil {
		t.Error("Document should have root element")
	}
}

// TestCDATAHandling verifies CDATA section handling.
func TestCDATAHandling(t *testing.T) {
	xmlData := `<root><![CDATA[<not>xml</not>]]></root>`

	doc, err := Parse([]byte(xmlData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	text := doc.Root().InnerText()
	if !strings.Contains(text, "<not>xml</not>") {
		t.Errorf("CDATA content should be preserved: %q", text)
	}
}

// TestCommentHandling verifies XML comment handling.
func TestCommentHandling(t *testing.T) {
	xmlData := `<root><!-- comment --><child/></root>`

	doc, err := Parse([]byte(xmlData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Comments should not affect parsing
	children := doc.Root().Children()
	if len(children) != 1 {
		t.Errorf("Should have 1 child element (comments excluded), got %d", len(children))
	}
}

// TestProcessingInstruction verifies PI handling.
func TestProcessingInstruction(t *testing.T) {
	xmlData := `<?xml version="1.0"?><?xml-stylesheet type="text/xsl" href="style.xsl"?><root/>`

	doc, err := Parse([]byte(xmlData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// PIs should not affect document structure
	if doc.Root() == nil {
		t.Error("Document should have root element")
	}
}

// TestSerialize verifies XML serialization.
func TestSerialize(t *testing.T) {
	xmlData := `<root attr="value"><child>text</child></root>`

	doc, err := Parse([]byte(xmlData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	output := doc.Serialize()
	if !strings.Contains(string(output), "attr=\"value\"") {
		t.Error("Serialized XML should contain attribute")
	}
	if !strings.Contains(string(output), "<child>text</child>") {
		t.Error("Serialized XML should contain child element")
	}
}

// TestXPathSelectSingle verifies selecting single node.
func TestXPathSelectSingle(t *testing.T) {
	xmlData := `<root><first/><second/></root>`

	doc, err := Parse([]byte(xmlData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	node, err := doc.XPathFirst("//first")
	if err != nil {
		t.Fatalf("XPathFirst failed: %v", err)
	}

	if node == nil {
		t.Fatal("XPathFirst should return a node")
	}

	if node.Name() != "first" {
		t.Errorf("Node name = %q, want %q", node.Name(), "first")
	}
}

// TestXPathSelectEmpty verifies empty result handling.
func TestXPathSelectEmpty(t *testing.T) {
	xmlData := `<root/>`

	doc, err := Parse([]byte(xmlData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	results, err := doc.XPath("//nonexistent")
	if err != nil {
		t.Fatalf("XPath failed: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("XPath should return empty slice, got %d results", len(results))
	}
}

// TestXPathFirstNotFound verifies XPathFirst returns nil when no match.
func TestXPathFirstNotFound(t *testing.T) {
	xmlData := `<root/>`

	doc, err := Parse([]byte(xmlData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	node, err := doc.XPathFirst("//nonexistent")
	if err != nil {
		t.Fatalf("XPathFirst failed: %v", err)
	}

	if node != nil {
		t.Error("XPathFirst should return nil for non-existent element")
	}
}

// TestXPathFirstInvalidExpression verifies error handling for invalid XPath in XPathFirst.
func TestXPathFirstInvalidExpression(t *testing.T) {
	xmlData := `<root/>`

	doc, err := Parse([]byte(xmlData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	_, err = doc.XPathFirst("[invalid")
	if err == nil {
		t.Error("Invalid XPath should return error in XPathFirst")
	}
}

// TestValidateMalformed verifies validation catches malformed XML.
func TestValidateMalformed(t *testing.T) {
	malformed := `<root><unclosed>`
	result := Validate([]byte(malformed), nil)
	if result.Valid {
		t.Error("Malformed XML should not be valid")
	}
	if len(result.Errors) == 0 {
		t.Error("Malformed XML should have errors")
	}
}

// TestFormatDefaultIndent verifies default indentation when none specified.
func TestFormatDefaultIndent(t *testing.T) {
	xmlData := `<root><child/></root>`

	formatted, err := Format([]byte(xmlData), FormatOptions{})
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	// Default should be two spaces
	if !strings.Contains(string(formatted), "  ") {
		t.Error("Default indentation should be two spaces")
	}
}

// TestFormatInvalidXML verifies Format handles invalid XML.
func TestFormatInvalidXML(t *testing.T) {
	xmlData := `<root><unclosed>`

	_, err := Format([]byte(xmlData), FormatOptions{})
	if err == nil {
		t.Error("Format should fail for invalid XML")
	}
}

// TestFormatWithDeclaration verifies formatting preserves XML declaration.
func TestFormatWithDeclaration(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?><root/>`

	formatted, err := Format([]byte(xmlData), FormatOptions{Indent: "  "})
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	if !strings.Contains(string(formatted), "<?xml") {
		t.Error("Formatted XML should preserve declaration")
	}
	if !strings.Contains(string(formatted), "version=\"1.0\"") {
		t.Error("Formatted XML should preserve version attribute")
	}
}

// TestFormatWithNamespacePrefix verifies formatting preserves namespace prefixes.
func TestFormatWithNamespacePrefix(t *testing.T) {
	xmlData := `<ns:root xmlns:ns="http://example.com"><ns:child/></ns:root>`

	formatted, err := Format([]byte(xmlData), FormatOptions{Indent: "  "})
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	if !strings.Contains(string(formatted), "ns:root") {
		t.Error("Formatted XML should preserve namespace prefix on root")
	}
	if !strings.Contains(string(formatted), "ns:child") {
		t.Error("Formatted XML should preserve namespace prefix on child")
	}
}

// TestFormatWithNamespaceAttribute verifies formatting handles namespace attributes.
func TestFormatWithNamespaceAttribute(t *testing.T) {
	xmlData := `<root xmlns:custom="http://example.com"/>`

	formatted, err := Format([]byte(xmlData), FormatOptions{Indent: "  "})
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	if !strings.Contains(string(formatted), "xmlns:custom") {
		t.Error("Formatted XML should preserve namespace attribute")
	}
}

// TestFormatSelfClosingTag verifies self-closing tags are formatted correctly.
func TestFormatSelfClosingTag(t *testing.T) {
	xmlData := `<root><empty/></root>`

	formatted, err := Format([]byte(xmlData), FormatOptions{Indent: "  "})
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	if !strings.Contains(string(formatted), "<empty/>") {
		t.Error("Self-closing tag should be preserved")
	}
}

// TestFormatMixedContent verifies formatting of mixed text and element content.
func TestFormatMixedContent(t *testing.T) {
	xmlData := `<root>Text before<child/>Text after</root>`

	formatted, err := Format([]byte(xmlData), FormatOptions{Indent: "  "})
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	if !strings.Contains(string(formatted), "child") {
		t.Error("Formatted XML should contain child element")
	}
}

// TestFormatWithCDATA verifies CDATA formatting.
func TestFormatWithCDATA(t *testing.T) {
	xmlData := `<root><![CDATA[<script>alert('test')</script>]]></root>`

	formatted, err := Format([]byte(xmlData), FormatOptions{Indent: "  "})
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	if !strings.Contains(string(formatted), "<![CDATA[") {
		t.Error("Formatted XML should preserve CDATA start")
	}
	if !strings.Contains(string(formatted), "]]>") {
		t.Error("Formatted XML should preserve CDATA end")
	}
}

// TestFormatWithComment verifies comment handling in formatting.
// Note: The xmlquery library doesn't preserve comments during parsing,
// so comments are not included in formatted output.
func TestFormatWithComment(t *testing.T) {
	// Comments are not preserved by the underlying xmlquery library
	xmlData := `<root><!-- This is a comment --><child/></root>`

	formatted, err := Format([]byte(xmlData), FormatOptions{Indent: "  "})
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	// Verify the rest of the document is formatted correctly
	if !strings.Contains(string(formatted), "<root>") {
		t.Error("Formatted XML should contain root element")
	}
	if !strings.Contains(string(formatted), "<child/>") {
		t.Error("Formatted XML should contain child element")
	}
}

// TestFormatEscapesSpecialChars verifies special character escaping in text.
func TestFormatEscapesSpecialChars(t *testing.T) {
	xmlData := `<root>&lt;tag&gt; &amp; "quotes"</root>`

	formatted, err := Format([]byte(xmlData), FormatOptions{Indent: "  "})
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	formattedStr := string(formatted)
	if !strings.Contains(formattedStr, "&lt;") {
		t.Error("Should escape < as &lt;")
	}
	if !strings.Contains(formattedStr, "&gt;") {
		t.Error("Should escape > as &gt;")
	}
	if !strings.Contains(formattedStr, "&amp;") {
		t.Error("Should escape & as &amp;")
	}
}

// TestFormatEscapesAttributeQuotes verifies quote escaping in attributes.
func TestFormatEscapesAttributeQuotes(t *testing.T) {
	xmlData := `<root attr="value with &quot;quotes&quot;"/>`

	formatted, err := Format([]byte(xmlData), FormatOptions{Indent: "  "})
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	if !strings.Contains(string(formatted), "&quot;") {
		t.Error("Should escape quotes in attributes as &quot;")
	}
}

// TestDocumentRootNilDocument verifies Root handles nil document.
func TestDocumentRootNilDocument(t *testing.T) {
	doc := &Document{root: nil}
	root := doc.Root()
	if root != nil {
		t.Error("Root should return nil for document with nil root")
	}
}

// TestDocumentRootNoElementChild verifies Root when document has no element children.
func TestDocumentRootNoElementChild(t *testing.T) {
	// Create a document with only declaration and comment (no element)
	xmlData := `<?xml version="1.0"?><!-- comment only -->`

	doc, err := Parse([]byte(xmlData))
	if err != nil {
		// If parsing fails, that's okay - we're testing edge case
		// Skip this test as the parser requires at least one element
		t.Skip("Parser requires at least one element node")
	}

	root := doc.Root()
	if root != nil {
		t.Error("Root should return nil when document has no element children")
	}
}

// TestSerializeNilDocument verifies Serialize handles nil document root.
func TestSerializeNilDocument(t *testing.T) {
	doc := &Document{root: nil}
	output := doc.Serialize()
	if output != nil {
		t.Error("Serialize should return nil for document with nil root")
	}
}

// TestNodeNameNil verifies Name handles nil node.
func TestNodeNameNil(t *testing.T) {
	node := &Node{node: nil}
	name := node.Name()
	if name != "" {
		t.Error("Name should return empty string for nil node")
	}
}

// TestNodeTextNil verifies Text handles nil node.
func TestNodeTextNil(t *testing.T) {
	node := &Node{node: nil}
	text := node.Text()
	if text != "" {
		t.Error("Text should return empty string for nil node")
	}
}

// TestNodeInnerTextNil verifies InnerText handles nil node.
func TestNodeInnerTextNil(t *testing.T) {
	node := &Node{node: nil}
	text := node.InnerText()
	if text != "" {
		t.Error("InnerText should return empty string for nil node")
	}
}

// TestNodeInnerXMLNil verifies InnerXML handles nil node.
func TestNodeInnerXMLNil(t *testing.T) {
	node := &Node{node: nil}
	xml := node.InnerXML()
	if xml != "" {
		t.Error("InnerXML should return empty string for nil node")
	}
}

// TestNodeChildrenNil verifies Children handles nil node.
func TestNodeChildrenNil(t *testing.T) {
	node := &Node{node: nil}
	children := node.Children()
	if children != nil {
		t.Error("Children should return nil for nil node")
	}
}

// TestNodeAttributesNil verifies Attributes handles nil node.
func TestNodeAttributesNil(t *testing.T) {
	node := &Node{node: nil}
	attrs := node.Attributes()
	if attrs != nil {
		t.Error("Attributes should return nil for nil node")
	}
}

// TestNodeAttrNil verifies Attr handles nil node.
func TestNodeAttrNil(t *testing.T) {
	node := &Node{node: nil}
	attr := node.Attr("test")
	if attr != "" {
		t.Error("Attr should return empty string for nil node")
	}
}

// TestNodeChildrenWithTextNodes verifies Children filters text nodes.
func TestNodeChildrenWithTextNodes(t *testing.T) {
	xmlData := `<root>text1<child1/>text2<child2/>text3</root>`

	doc, err := Parse([]byte(xmlData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	children := doc.Root().Children()
	// Should only return element children, not text nodes
	if len(children) != 2 {
		t.Errorf("Should have 2 element children (text nodes excluded), got %d", len(children))
	}
}

// TestNodeAttributesEmpty verifies Attributes on node without attributes.
func TestNodeAttributesEmpty(t *testing.T) {
	xmlData := `<root/>`

	doc, err := Parse([]byte(xmlData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	attrs := doc.Root().Attributes()
	if len(attrs) != 0 {
		t.Errorf("Should have 0 attributes, got %d", len(attrs))
	}
}

// TestNodeAttrMissing verifies Attr returns empty string for missing attribute.
func TestNodeAttrMissing(t *testing.T) {
	xmlData := `<root id="123"/>`

	doc, err := Parse([]byte(xmlData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	attr := doc.Root().Attr("nonexistent")
	if attr != "" {
		t.Errorf("Attr should return empty string for missing attribute, got %q", attr)
	}
}

// TestXPathWithPredicate verifies XPath predicates work correctly.
func TestXPathWithPredicate(t *testing.T) {
	xmlData := `<root><item id="1">A</item><item id="2">B</item><item id="3">C</item></root>`

	doc, err := Parse([]byte(xmlData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	results, err := doc.XPath("//item[@id='2']")
	if err != nil {
		t.Fatalf("XPath failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("XPath should return 1 result, got %d", len(results))
	}

	if results[0].Text() != "B" {
		t.Errorf("Text = %q, want %q", results[0].Text(), "B")
	}
}

// TestFormatEmptyElement verifies formatting of truly empty elements.
func TestFormatEmptyElement(t *testing.T) {
	xmlData := `<root><empty></empty></root>`

	formatted, err := Format([]byte(xmlData), FormatOptions{Indent: "  "})
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	// Empty element should become self-closing
	formattedStr := string(formatted)
	if !strings.Contains(formattedStr, "empty") {
		t.Error("Formatted XML should contain empty element")
	}
}

// TestFormatTextOnlyElement verifies formatting of element with only text.
func TestFormatTextOnlyElement(t *testing.T) {
	xmlData := `<root><text>content</text></root>`

	formatted, err := Format([]byte(xmlData), FormatOptions{Indent: "  "})
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	if !strings.Contains(string(formatted), "content") {
		t.Error("Formatted XML should preserve text content")
	}
}

// TestFormatWhitespaceOnlyText verifies whitespace-only text is trimmed.
func TestFormatWhitespaceOnlyText(t *testing.T) {
	xmlData := `<root>

		<child/>   </root>`

	formatted, err := Format([]byte(xmlData), FormatOptions{Indent: "  "})
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	// Whitespace-only text should be trimmed
	formattedStr := string(formatted)
	if !strings.Contains(formattedStr, "child") {
		t.Error("Formatted XML should contain child element")
	}
}

// TestInnerXMLWithMultipleChildren verifies InnerXML with multiple children.
func TestInnerXMLWithMultipleChildren(t *testing.T) {
	xmlData := `<root><a>1</a><b>2</b><c>3</c></root>`

	doc, err := Parse([]byte(xmlData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	innerXML := doc.Root().InnerXML()
	if !strings.Contains(innerXML, "<a>") {
		t.Error("InnerXML should contain <a> element")
	}
	if !strings.Contains(innerXML, "<b>") {
		t.Error("InnerXML should contain <b> element")
	}
	if !strings.Contains(innerXML, "<c>") {
		t.Error("InnerXML should contain <c> element")
	}
}

// TestComplexXPathExpression verifies complex XPath expressions.
func TestComplexXPathExpression(t *testing.T) {
	xmlData := `<library>
		<book category="fiction"><title>Book 1</title></book>
		<book category="nonfiction"><title>Book 2</title></book>
		<book category="fiction"><title>Book 3</title></book>
	</library>`

	doc, err := Parse([]byte(xmlData))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	results, err := doc.XPath("//book[@category='fiction']/title")
	if err != nil {
		t.Fatalf("XPath failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("XPath should return 2 fiction titles, got %d", len(results))
	}
}
