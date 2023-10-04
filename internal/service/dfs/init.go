package dfs

import (
	"context"
	"database/sql"
	"io"

	"github.com/theduckcompany/duckcloud/internal/service/dfs/uploads"
	"github.com/theduckcompany/duckcloud/internal/service/files"
	"github.com/theduckcompany/duckcloud/internal/service/folders"
	"github.com/theduckcompany/duckcloud/internal/service/inodes"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
)

//go:generate mockery --name FS
type FS interface {
	Folder() *folders.Folder
	CreateDir(ctx context.Context, name string) (*inodes.INode, error)
	ListDir(ctx context.Context, name string, cmd *storage.PaginateCmd) ([]inodes.INode, error)
	RemoveAll(ctx context.Context, name string) error
	Rename(ctx context.Context, oldName, newName string) error
	Get(ctx context.Context, name string) (*inodes.INode, error)
	Upload(ctx context.Context, name string, w io.Reader) error
	Download(ctx context.Context, name string) (io.ReadSeekCloser, error)
}

//go:generate mockery --name Service
type Service interface {
	GetFolderFS(folder *folders.Folder) FS
}

func Init(db *sql.DB, inodes inodes.Service, files files.Service, folders folders.Service, tools tools.Tools) Service {
	uploads := uploads.Init(db, tools)

	return NewFSService(inodes, files, folders, uploads)
}
