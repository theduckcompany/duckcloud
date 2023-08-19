package dav

import (
	"context"
	"os"

	"github.com/theduckcompany/duckcloud/src/service/blocks"
	"github.com/theduckcompany/duckcloud/src/service/fs"
	"github.com/theduckcompany/duckcloud/src/service/inodes"
	"github.com/theduckcompany/duckcloud/src/tools/uuid"
	"golang.org/x/net/webdav"
)

const (
	currentUser = uuid.UUID("cdc4fa93-cd92-44a8-9d85-56fb2c28e84c")
	root        = uuid.UUID("e891a010-254e-457e-8206-f282e130802a")
)

type davFS struct {
	inodes inodes.Service
	blocks blocks.Service
}

func (s *davFS) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	// username := ctx.Value(usernameKeyCtx)
	// password := ctx.Value(passwordKeyCtx)

	return fs.NewFSService(currentUser, root, s.inodes, s.blocks).CreateDir(ctx, name, perm)
}

func (s *davFS) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	return fs.NewFSService(currentUser, root, s.inodes, s.blocks).OpenFile(ctx, name, flag, perm)
}

func (s *davFS) RemoveAll(ctx context.Context, name string) error {
	return fs.NewFSService(currentUser, root, s.inodes, s.blocks).RemoveAll(ctx, name)
}

func (s *davFS) Rename(ctx context.Context, oldName, newName string) error {
	return webdav.ErrNotImplemented
}

func (s *davFS) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	return fs.NewFSService(currentUser, root, s.inodes, s.blocks).Stat(ctx, name)
}
