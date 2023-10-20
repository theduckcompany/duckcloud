package dfs

import (
	"context"
	"database/sql"
	"fmt"
	"io"

	"github.com/spf13/afero"
	"github.com/theduckcompany/duckcloud/internal/service/dfs/internal/files"
	"github.com/theduckcompany/duckcloud/internal/service/dfs/internal/inodes"
	"github.com/theduckcompany/duckcloud/internal/service/folders"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/runner"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
	"go.uber.org/fx"
)

//go:generate mockery --name FS
type FS interface {
	Folder() *folders.Folder
	CreateDir(ctx context.Context, dirPath string) (*INode, error)
	ListDir(ctx context.Context, dirPath string, cmd *storage.PaginateCmd) ([]INode, error)
	Remove(ctx context.Context, path string) error
	Rename(ctx context.Context, oldPath, newPath string) error
	Get(ctx context.Context, path string) (*INode, error)
	Upload(ctx context.Context, filePath string, w io.Reader) error
	Download(ctx context.Context, filePath string) (io.ReadSeekCloser, error)
}

type Config struct {
	Path string
}

//go:generate mockery --name Service
type Service interface {
	GetFolderFS(folder *folders.Folder) FS
	CreateFS(ctx context.Context, owners []uuid.UUID) (*folders.Folder, error)
	RemoveFS(ctx context.Context, folder *folders.Folder) error
}

type Result struct {
	fx.Out
	Service        Service
	FileUploadTask runner.TaskRunner `group:"tasks"`
	FSGCTask       runner.TaskRunner `group:"tasks"`
	FSMoveTask     runner.TaskRunner `group:"tasks"`
}

func Init(cfg Config, fs afero.Fs, db *sql.DB, folders folders.Service, scheduler scheduler.Service, tools tools.Tools) (Result, error) {
	inodes := inodes.Init(scheduler, tools, db)
	files, err := files.Init(cfg.Path, fs, tools)
	if err != nil {
		return Result{}, fmt.Errorf("failed to init files: %w", err)
	}

	return Result{
		Service:        NewFSService(inodes, files, folders, scheduler, tools),
		FileUploadTask: NewFileUploadTaskRunner(folders, files, inodes),
		FSGCTask:       NewFSGGCTaskRunner(inodes, files, folders, tools),
		FSMoveTask:     NewFSMoveTaskRunner(inodes, folders),
	}, nil
}
