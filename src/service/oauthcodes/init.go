package oauthcodes

import (
	"context"
	"database/sql"

	"github.com/theduckcompany/duckcloud/src/tools"
)

//go:generate mockery --name Service
type Service interface {
	Create(ctx context.Context, input *CreateCmd) error
	RemoveByCode(ctx context.Context, code string) error
	GetByCode(ctx context.Context, code string) (*Code, error)
}

func Init(tools tools.Tools, db *sql.DB) Service {
	storage := newSqlStorage(db)

	return NewService(tools, storage)
}
