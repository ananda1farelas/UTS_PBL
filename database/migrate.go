package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

const migrationsDir = "migration"

// RunMigrations applies every .sql file inside migration/ that has not been
// applied yet, in filename order (001_, 002_, ...). Applied filenames are
// tracked in schema_migrations so re-running the app is a no-op.
func RunMigrations(db *sql.DB) error {
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			filename VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)
	`); err != nil {
		return fmt.Errorf("migrate: failed to create schema_migrations table: %w", err)
	}

	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.sql"))
	if err != nil {
		return fmt.Errorf("migrate: failed to list migration files: %w", err)
	}
	sort.Strings(files)

	for _, file := range files {
		filename := filepath.Base(file)

		var alreadyApplied bool
		err := db.QueryRow(
			`SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE filename = $1)`,
			filename,
		).Scan(&alreadyApplied)
		if err != nil {
			return fmt.Errorf("migrate: failed to check migration %s: %w", filename, err)
		}
		if alreadyApplied {
			continue
		}

		sqlBytes, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("migrate: failed to read %s: %w", filename, err)
		}

		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("migrate: failed to begin transaction for %s: %w", filename, err)
		}

		if _, err := tx.Exec(string(sqlBytes)); err != nil {
			tx.Rollback()
			return fmt.Errorf("migrate: failed to apply %s: %w", filename, err)
		}

		if _, err := tx.Exec(
			`INSERT INTO schema_migrations (filename) VALUES ($1)`,
			filename,
		); err != nil {
			tx.Rollback()
			return fmt.Errorf("migrate: failed to record %s: %w", filename, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("migrate: failed to commit %s: %w", filename, err)
		}
	}

	return nil
}
