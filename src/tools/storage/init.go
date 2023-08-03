package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Peltoche/neurone/src/tools"
	"go.uber.org/fx"
)

func Init(lc fx.Lifecycle, cfg Config, tools tools.Tools) (*sql.DB, error) {
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
