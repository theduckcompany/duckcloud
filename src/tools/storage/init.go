package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Peltoche/neurone/src/tools"
	"go.uber.org/fx"
)

type Params struct {
	fx.In
	LC    fx.Lifecycle
	Cfg   Config
	Tools tools.Tools
}

func Init(p Params) (*sql.DB, error) {
	db, err := NewSQliteClient(p.Cfg)
	if err != nil {
		return nil, fmt.Errorf("sqlite error: %w", err)
	}

	p.LC.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return db.PingContext(ctx)
		},
		OnStop: func(context.Context) error {
			return db.Close()
		},
	})

	p.LC.Append(fx.Hook{
		OnStart: func(context.Context) error {
			return RunMigrations(p.Cfg, p.Tools)
		},
	})

	return db, nil
}
