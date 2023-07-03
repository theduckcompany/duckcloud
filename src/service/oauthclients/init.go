package oauthclients

import (
	"context"
	"database/sql"

	"github.com/Peltoche/neurone/src/tools"
	"go.uber.org/fx"
)

type Service interface {
	Create(ctx context.Context, cmd *CreateCmd) error
	GetByID(ctx context.Context, clientID string) (*Client, error)
}

func Init(lc fx.Lifecycle, tools tools.Tools, db *sql.DB) Service {
	storage := newSqlStorage(db)

	svc := NewService(tools, storage)

	// lc.Append(fx.Hook{OnStart: svc.BootstrapWebApp})

	return svc
}
