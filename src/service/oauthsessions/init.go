package oauthsessions

import (
	"context"
	"database/sql"

	"github.com/Peltoche/neurone/src/tools"
)

type Service interface {
	CreateSession(ctx context.Context, input *CreateCmd) error
	RemoveByAccessToken(ctx context.Context, access string) error
	RemoveByRefreshToken(ctx context.Context, refresh string) error
	GetByAccessToken(ctx context.Context, access string) (*Session, error)
	GetByRefreshToken(ctx context.Context, refresh string) (*Session, error)
}

func Init(tools tools.Tools, db *sql.DB) Service {
	storage := newSqlStorage(db)

	return NewService(tools, storage)
}
