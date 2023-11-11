package oauthcodes

import (
	"context"
	"database/sql"

	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/secret"
)

//go:generate mockery --name Service
type Service interface {
	Create(ctx context.Context, input *CreateCmd) error
	RemoveByCode(ctx context.Context, code secret.Text) error
	GetByCode(ctx context.Context, code secret.Text) (*Code, error)
}

func Init(tools tools.Tools, db *sql.DB) Service {
	storage := newSqlStorage(db)

	return NewService(tools, storage)
}
