package fs

import (
	"context"
	"io"
	"io/fs"
	"net/http"
	"os"
)

//go:generate mockery --name FS
type FS interface {
	CreateDir(ctx context.Context, name string) error
	Open(name string) (fs.File, error)
	OpenFile(ctx context.Context, name string, flag int) (FileOrDirectory, error)
	RemoveAll(ctx context.Context, name string) error
	Rename(ctx context.Context, oldName, newName string) error
	Stat(ctx context.Context, name string) (os.FileInfo, error)
}

//go:generate mockery --name FileOrDirectory
type FileOrDirectory interface {
	http.File
	io.Writer
	fs.ReadDirFile
}
