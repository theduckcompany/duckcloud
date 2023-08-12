package dav

import (
	"context"
	"fmt"
	"os"

	"github.com/Peltoche/neurone/src/service/fs"
	"github.com/Peltoche/neurone/src/service/inodes"
	"github.com/Peltoche/neurone/src/tools/uuid"
	"golang.org/x/net/webdav"
)

const (
	currentUser = uuid.UUID("291c8167-a6c0-43b5-bd7b-add1d5404ea3")
	root        = uuid.UUID("e09d9b27-53c5-4ba3-8f3d-2a6bb70f55e5")
)

type davFS struct {
	inodes inodes.Service
}

func (s *davFS) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	// username := ctx.Value(usernameKeyCtx)
	// password := ctx.Value(passwordKeyCtx)

	return fs.NewFSService(currentUser, root, s.inodes).CreateDir(ctx, name, perm)
}

func (s *davFS) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	return fs.NewFSService(currentUser, root, s.inodes).OpenFile(ctx, name, flag, perm)
}

func (s *davFS) RemoveAll(ctx context.Context, name string) error {
	return fs.NewFSService(currentUser, root, s.inodes).RemoveAll(ctx, name)
}

func (s *davFS) Rename(ctx context.Context, oldName, newName string) error {
	fmt.Printf("Rename %q -> %q: \n\n", oldName, newName)
	return webdav.ErrNotImplemented
}

func (s *davFS) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	return fs.NewFSService(currentUser, root, s.inodes).Stat(ctx, name)
}
