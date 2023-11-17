package config

import (
	"context"
	"database/sql"

	"github.com/theduckcompany/duckcloud/internal/tools/secret"
)

//go:generate mockery --name Service
type Service interface {
	SetMasterKey(ctx context.Context, key *secret.Key) error
	GetMasterKey(ctx context.Context) (*secret.Key, error)
}

func Init(db *sql.DB) Service {
	storage := newSqlStorage(db)

	return NewService(storage)
}
