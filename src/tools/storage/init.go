package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Peltoche/neurone/src/tools"
	"github.com/spf13/afero"
	"go.uber.org/fx"
)

func Init(lc fx.Lifecycle, fs afero.Fs, cfg Config, tools tools.Tools) (*sql.DB, error) {
	err := fs.MkdirAll(filepath.Dir(cfg.Path), 0o700)
	if err != nil && !errors.Is(err, os.ErrExist) {
		return nil, fmt.Errorf("failed to create the database directory: %w", err)
	}

	db, err := NewSQliteClient(cfg, tools.Logger())
	if err != nil {
		return nil, fmt.Errorf("sqlite error: %w", err)
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := db.PingContext(ctx); err != nil {
				return fmt.Errorf("db unreachable: %w", err)
			}
			return nil
		},
		OnStop: func(context.Context) error {
			return db.Close()
		},
	})

	return db, nil
}
