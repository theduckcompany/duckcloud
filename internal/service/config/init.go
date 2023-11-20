package config

import (
	"context"
	"database/sql"

	"github.com/theduckcompany/duckcloud/internal/tools/secret"
)

//go:generate mockery --name Service
type Service interface {
	SetMasterKey(ctx context.Context, key *secret.SealedKey) error
	GetMasterKey(ctx context.Context) (*secret.SealedKey, error)
}

func Init(ctx context.Context, db *sql.DB) Service {
	storage := newSqlStorage(db)

	return NewService(storage)
}
