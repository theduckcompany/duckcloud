package oauthclients

import (
	"context"
	"database/sql"

	"github.com/Peltoche/neurone/src/tools"
	"github.com/Peltoche/neurone/src/tools/uuid"
)

type Service interface {
	BootstrapWebApp(ctx context.Context) error
	GetByID(ctx context.Context, uuid uuid.UUID) (*Client, error)
}

func Init(tools tools.Tools, db *sql.DB) Service {
	storage := newSqlStorage(db)

	return NewService(tools, storage)
}
