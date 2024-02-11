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
	SetTotalSize(ctx context.Context, totalSize uint64) error
	GetTotalSize(ctx context.Context) (uint64, error)
}

func Init(db *sql.DB) Service {
	storage := newSqlStorage(db)

	return NewService(storage)
}
