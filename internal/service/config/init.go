package config

import (
	"context"

	"github.com/theduckcompany/duckcloud/internal/tools/secret"
	"github.com/theduckcompany/duckcloud/internal/tools/sqlstorage"
)

//go:generate mockery --name Service
type Service interface {
	SetMasterKey(ctx context.Context, key *secret.SealedKey) error
	GetMasterKey(ctx context.Context) (*secret.SealedKey, error)
}

func Init(db sqlstorage.Querier) Service {
	storage := newSqlStorage(db)

	return newService(storage)
}
