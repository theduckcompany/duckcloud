package stats

import (
	"context"
	"database/sql"
)

//go:generate mockery --name Service
type Service interface {
	SetTotalSize(ctx context.Context, totalSize uint64) error
	GetTotalSize(ctx context.Context) (uint64, error)
}

func Init(db *sql.DB) Service {
	storage := newSqlStorage(db)

	return NewService(storage)
}
