package scheduler

import (
	"context"

	"github.com/theduckcompany/duckcloud/internal/service/tasks/internal/taskstorage"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/sqlstorage"
)

//go:generate mockery --name Service
type Service interface {
	Run(ctx context.Context) error
	RegisterFileUploadTask(ctx context.Context, args *FileUploadArgs) error
	RegisterFSMoveTask(ctx context.Context, args *FSMoveArgs) error
	RegisterUserCreateTask(ctx context.Context, args *UserCreateArgs) error
	RegisterUserDeleteTask(ctx context.Context, args *UserDeleteArgs) error
	RegisterFSRefreshSizeTask(ctx context.Context, args *FSRefreshSizeArg) error
	RegisterFSRemoveDuplicateFile(ctx context.Context, args *FSRemoveDuplicateFileArgs) error
	RegisterSpaceCreateTask(ctx context.Context, args *SpaceCreateArgs) error
}

func Init(db sqlstorage.Querier, tools tools.Tools) Service {
	storage := taskstorage.NewSqlStorage(db)

	return NewService(storage, tools)
}
