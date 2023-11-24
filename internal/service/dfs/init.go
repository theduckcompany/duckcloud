package dfs

import (
	"context"
	"database/sql"
	"io"

	"github.com/theduckcompany/duckcloud/internal/service/dfs/internal/inodes"
	"github.com/theduckcompany/duckcloud/internal/service/files"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/runner"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
	"go.uber.org/fx"
)

//go:generate mockery --name FS
type FS interface {
	SpaceID() uuid.UUID
	CreateDir(ctx context.Context, dirPath string) (*INode, error)
	ListDir(ctx context.Context, dirPath string, cmd *storage.PaginateCmd) ([]INode, error)
	Remove(ctx context.Context, path string) error
	Move(ctx context.Context, oldPath, newPath string) error
	Get(ctx context.Context, path string) (*INode, error)
	Upload(ctx context.Context, filePath string, w io.Reader) error
	Download(ctx context.Context, filePath string) (io.ReadSeekCloser, error)
}

//go:generate mockery --name Service
type Service interface {
	GetSpaceFS(spaceID uuid.UUID) FS
	CreateSpaceFS(ctx context.Context, spaceID uuid.UUID) (*INode, error)
	RemoveSpaceFS(ctx context.Context, spaceId uuid.UUID) error
}

type Result struct {
	fx.Out
	Service                      Service
	FSGCTask                     runner.TaskRunner `group:"tasks"`
	FSMoveTask                   runner.TaskRunner `group:"tasks"`
	FSRefreshSizeTask            runner.TaskRunner `group:"tasks"`
	FSRemoveDuplicateFilesRunner runner.TaskRunner `group:"tasks"`
}

func Init(db *sql.DB, files files.Service, scheduler scheduler.Service, tools tools.Tools) (Result, error) {
	inodes := inodes.Init(scheduler, tools, db)

	return Result{
		Service:                      NewFSService(inodes, files, scheduler, tools),
		FSGCTask:                     NewFSGGCTaskRunner(inodes, files, tools),
		FSMoveTask:                   NewFSMoveTaskRunner(inodes, scheduler),
		FSRefreshSizeTask:            NewFSRefreshSizeTaskRunner(inodes, files),
		FSRemoveDuplicateFilesRunner: NewFSRemoveDuplicateFileRunner(inodes, files, scheduler),
	}, nil
}
