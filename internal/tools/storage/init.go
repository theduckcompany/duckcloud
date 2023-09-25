package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
)

func Init(fs afero.Fs, cfg *Config, logger *slog.Logger) (*sql.DB, error) {
	err := fs.MkdirAll(filepath.Dir(cfg.Path), 0o700)
	if err != nil && !errors.Is(err, os.ErrExist) {
		return nil, fmt.Errorf("failed to create the database directory: %w", err)
	}

	db, err := NewSQliteClient(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("sqlite error: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("db unreachable: %w", err)
	}

	return db, nil
}
