package storage

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"net/url"

	"github.com/Peltoche/neurone/src/tools/logger"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/mattn/go-sqlite3"
)

type Config struct {
	URL url.URL `mapstructure:"url"`
}

//go:embed db/migration/*.sql
var fs embed.FS

func NewSQliteDBWithMigrate(cfg Config, logger *logger.Logger) (*sql.DB, error) {
	d, err := iofs.New(fs, "db/migration")
	if err != nil {
		return nil, fmt.Errorf("failed to load the migrated files: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, cfg.URL.String())
	if err != nil {
		return nil, fmt.Errorf("failed to create a migrate manager: %w", err)
	}

	m.Log = &migrateLogger{logger}
	err = m.Up()
	if errors.Is(err, migrate.ErrNoChange) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("database migration error: %w", err)
	}

	db, err := sql.Open("sqlite3", cfg.URL.String())
	if err != nil {
		return nil, fmt.Errorf("failed to create the sqlite3 db: %w", err)
	}

	return db, nil
}

type migrateLogger struct {
	Logger *logger.Logger
}

func (t *migrateLogger) Printf(format string, v ...any) {
	t.Logger.Debug(fmt.Sprintf(format, v...))
}

func (t *migrateLogger) Verbose() bool {
	return true
}
