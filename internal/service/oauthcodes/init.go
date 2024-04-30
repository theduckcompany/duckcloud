package oauthcodes

import (
	"context"

	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/secret"
	"github.com/theduckcompany/duckcloud/internal/tools/sqlstorage"
)

//go:generate mockery --name Service
type Service interface {
	Create(ctx context.Context, input *CreateCmd) error
	RemoveByCode(ctx context.Context, code secret.Text) error
	GetByCode(ctx context.Context, code secret.Text) (*Code, error)
}

func Init(tools tools.Tools, db sqlstorage.Querier) Service {
	storage := newSqlStorage(db)

	return newService(tools, storage)
}
