package scheduler

import (
	"context"
	"database/sql"

	"github.com/theduckcompany/duckcloud/internal/service/tasks/internal/storage"
	"github.com/theduckcompany/duckcloud/internal/tools"
)

//go:generate mockery --name Service
type Service interface {
	RegisterFileUploadTask(ctx context.Context, args *FileUploadArgs) error
	RegisterUserCreateTask(ctx context.Context, args *UserCreateArgs) error
	RegisterUserDeleteTask(ctx context.Context, args *UserDeleteArgs) error
}

func Init(db *sql.DB, tools tools.Tools) Service {
	storage := storage.NewSqlStorage(db)

	return NewService(storage, tools)
}
