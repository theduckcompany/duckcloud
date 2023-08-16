package dav

import (
	"context"
	"os"

	"github.com/Peltoche/neurone/src/service/blocks"
	"github.com/Peltoche/neurone/src/service/fs"
	"github.com/Peltoche/neurone/src/service/inodes"
	"github.com/Peltoche/neurone/src/tools/uuid"
	"golang.org/x/net/webdav"
)

const (
	currentUser = uuid.UUID("7b44db46-a6b6-44c0-b6d5-d22d11c3bc6a")
	root        = uuid.UUID("338b2c56-aa78-4f0d-bd4a-cec46b7c69b9")
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
