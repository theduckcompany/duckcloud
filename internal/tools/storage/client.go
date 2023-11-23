package storage

import (
	"database/sql"
	"fmt"
)

type Config struct {
	Path string `json:"path"`
}

func NewSQliteClient(cfg *Config) (*sql.DB, error) {
	var db *sql.DB
	var err error

	dsn := "file:" + cfg.Path + "?_journal=WAL&_synchronous=normal&_busy_timeout=500"

	db, err = sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open %q: %w", dsn, err)
	}

	db.SetMaxOpenConns(1)

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}
