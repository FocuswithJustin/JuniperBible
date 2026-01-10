//go:build cgo_sqlite

package main

import (
	_ "github.com/mattn/go-sqlite3" // CGO SQLite driver
)

const sqliteDriver = "sqlite3"
