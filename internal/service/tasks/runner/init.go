package runner

import (
	"context"
	"encoding/json"

	"github.com/theduckcompany/duckcloud/internal/service/tasks/internal/taskstorage"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/sqlstorage"
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

func Init(runners []TaskRunner, tools tools.Tools, db sqlstorage.Querier) Service {
	storage := taskstorage.NewSqlStorage(db)

	return NewService(tools, storage, runners)
}
