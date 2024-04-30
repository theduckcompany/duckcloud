package stats

import (
	"context"

	"github.com/theduckcompany/duckcloud/internal/tools/sqlstorage"
)

//go:generate mockery --name Service
type Service interface {
	SetTotalSize(ctx context.Context, totalSize uint64) error
	GetTotalSize(ctx context.Context) (uint64, error)
}

func Init(db sqlstorage.Querier) Service {
	storage := newSqlStorage(db)

	return newService(storage)
}
