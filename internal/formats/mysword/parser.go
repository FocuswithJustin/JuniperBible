// parser.go implements MySword Bible format parsing.
// MySword is an Android Bible app that uses SQLite databases with extensions:
// - .mybible: Bible text (may also contain commentaries/dictionaries)
// - .commentaries.mybible: Commentary
// - .dictionary.mybible: Dictionary
//
// MySword schema is similar to e-Sword but with some differences:
// - Books table: Book, Chapter, Verse, Scripture (same as e-Sword)
// - info table contains module metadata (description, detailed_info, etc.)
package mysword

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/FocuswithJustin/JuniperBible/core/sqlite"
)

// Parser parses MySword Bible files (.mybible).
type Parser struct {
	db       *sql.DB
	filePath string
	metadata map[string]string
}

// NewParser creates a new MySword parser.
func NewParser(filePath string) (*Parser, error) {
	db, err := sqlite.OpenReadOnly(filePath)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	parser := &Parser{
		db:       db,
		filePath: filePath,
		metadata: make(map[string]string),
	}

	// Load metadata from info table
	if err := parser.loadMetadata(); err != nil {
		// Non-fatal, continue with empty metadata
		parser.metadata = make(map[string]string)
	}

	return parser, nil
}

// Close closes the database connection.
func (p *Parser) Close() error {
	if p.db != nil {
		err := p.db.Close()
		p.db = nil
		return err
	}
	return nil
}

// loadMetadata loads metadata from the info table.
func (p *Parser) loadMetadata() error {
	rows, err := p.db.Query("SELECT name, value FROM info")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var name, value string
		if err := rows.Scan(&name, &value); err != nil {
			continue
		}
		p.metadata[name] = value
	}

	return rows.Err()
}

// GetMetadata returns a metadata value by name.
func (p *Parser) GetMetadata(name string) string {
	return p.metadata[name]
}

// GetAllVerses retrieves all verses from the Books table.
func (p *Parser) GetAllVerses() ([]Verse, error) {
	// Try Books table first, fall back to Bible table
	query := "SELECT Book, Chapter, Verse, Scripture FROM Books ORDER BY Book, Chapter, Verse"
	rows, err := p.db.Query(query)
	if err != nil {
		// Try alternate table name
		query = "SELECT Book, Chapter, Verse, Scripture FROM Bible ORDER BY Book, Chapter, Verse"
		rows, err = p.db.Query(query)
		if err != nil {
			return nil, err
		}
	}
	defer rows.Close()

	var verses []Verse
	for rows.Next() {
		var v Verse
		var scripture sql.NullString
		if err := rows.Scan(&v.Book, &v.Chapter, &v.Verse, &scripture); err != nil {
			continue
		}
		v.Text = stripHTML(scripture.String)
		verses = append(verses, v)
	}

	return verses, rows.Err()
}

// Verse represents a single verse from a MySword Bible.
type Verse struct {
	Book    int
	Chapter int
	Verse   int
	Text    string
}

// stripHTML removes basic HTML tags from text.
func stripHTML(text string) string {
	// Simple HTML stripping - remove tags but keep content
	result := text
	for strings.Contains(result, "<") && strings.Contains(result, ">") {
		start := strings.Index(result, "<")
		end := strings.Index(result[start:], ">")
		if end == -1 {
			break
		}
		result = result[:start] + result[start+end+1:]
	}
	return strings.TrimSpace(result)
}

// DetectModuleType determines the type of MySword module.
func DetectModuleType(filename string) string {
	base := strings.ToLower(filename)
	if strings.HasSuffix(base, ".commentaries.mybible") {
		return "commentary"
	} else if strings.HasSuffix(base, ".dictionary.mybible") {
		return "dictionary"
	} else if strings.HasSuffix(base, ".mybible") {
		return "bible"
	}
	return ""
}
