package runner

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/theduckcompany/duckcloud/internal/service/tasks/internal/storage"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"go.uber.org/fx"
)

//go:generate mockery --name Service
type Service interface {
	RunLoop()
	RunSingleJob(ctx context.Context) error
	Stop()
}

//go:generate mockery --name TaskRunner
type TaskRunner interface {
	Run(ctx context.Context, args json.RawMessage) error
	Name() string
}

func Init(runners []TaskRunner, lc fx.Lifecycle, tools tools.Tools, db *sql.DB) Service {
	storage := storage.NewSqlStorage(db)

	svc := NewService(tools, storage, runners)

	if lc != nil {
		lc.Append(fx.Hook{
			OnStart: func(context.Context) error {
				//nolint:contextcheck // there is no context
				go svc.RunLoop()
				return nil
			},
			OnStop: func(context.Context) error {
				svc.Stop()
				return nil
			},
		})
	}

	return svc
}
