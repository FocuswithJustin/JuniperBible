// parser.go implements MyBible.zone Bible format parsing.
// MyBible is an Android Bible app that uses SQLite databases with extension:
// - .SQLite3: Bible text (MyBible.zone format)
//
// MyBible.zone schema uses lowercase table/column names:
// - verses table: book_number, chapter, verse, text
// - books table: book_number, book_name, book_color
// - info table: name, value pairs for metadata
package mybible

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/FocuswithJustin/JuniperBible/core/sqlite"
)

// Parser parses MyBible Bible files (.SQLite3).
type Parser struct {
	db       *sql.DB
	filePath string
	metadata map[string]string
}

// NewParser creates a new MyBible parser.
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

// GetAllVerses retrieves all verses from the verses table.
func (p *Parser) GetAllVerses() ([]Verse, error) {
	query := "SELECT book_number, chapter, verse, text FROM verses ORDER BY book_number, chapter, verse"
	rows, err := p.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var verses []Verse
	for rows.Next() {
		var v Verse
		var text sql.NullString
		if err := rows.Scan(&v.BookNumber, &v.Chapter, &v.Verse, &text); err != nil {
			continue
		}
		v.Text = stripHTML(text.String)
		verses = append(verses, v)
	}

	return verses, rows.Err()
}

// Verse represents a single verse from a MyBible database.
type Verse struct {
	BookNumber int
	Chapter    int
	Verse      int
	Text       string
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
