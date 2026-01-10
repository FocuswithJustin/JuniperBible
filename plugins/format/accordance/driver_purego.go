//go:build !cgo_sqlite

package main

import (
	_ "modernc.org/sqlite" // Pure Go SQLite driver
)

const sqliteDriver = "sqlite"
