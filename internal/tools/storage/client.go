package storage

import (
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/mattn/go-sqlite3"
	"github.com/qustavo/sqlhooks/v2"
)

type Config struct {
	Path  string `json:"path"`
	Debug bool   `json:"debug"`
}

func NewSQliteClient(cfg *Config, log *slog.Logger) (*sql.DB, error) {
	var db *sql.DB
	var err error

	dsn := "file:" + cfg.Path + "?_journal=WAL&_synchronous=normal&_busy_timeout=500"

	if cfg.Debug && log != nil {
		sql.Register("sqlite3WithDebugHooks", sqlhooks.Wrap(&sqlite3.SQLiteDriver{}, newDebugHook(log)))
		db, err = sql.Open("sqlite3WithDebugHooks", dsn)
	} else {
		db, err = sql.Open("sqlite3", dsn)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create the sqlite db: %w", err)
	}

	db.SetMaxOpenConns(1)

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}
