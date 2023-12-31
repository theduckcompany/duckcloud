package dfs

import (
	"context"
	"database/sql"
	"io"

	"github.com/theduckcompany/duckcloud/internal/service/files"
	"github.com/theduckcompany/duckcloud/internal/service/spaces"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/runner"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
	"go.uber.org/fx"
)

//go:generate mockery --name Service
type Service interface {
	Destroy(ctx context.Context, space *spaces.Space) error
	CreateFS(ctx context.Context, user *users.User, owners []uuid.UUID) (*spaces.Space, error)
	CreateDir(ctx context.Context, cmd *CreateDirCmd) (*INode, error)
	ListDir(ctx context.Context, cmd *PathCmd, paginateCmd *storage.PaginateCmd) ([]INode, error)
	Remove(ctx context.Context, cmd *PathCmd) error
	Rename(ctx context.Context, inode *INode, newName string) (*INode, error)
	Move(ctx context.Context, cmd *MoveCmd) error
	Get(ctx context.Context, cmd *PathCmd) (*INode, error)
	Upload(ctx context.Context, cmd *UploadCmd) error
	Download(ctx context.Context, cmd *PathCmd) (io.ReadSeekCloser, error)
	createDir(ctx context.Context, createdBy *users.User, parent *INode, name string) (*INode, error)
	removeINode(ctx context.Context, inode *INode) error
}

type Result struct {
	fx.Out
	Service                      Service
	FSGCTask                     runner.TaskRunner `group:"tasks"`
	FSMoveTask                   runner.TaskRunner `group:"tasks"`
	FSRefreshSizeTask            runner.TaskRunner `group:"tasks"`
	FSRemoveDuplicateFilesRunner runner.TaskRunner `group:"tasks"`
}

func Init(db *sql.DB, spaces spaces.Service, files files.Service, scheduler scheduler.Service, users users.Service, tools tools.Tools) (Result, error) {
	storage := newSqlStorage(db, tools)
	svc := NewService(storage, files, spaces, scheduler, tools)

	return Result{
		Service:                      svc,
		FSGCTask:                     NewFSGGCTaskRunner(storage, files, spaces, tools),
		FSMoveTask:                   NewFSMoveTaskRunner(svc, storage, spaces, users, scheduler),
		FSRefreshSizeTask:            NewFSRefreshSizeTaskRunner(storage, files),
		FSRemoveDuplicateFilesRunner: NewFSRemoveDuplicateFileRunner(storage, files, scheduler),
	}, nil
}
