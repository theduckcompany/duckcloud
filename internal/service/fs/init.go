package fs

import (
	context "context"
	io "io"

	"github.com/theduckcompany/duckcloud/internal/service/files"
	"github.com/theduckcompany/duckcloud/internal/service/folders"
	"github.com/theduckcompany/duckcloud/internal/service/inodes"
	storage "github.com/theduckcompany/duckcloud/internal/tools/storage"
)

//go:generate mockery --name FS
type FS interface {
	CreateDir(ctx context.Context, name string) (*inodes.INode, error)
	CreateFile(ctx context.Context, name string) (*inodes.INode, error)
	ListDir(ctx context.Context, name string, cmd *storage.PaginateCmd) ([]inodes.INode, error)
	RemoveAll(ctx context.Context, name string) error
	Rename(ctx context.Context, oldName, newName string) error
	Get(ctx context.Context, name string) (*inodes.INode, error)
	Upload(ctx context.Context, inode *inodes.INode, w io.Reader) error
	Download(ctx context.Context, inode *inodes.INode) (io.ReadCloser, error)
}

//go:generate mockery --name Service
type Service interface {
	GetFolderFS(folder *folders.Folder) FS
}

func Init(inodes inodes.Service, files files.Service, folders folders.Service) Service {
	return NewFSService(inodes, files, folders)
}
