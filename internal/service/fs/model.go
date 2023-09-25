package fs

import (
	"context"
	"io"
	"net/http"

	"github.com/theduckcompany/duckcloud/internal/service/inodes"
)

//go:generate mockery --name FS
type FS interface {
	CreateDir(ctx context.Context, name string) (*inodes.INode, error)
	OpenFile(ctx context.Context, name string, flag int) (FileOrDirectory, error)
	RemoveAll(ctx context.Context, name string) error
	Rename(ctx context.Context, oldName, newName string) error
	Get(ctx context.Context, name string) (*inodes.INode, error)
}

//go:generate mockery --name FileOrDirectory
type FileOrDirectory interface {
	http.File
	io.Writer
}
