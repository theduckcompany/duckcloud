package storage

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/mattn/go-sqlite3"
	"github.com/qustavo/sqlhooks/v2"
	"golang.org/x/exp/slog"
)

type Config struct {
	DSN   string `mapstructure:"dsn"`
	Debug bool   `mapstructure:"debug"`
}

func NewSQliteClient(cfg Config, log *slog.Logger) (*sql.DB, error) {
	var db *sql.DB
	var err error

	// Sqlite3 doesn't handle well the `sqlite3://` scheme and expect
	// the 'file:' pattern.
	dsn := strings.Replace(cfg.DSN, "sqlite3://", "file:", 1)

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
