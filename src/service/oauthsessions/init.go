package oauthsessions

import (
	"context"
	"database/sql"

	"github.com/theduckcompany/duckcloud/src/tools"
)

//go:generate mockery --name Service
type Service interface {
	Create(ctx context.Context, input *CreateCmd) (*Session, error)
	RemoveByAccessToken(ctx context.Context, access string) error
	RemoveByRefreshToken(ctx context.Context, refresh string) error
	GetByAccessToken(ctx context.Context, access string) (*Session, error)
	GetByRefreshToken(ctx context.Context, refresh string) (*Session, error)
}

func Init(tools tools.Tools, db *sql.DB) Service {
	storage := newSqlStorage(db)

	return NewService(tools, storage)
}
