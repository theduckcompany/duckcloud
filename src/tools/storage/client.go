package storage

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type Config struct {
	DSN string `mapstructure:"dsn"`
}

func NewSQliteClient(cfg Config) (*sql.DB, error) {
	// Sqlite3 doesn't handle well the `sqlite3://` scheme and expect
	// the 'file:' pattern.
	dsn := strings.Replace(cfg.DSN, "sqlite3://", "file:", 1)

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to create the sqlite db: %w", err)
	}

	return db, nil
}
