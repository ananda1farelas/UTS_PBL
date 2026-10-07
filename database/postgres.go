package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"

	"github.com/ananda1farelas/latihan-fiber/UTS/config"
)

// DB is the shared PostgreSQL connection pool used across the app.
var DB *sql.DB

// Connect opens the PostgreSQL connection pool and verifies it with Ping.
// Call it once at startup, after config.Load().
func Connect(cfg *config.Config) error {
	db, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		return fmt.Errorf("database: failed to open connection: %w", err)
	}

	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	if err := db.Ping(); err != nil {
		return fmt.Errorf("database: failed to ping postgres: %w", err)
	}

	DB = db
	return nil
}

// Close closes the shared connection pool. Call it on graceful shutdown.
func Close() {
	if DB != nil {
		_ = DB.Close()
	}
}
