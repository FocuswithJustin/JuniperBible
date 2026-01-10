package sqlite

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// DivergenceTest defines a test case for comparing CGO vs pure Go behavior.
// These tests should produce identical results regardless of driver.
type DivergenceTest struct {
	Name     string
	Setup    func(db *sql.DB) error
	Query    func(db *sql.DB) (string, error)
	Expected string // If set, both drivers must return this exact value
}

// divergenceTests contains tests that must produce identical results
// across both CGO and pure Go SQLite implementations.
var divergenceTests = []DivergenceTest{
	{
		Name: "basic_integer",
		Setup: func(db *sql.DB) error {
			_, err := db.Exec(`CREATE TABLE t (v INTEGER)`)
			if err != nil {
				return err
			}
			_, err = db.Exec(`INSERT INTO t VALUES (42)`)
			return err
		},
		Query: func(db *sql.DB) (string, error) {
			var v int
			err := db.QueryRow(`SELECT v FROM t`).Scan(&v)
			return fmt.Sprintf("%d", v), err
		},
		Expected: "42",
	},
	{
		Name: "basic_text",
		Setup: func(db *sql.DB) error {
			_, err := db.Exec(`CREATE TABLE t (v TEXT)`)
			if err != nil {
				return err
			}
			_, err = db.Exec(`INSERT INTO t VALUES ('hello world')`)
			return err
		},
		Query: func(db *sql.DB) (string, error) {
			var v string
			err := db.QueryRow(`SELECT v FROM t`).Scan(&v)
			return v, err
		},
		Expected: "hello world",
	},
	{
		Name: "unicode_text",
		Setup: func(db *sql.DB) error {
			_, err := db.Exec(`CREATE TABLE t (v TEXT)`)
			if err != nil {
				return err
			}
			_, err = db.Exec(`INSERT INTO t VALUES ('בְּרֵאשִׁית בָּרָא אֱלֹהִים')`)
			return err
		},
		Query: func(db *sql.DB) (string, error) {
			var v string
			err := db.QueryRow(`SELECT v FROM t`).Scan(&v)
			return v, err
		},
		Expected: "בְּרֵאשִׁית בָּרָא אֱלֹהִים",
	},
	{
		Name: "null_handling",
		Setup: func(db *sql.DB) error {
			_, err := db.Exec(`CREATE TABLE t (v TEXT)`)
			if err != nil {
				return err
			}
			_, err = db.Exec(`INSERT INTO t VALUES (NULL)`)
			return err
		},
		Query: func(db *sql.DB) (string, error) {
			var v sql.NullString
			err := db.QueryRow(`SELECT v FROM t`).Scan(&v)
			if v.Valid {
				return v.String, err
			}
			return "<NULL>", err
		},
		Expected: "<NULL>",
	},
	{
		Name: "blob_data",
		Setup: func(db *sql.DB) error {
			_, err := db.Exec(`CREATE TABLE t (v BLOB)`)
			if err != nil {
				return err
			}
			_, err = db.Exec(`INSERT INTO t VALUES (X'DEADBEEF')`)
			return err
		},
		Query: func(db *sql.DB) (string, error) {
			var v []byte
			err := db.QueryRow(`SELECT v FROM t`).Scan(&v)
			return hex.EncodeToString(v), err
		},
		Expected: "deadbeef",
	},
	{
		Name: "float_precision",
		Setup: func(db *sql.DB) error {
			_, err := db.Exec(`CREATE TABLE t (v REAL)`)
			if err != nil {
				return err
			}
			_, err = db.Exec(`INSERT INTO t VALUES (3.141592653589793)`)
			return err
		},
		Query: func(db *sql.DB) (string, error) {
			var v float64
			err := db.QueryRow(`SELECT v FROM t`).Scan(&v)
			return fmt.Sprintf("%.15f", v), err
		},
		Expected: "3.141592653589793",
	},
	{
		Name: "aggregate_sum",
		Setup: func(db *sql.DB) error {
			_, err := db.Exec(`CREATE TABLE t (v INTEGER)`)
			if err != nil {
				return err
			}
			for i := 1; i <= 100; i++ {
				_, err = db.Exec(`INSERT INTO t VALUES (?)`, i)
				if err != nil {
					return err
				}
			}
			return nil
		},
		Query: func(db *sql.DB) (string, error) {
			var v int
			err := db.QueryRow(`SELECT SUM(v) FROM t`).Scan(&v)
			return fmt.Sprintf("%d", v), err
		},
		Expected: "5050",
	},
	{
		Name: "string_functions",
		Setup: func(db *sql.DB) error {
			_, err := db.Exec(`CREATE TABLE t (v TEXT)`)
			if err != nil {
				return err
			}
			_, err = db.Exec(`INSERT INTO t VALUES ('Hello World')`)
			return err
		},
		Query: func(db *sql.DB) (string, error) {
			var v string
			err := db.QueryRow(`SELECT UPPER(v) || '|' || LOWER(v) || '|' || LENGTH(v) FROM t`).Scan(&v)
			return v, err
		},
		Expected: "HELLO WORLD|hello world|11",
	},
	{
		Name: "multi_row_order",
		Setup: func(db *sql.DB) error {
			_, err := db.Exec(`CREATE TABLE t (id INTEGER, v TEXT)`)
			if err != nil {
				return err
			}
			values := []string{"charlie", "alpha", "bravo"}
			for i, v := range values {
				_, err = db.Exec(`INSERT INTO t VALUES (?, ?)`, i+1, v)
				if err != nil {
					return err
				}
			}
			return nil
		},
		Query: func(db *sql.DB) (string, error) {
			rows, err := db.Query(`SELECT v FROM t ORDER BY v`)
			if err != nil {
				return "", err
			}
			defer rows.Close()
			var result string
			for rows.Next() {
				var v string
				if err := rows.Scan(&v); err != nil {
					return "", err
				}
				if result != "" {
					result += ","
				}
				result += v
			}
			return result, rows.Err()
		},
		Expected: "alpha,bravo,charlie",
	},
	{
		Name: "transaction_rollback",
		Setup: func(db *sql.DB) error {
			_, err := db.Exec(`CREATE TABLE t (v INTEGER)`)
			if err != nil {
				return err
			}
			_, err = db.Exec(`INSERT INTO t VALUES (1)`)
			if err != nil {
				return err
			}
			// Start transaction, insert, rollback
			tx, err := db.Begin()
			if err != nil {
				return err
			}
			_, err = tx.Exec(`INSERT INTO t VALUES (2)`)
			if err != nil {
				tx.Rollback()
				return err
			}
			tx.Rollback() // Intentional rollback
			return nil
		},
		Query: func(db *sql.DB) (string, error) {
			var count int
			err := db.QueryRow(`SELECT COUNT(*) FROM t`).Scan(&count)
			return fmt.Sprintf("%d", count), err
		},
		Expected: "1", // Only the first insert should be present
	},
}

// TestDivergence runs all divergence tests against the current driver.
// These tests verify that the current driver produces expected results.
func TestDivergence(t *testing.T) {
	for _, tt := range divergenceTests {
		t.Run(tt.Name, func(t *testing.T) {
			tempDir, err := os.MkdirTemp("", "sqlite-divergence-*")
			if err != nil {
				t.Fatalf("failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tempDir)

			dbPath := filepath.Join(tempDir, "test.db")
			db, err := Open(dbPath)
			if err != nil {
				t.Fatalf("failed to open database: %v", err)
			}
			defer db.Close()

			if err := tt.Setup(db); err != nil {
				t.Fatalf("setup failed: %v", err)
			}

			result, err := tt.Query(db)
			if err != nil {
				t.Fatalf("query failed: %v", err)
			}

			if tt.Expected != "" && result != tt.Expected {
				t.Errorf("divergence detected!\n  driver: %s\n  expected: %s\n  got: %s",
					DriverType(), tt.Expected, result)
			}
		})
	}
}

// TestDivergenceHash generates a hash of all test results.
// This hash should be identical for both CGO and pure Go drivers.
func TestDivergenceHash(t *testing.T) {
	hasher := sha256.New()

	for _, tt := range divergenceTests {
		tempDir, err := os.MkdirTemp("", "sqlite-hash-*")
		if err != nil {
			t.Fatalf("failed to create temp dir: %v", err)
		}

		dbPath := filepath.Join(tempDir, "test.db")
		db, err := Open(dbPath)
		if err != nil {
			os.RemoveAll(tempDir)
			t.Fatalf("failed to open database for %s: %v", tt.Name, err)
		}

		if err := tt.Setup(db); err != nil {
			db.Close()
			os.RemoveAll(tempDir)
			t.Fatalf("setup failed for %s: %v", tt.Name, err)
		}

		result, err := tt.Query(db)
		if err != nil {
			db.Close()
			os.RemoveAll(tempDir)
			t.Fatalf("query failed for %s: %v", tt.Name, err)
		}

		// Add to hash
		hasher.Write([]byte(tt.Name))
		hasher.Write([]byte(":"))
		hasher.Write([]byte(result))
		hasher.Write([]byte("\n"))

		db.Close()
		os.RemoveAll(tempDir)
	}

	hash := hex.EncodeToString(hasher.Sum(nil))

	// This hash must be identical for both CGO and pure Go drivers.
	// If this test fails, the drivers have diverged and need investigation.
	expectedHash := "e2fbdfdc9e33fac6b4e2812c044689135c749e4d70f5d2850e1a4ac4205849f5"

	t.Logf("Driver: %s", DriverType())
	t.Logf("Divergence hash: %s", hash)

	if hash != expectedHash {
		t.Errorf("DIVERGENCE DETECTED!\n  Driver: %s\n  Expected: %s\n  Got: %s\n"+
			"This means the %s driver produces different results than expected.\n"+
			"If this is intentional, update the expectedHash constant.",
			DriverType(), expectedHash, hash, DriverType())
	}
}

// GoldenResults stores the expected results from divergence tests.
// Update this when adding new tests or if behavior intentionally changes.
var GoldenResults = map[string]string{
	"basic_integer":        "42",
	"basic_text":           "hello world",
	"unicode_text":         "בְּרֵאשִׁית בָּרָא אֱלֹהִים",
	"null_handling":        "<NULL>",
	"blob_data":            "deadbeef",
	"float_precision":      "3.141592653589793",
	"aggregate_sum":        "5050",
	"string_functions":     "HELLO WORLD|hello world|11",
	"multi_row_order":      "alpha,bravo,charlie",
	"transaction_rollback": "1",
}

// TestGoldenResults verifies all divergence tests match golden results.
func TestGoldenResults(t *testing.T) {
	for _, tt := range divergenceTests {
		golden, ok := GoldenResults[tt.Name]
		if !ok {
			t.Errorf("missing golden result for test: %s", tt.Name)
			continue
		}
		if tt.Expected != "" && tt.Expected != golden {
			t.Errorf("test %s: Expected value mismatch with golden (expected=%s, golden=%s)",
				tt.Name, tt.Expected, golden)
		}
	}
}
