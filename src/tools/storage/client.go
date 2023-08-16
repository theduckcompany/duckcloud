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

func NewSQliteClient(cfg Config, log *slog.Logger) (*sql.DB, error) {
	var db *sql.DB
	var err error

	dsn := "file:" + cfg.Path
	log.Info(fmt.Sprintf("load database file from %s", cfg.Path))

	switch cfg.Debug {
	case true:
		sql.Register("sqlite3WithDebugHooks", sqlhooks.Wrap(&sqlite3.SQLiteDriver{}, newDebugHook(log)))
		db, err = sql.Open("sqlite3WithDebugHooks", dsn)
	case false:
		db, err = sql.Open("sqlite3", dsn)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create the sqlite db: %w", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}
