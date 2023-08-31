package dav

import (
	"context"
	"os"
	"path"
	"strings"

	"github.com/theduckcompany/duckcloud/src/service/davsessions"
	"github.com/theduckcompany/duckcloud/src/service/files"
	"github.com/theduckcompany/duckcloud/src/service/fs"
	"github.com/theduckcompany/duckcloud/src/service/inodes"
	"golang.org/x/net/webdav"
)

type davFS struct {
	inodes inodes.Service
	files  files.Service
}

func (s *davFS) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	session := ctx.Value(sessionKeyCtx).(*davsessions.DavSession)

	name = cleanPath(name)
	rootFS := session.FoldersIDs()[0] // TODO: Handle several folders

	return fs.NewFSService(session.UserID(), rootFS, s.inodes, s.files).CreateDir(ctx, name, perm)
}

func (s *davFS) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	session := ctx.Value(sessionKeyCtx).(*davsessions.DavSession)

	name = cleanPath(name)
	rootFS := session.FoldersIDs()[0] // TODO: Handle several folders

	return fs.NewFSService(session.UserID(), rootFS, s.inodes, s.files).OpenFile(ctx, name, flag, perm)
}

func (s *davFS) RemoveAll(ctx context.Context, name string) error {
	session := ctx.Value(sessionKeyCtx).(*davsessions.DavSession)

	name = cleanPath(name)
	rootFS := session.FoldersIDs()[0] // TODO: Handle several folders

	return fs.NewFSService(session.UserID(), rootFS, s.inodes, s.files).RemoveAll(ctx, name)
}

func (s *davFS) Rename(ctx context.Context, oldName, newName string) error {
	// session := ctx.Value(sessionKeyCtx).(*davsessions.DavSession)

	// oldName = cleanPath(oldName)
	// newName = cleanPath(newName)

	return webdav.ErrNotImplemented
}

func (s *davFS) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	session := ctx.Value(sessionKeyCtx).(*davsessions.DavSession)

	name = cleanPath(name)
	rootFS := session.FoldersIDs()[0] // TODO: Handle several folders

	return fs.NewFSService(session.UserID(), rootFS, s.inodes, s.files).Stat(ctx, name)
}

func cleanPath(name string) string {
	return strings.Trim(path.Clean(name), "/")
}
