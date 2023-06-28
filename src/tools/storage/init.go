package storage

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"path"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/exp/slog"
)

type Config struct {
	Path string `mapstructure:"url"`
}

//go:embed db/migration/*.sql
var fs embed.FS

func NewSQliteDBWithMigrate(cfg Config, logger *slog.Logger) (*sql.DB, error) {
	d, err := iofs.New(fs, "db/migration")
	if err != nil {
		return nil, fmt.Errorf("failed to load the migrated files: %w", err)
	}

	dbPath := path.Clean(cfg.Path)

	m, err := migrate.NewWithSourceInstance("iofs", d, "sqlite://"+dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create a migrate manager: %w", err)
	}

	m.Log = &migrateLogger{logger}
	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return nil, fmt.Errorf("database migration error: %w", err)
	}

	db, err := sql.Open("sqlite", "file:"+dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create the sqlite db: %w", err)
	}

	return db, nil
}

type migrateLogger struct {
	Logger *slog.Logger
}

func (t *migrateLogger) Printf(format string, v ...any) {
	t.Logger.Debug(fmt.Sprintf(format, v...))
}

func (t *migrateLogger) Verbose() bool {
	return true
}
