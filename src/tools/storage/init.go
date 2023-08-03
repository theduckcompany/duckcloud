package storage

import (
	"context"
	"database/sql"
	"fmt"

	"go.uber.org/fx"
)

func Init(lc fx.Lifecycle, cfg Config) (*sql.DB, error) {
	db, err := NewSQliteClient(cfg)
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
