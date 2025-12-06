package db

import (
	"database/sql"
	_ "embed"
	"fmt"
	"strings"
)

//go:embed schema.sql
var schemaSQL string

// Migrate runs database migrations
func (db *DB) Migrate() error {
	// Remove SQL comments first
	lines := strings.Split(schemaSQL, "\n")
	var cleanLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "--") {
			continue
		}
		cleanLines = append(cleanLines, line)
	}
	cleanSQL := strings.Join(cleanLines, "\n")

	// Split schema into individual statements
	statements := strings.Split(cleanSQL, ";")

	// Execute each statement
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}

		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("migration failed: %w\nStatement: %s", err, stmt)
		}
	}

	return nil
}

// GetSchemaVersion returns the current schema version
func (db *DB) GetSchemaVersion() (int, error) {
	// Create schema_version table if it doesn't exist
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_version (
			version INTEGER PRIMARY KEY,
			applied_at TEXT NOT NULL DEFAULT (datetime('now'))
		)
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to create schema_version table: %w", err)
	}

	var version int
	err = db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_version").Scan(&version)
	if err != nil {
		return 0, fmt.Errorf("failed to get schema version: %w", err)
	}

	return version, nil
}

// SetSchemaVersion sets the current schema version
func (db *DB) SetSchemaVersion(version int) error {
	_, err := db.Exec("INSERT INTO schema_version (version) VALUES (?)", version)
	if err != nil {
		return fmt.Errorf("failed to set schema version: %w", err)
	}
	return nil
}

// CheckTables verifies that all expected tables exist
func (db *DB) CheckTables() error {
	expectedTables := []string{
		"subscribers",
		"campaigns",
		"campaign_logs",
		"settings",
	}

	for _, table := range expectedTables {
		var name string
		err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&name)
		if err == sql.ErrNoRows {
			return fmt.Errorf("table %s does not exist", table)
		}
		if err != nil {
			return fmt.Errorf("failed to check table %s: %w", table, err)
		}
	}

	return nil
}
