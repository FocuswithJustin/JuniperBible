// Package markup provides converters for Bible markup formats to Markdown.
package markup

import (
	"regexp"
	"strings"
)

// GBFConverter converts GBF (General Bible Format) to Markdown.
//
// GBF is an older SWORD markup format that uses angle-bracket codes:
//   - <FR>...<Fr> - Red letter (words of Jesus)
//   - <FI>...<Fi> - Italic
//   - <FB>...<Fb> - Bold
//   - <FO>...<Fo> - Old Testament quote
//   - <FS>...<Fs> - Superscript
//   - <FU>...<Fu> - Underline
//   - <RF>...<Rf> - Footnote
//   - <RX>...<Rx> - Cross-reference
//   - <WH>...<Wh> - Hebrew word with Strong's
//   - <WG>...<Wg> - Greek word with Strong's
//   - <WT>...<Wt> - Morphology tag
//   - <CM> - Paragraph/section mark
//   - <CL> - Line break
//   - <CI> - Indent
type GBFConverter struct {
	// PreserveStrongs keeps Strong's numbers as annotations
	PreserveStrongs bool

	// PreserveMorphology keeps morphology codes
	PreserveMorphology bool
}

// NewGBFConverter creates a converter with default settings.
func NewGBFConverter() *GBFConverter {
	return &GBFConverter{
		PreserveStrongs:    true,
		PreserveMorphology: false,
	}
}

// GBFResult contains the converted text and extracted metadata.
type GBFResult struct {
	Text       string
	Strongs    []string
	Morphology []string
	Footnotes  []string
	CrossRefs  []string
	HasRedText bool
}

// Convert transforms GBF markup to Markdown.
func (c *GBFConverter) Convert(gbf string) *GBFResult {
	result := &GBFResult{
		Strongs:    make([]string, 0),
		Morphology: make([]string, 0),
		Footnotes:  make([]string, 0),
		CrossRefs:  make([]string, 0),
	}

	text := gbf

	// Extract Strong's numbers
	if c.PreserveStrongs {
		result.Strongs = c.extractStrongs(text)
	}

	// Extract morphology
	if c.PreserveMorphology {
		result.Morphology = c.extractMorphology(text)
	}

	// Extract footnotes
	result.Footnotes = c.extractFootnotes(text)

	// Extract cross-references
	result.CrossRefs = c.extractCrossRefs(text)

	// Check for red letter text
	result.HasRedText = strings.Contains(text, "<FR>")

	// Convert to Markdown
	text = c.convertToMarkdown(text)

	result.Text = strings.TrimSpace(text)
	return result
}

// extractStrongs extracts Strong's numbers from GBF word tags.
func (c *GBFConverter) extractStrongs(text string) []string {
	strongs := make([]string, 0)
	seen := make(map[string]bool)

	// Hebrew: <WH1234> or <WHxxxx>
	reH := regexp.MustCompile(`<WH(\d+)>`)
	matchesH := reH.FindAllStringSubmatch(text, -1)
	for _, match := range matchesH {
		if len(match) > 1 {
			num := "H" + match[1]
			if !seen[num] {
				strongs = append(strongs, num)
				seen[num] = true
			}
		}
	}

	// Greek: <WG1234> or <WGxxxx>
	reG := regexp.MustCompile(`<WG(\d+)>`)
	matchesG := reG.FindAllStringSubmatch(text, -1)
	for _, match := range matchesG {
		if len(match) > 1 {
			num := "G" + match[1]
			if !seen[num] {
				strongs = append(strongs, num)
				seen[num] = true
			}
		}
	}

	return strongs
}

// extractMorphology extracts morphology codes from GBF.
func (c *GBFConverter) extractMorphology(text string) []string {
	re := regexp.MustCompile(`<WT([^>]+)>`)
	matches := re.FindAllStringSubmatch(text, -1)

	morphs := make([]string, 0, len(matches))
	seen := make(map[string]bool)

	for _, match := range matches {
		if len(match) > 1 && !seen[match[1]] {
			morphs = append(morphs, match[1])
			seen[match[1]] = true
		}
	}

	return morphs
}

// extractFootnotes extracts footnote content from GBF.
func (c *GBFConverter) extractFootnotes(text string) []string {
	re := regexp.MustCompile(`<RF>([^<]*)<Rf>`)
	matches := re.FindAllStringSubmatch(text, -1)

	notes := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) > 1 && strings.TrimSpace(match[1]) != "" {
			notes = append(notes, strings.TrimSpace(match[1]))
		}
	}

	return notes
}

// extractCrossRefs extracts cross-reference content from GBF.
func (c *GBFConverter) extractCrossRefs(text string) []string {
	re := regexp.MustCompile(`<RX>([^<]*)<Rx>`)
	matches := re.FindAllStringSubmatch(text, -1)

	refs := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) > 1 && strings.TrimSpace(match[1]) != "" {
			refs = append(refs, strings.TrimSpace(match[1]))
		}
	}

	return refs
}

// convertToMarkdown converts GBF markup to Markdown.
func (c *GBFConverter) convertToMarkdown(text string) string {
	// Use placeholders for content we want to preserve
	const redLetterStart = "\x00RED_START\x00"
	const redLetterEnd = "\x00RED_END\x00"

	// Convert red letter text
	re := regexp.MustCompile(`<FR>([^<]*)<Fr>`)
	text = re.ReplaceAllString(text, redLetterStart+"$1"+redLetterEnd)

	// Convert italic
	re = regexp.MustCompile(`<FI>([^<]*)<Fi>`)
	text = re.ReplaceAllString(text, "*$1*")

	// Convert bold
	re = regexp.MustCompile(`<FB>([^<]*)<Fb>`)
	text = re.ReplaceAllString(text, "**$1**")

	// Convert OT quotations (use blockquote style)
	re = regexp.MustCompile(`<FO>([^<]*)<Fo>`)
	text = re.ReplaceAllString(text, "> $1")

	// Convert superscript (use caret notation like Strong's)
	re = regexp.MustCompile(`<FS>([^<]*)<Fs>`)
	text = re.ReplaceAllString(text, "^$1^")

	// Convert underline to emphasis (Markdown has no underline)
	re = regexp.MustCompile(`<FU>([^<]*)<Fu>`)
	text = re.ReplaceAllString(text, "_$1_")

	// Handle Strong's numbers
	if c.PreserveStrongs {
		// Hebrew Strong's: <WH1234>word<Wh>
		re = regexp.MustCompile(`<WH(\d+)>([^<]*)<Wh>`)
		text = re.ReplaceAllString(text, "$2^H$1^")

		// Greek Strong's: <WG1234>word<Wg>
		re = regexp.MustCompile(`<WG(\d+)>([^<]*)<Wg>`)
		text = re.ReplaceAllString(text, "$2^G$1^")
	} else {
		// Just extract words, remove Strong's tags
		re = regexp.MustCompile(`<WH\d+>([^<]*)<Wh>`)
		text = re.ReplaceAllString(text, "$1")
		re = regexp.MustCompile(`<WG\d+>([^<]*)<Wg>`)
		text = re.ReplaceAllString(text, "$1")
	}

	// Remove morphology tags
	re = regexp.MustCompile(`<WT[^>]*>`)
	text = re.ReplaceAllString(text, "")
	re = regexp.MustCompile(`<Wt>`)
	text = re.ReplaceAllString(text, "")

	// Remove footnotes (already extracted)
	re = regexp.MustCompile(`<RF>[^<]*<Rf>`)
	text = re.ReplaceAllString(text, "")

	// Remove cross-references (already extracted)
	re = regexp.MustCompile(`<RX>[^<]*<Rx>`)
	text = re.ReplaceAllString(text, "")

	// Convert paragraph/section marks
	re = regexp.MustCompile(`<CM>`)
	text = re.ReplaceAllString(text, "\n\n")

	// Convert line breaks
	re = regexp.MustCompile(`<CL>`)
	text = re.ReplaceAllString(text, "\n")

	// Convert indents
	re = regexp.MustCompile(`<CI>`)
	text = re.ReplaceAllString(text, "    ")

	// Remove title markers but keep content
	re = regexp.MustCompile(`<TS>([^<]*)<Ts>`)
	text = re.ReplaceAllString(text, "\n### $1\n")

	// Remove any remaining GBF tags
	re = regexp.MustCompile(`<[A-Z][A-Za-z0-9]*>`)
	text = re.ReplaceAllString(text, "")
	re = regexp.MustCompile(`<[A-Za-z][a-z]>`)
	text = re.ReplaceAllString(text, "")

	// Restore placeholders
	text = strings.ReplaceAll(text, redLetterStart, `<span class="red-letter">`)
	text = strings.ReplaceAll(text, redLetterEnd, `</span>`)

	// Clean up whitespace
	re = regexp.MustCompile(`[ \t]+`)
	text = re.ReplaceAllString(text, " ")

	// Remove leading/trailing whitespace from lines
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimSpace(line)
	}
	text = strings.Join(lines, "\n")

	// Remove multiple blank lines
	re = regexp.MustCompile(`\n{3,}`)
	text = re.ReplaceAllString(text, "\n\n")

	return text
}

// ConvertText converts GBF to plain Markdown text.
func (c *GBFConverter) ConvertText(gbf string) string {
	result := c.Convert(gbf)
	return result.Text
}
