package oauthclients

import (
	"context"
	"database/sql"

	"github.com/Peltoche/neurone/src/tools"
)

//go:generate mockery --name Service
type Service interface {
	Create(ctx context.Context, cmd *CreateCmd) (*Client, error)
	GetByID(ctx context.Context, clientID string) (*Client, error)
}

func Init(tools tools.Tools, db *sql.DB) Service {
	storage := newSqlStorage(db)

	return NewService(tools, storage)
}
