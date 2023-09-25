package fs

import (
	"context"
	"io"
	"net/http"
	"os"

	"github.com/theduckcompany/duckcloud/internal/service/inodes"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
)

//go:generate mockery --name FS
type FS interface {
	CreateDir(ctx context.Context, name string) (*inodes.INode, error)
	ReadDir(ctx context.Context, name string, cmd *storage.PaginateCmd) ([]inodes.INode, error)
	OpenFile(ctx context.Context, name string, flag int) (FileOrDirectory, error)
	RemoveAll(ctx context.Context, name string) error
	Rename(ctx context.Context, oldName, newName string) error
	Stat(ctx context.Context, name string) (os.FileInfo, error)
}

//go:generate mockery --name FileOrDirectory
type FileOrDirectory interface {
	http.File
	io.Writer
}
