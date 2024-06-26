package oauthclients

import (
	"context"

	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/sqlstorage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

//go:generate mockery --name Service
type Service interface {
	Create(ctx context.Context, cmd *CreateCmd) (*Client, error)
	GetByID(ctx context.Context, clientID uuid.UUID) (*Client, error)
}

func Init(tools tools.Tools, db sqlstorage.Querier) Service {
	storage := newSqlStorage(db)

	return newService(tools, storage)
}
