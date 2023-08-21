package dav

import (
	"context"
	"os"

	"github.com/theduckcompany/duckcloud/src/service/files"
	"github.com/theduckcompany/duckcloud/src/service/fs"
	"github.com/theduckcompany/duckcloud/src/service/inodes"
	"github.com/theduckcompany/duckcloud/src/service/users"
	"golang.org/x/net/webdav"
)

type davFS struct {
	inodes inodes.Service
	files  files.Service
}

func (s *davFS) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	user := ctx.Value(userKeyCtx).(*users.User)

	return fs.NewFSService(user.ID(), user.RootFS(), s.inodes, s.files).CreateDir(ctx, name, perm)
}

func (s *davFS) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	user := ctx.Value(userKeyCtx).(*users.User)

	return fs.NewFSService(user.ID(), user.RootFS(), s.inodes, s.files).OpenFile(ctx, name, flag, perm)
}

func (s *davFS) RemoveAll(ctx context.Context, name string) error {
	user := ctx.Value(userKeyCtx).(*users.User)

	return fs.NewFSService(user.ID(), user.RootFS(), s.inodes, s.files).RemoveAll(ctx, name)
}

func (s *davFS) Rename(ctx context.Context, oldName, newName string) error {
	// user := ctx.Value(userKeyCtx).(*users.User)

	return webdav.ErrNotImplemented
}

func (s *davFS) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	user := ctx.Value(userKeyCtx).(*users.User)

	return fs.NewFSService(user.ID(), user.RootFS(), s.inodes, s.files).Stat(ctx, name)
}
