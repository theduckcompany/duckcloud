package dav

import (
	"context"
	"os"

	"github.com/theduckcompany/duckcloud/src/service/files"
	"github.com/theduckcompany/duckcloud/src/service/fs"
	"github.com/theduckcompany/duckcloud/src/service/inodes"
	"github.com/theduckcompany/duckcloud/src/tools/uuid"
	"golang.org/x/net/webdav"
)

const (
	currentUser = uuid.UUID("a6e4082a-124f-403b-92b4-de3253a908a4")
	root        = uuid.UUID("cdd75a48-c4e4-468f-be3c-b171c028e281")
)

type davFS struct {
	inodes inodes.Service
	files  files.Service
}

func (s *davFS) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	// username := ctx.Value(usernameKeyCtx)
	// password := ctx.Value(passwordKeyCtx)

	return fs.NewFSService(currentUser, root, s.inodes, s.files).CreateDir(ctx, name, perm)
}

func (s *davFS) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	return fs.NewFSService(currentUser, root, s.inodes, s.files).OpenFile(ctx, name, flag, perm)
}

func (s *davFS) RemoveAll(ctx context.Context, name string) error {
	return fs.NewFSService(currentUser, root, s.inodes, s.files).RemoveAll(ctx, name)
}

func (s *davFS) Rename(ctx context.Context, oldName, newName string) error {
	return webdav.ErrNotImplemented
}

func (s *davFS) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	return fs.NewFSService(currentUser, root, s.inodes, s.files).Stat(ctx, name)
}
