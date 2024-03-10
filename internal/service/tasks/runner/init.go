package runner

import (
	"context"
	"database/sql"
	"encoding/json"

	sqlstorage "github.com/theduckcompany/duckcloud/internal/service/tasks/internal/storage"
	"github.com/theduckcompany/duckcloud/internal/tools"
)

//go:generate mockery --name Service
type Service interface {
	Run(ctx context.Context) error
}

//go:generate mockery --name TaskRunner
type TaskRunner interface {
	Run(ctx context.Context, args json.RawMessage) error
	Name() string
}

func Init(runners []TaskRunner, tools tools.Tools, db *sql.DB) Service {
	storage := sqlstorage.NewSqlStorage(db)

	return NewService(tools, storage, runners)
}
