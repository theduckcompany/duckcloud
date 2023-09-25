package dav

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/theduckcompany/duckcloud/src/service/davsessions"
	"github.com/theduckcompany/duckcloud/src/service/files"
	"github.com/theduckcompany/duckcloud/src/service/folders"
	"github.com/theduckcompany/duckcloud/src/service/fs"
	"github.com/theduckcompany/duckcloud/src/service/inodes"
	"golang.org/x/net/webdav"
)

type davFS struct {
	inodes  inodes.Service
	files   files.Service
	folders folders.Service
}

func (s *davFS) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	session := ctx.Value(sessionKeyCtx).(*davsessions.DavSession)

	name = cleanPath(name)

	// TODO: Handle several folders
	folder, err := s.folders.GetUserFolder(ctx, session.UserID(), session.FoldersIDs()[0])
	if err != nil {
		return fmt.Errorf("failed to folders.GetByID: %w", err)
	}

	return fs.NewFSService(s.inodes, s.files, folder, s.folders).CreateDir(ctx, name)
}

func (s *davFS) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	session := ctx.Value(sessionKeyCtx).(*davsessions.DavSession)

	name = cleanPath(name)

	// TODO: Handle several folders
	folder, err := s.folders.GetUserFolder(ctx, session.UserID(), session.FoldersIDs()[0])
	if err != nil {
		return nil, fmt.Errorf("failed to folders.GetByID: %w", err)
	}

	return fs.NewFSService(s.inodes, s.files, folder, s.folders).OpenFile(ctx, name, flag)
}

func (s *davFS) RemoveAll(ctx context.Context, name string) error {
	session := ctx.Value(sessionKeyCtx).(*davsessions.DavSession)

	name = cleanPath(name)

	// TODO: Handle several folders
	folder, err := s.folders.GetUserFolder(ctx, session.UserID(), session.FoldersIDs()[0])
	if err != nil {
		return fmt.Errorf("failed to folders.GetByID: %w", err)
	}

	return fs.NewFSService(s.inodes, s.files, folder, s.folders).RemoveAll(ctx, name)
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

	// TODO: Handle several folders
	folder, err := s.folders.GetUserFolder(ctx, session.UserID(), session.FoldersIDs()[0])
	if err != nil {
		return nil, fmt.Errorf("failed to folders.GetByID: %w", err)
	}

	return fs.NewFSService(s.inodes, s.files, folder, s.folders).Stat(ctx, name)
}

func cleanPath(name string) string {
	return strings.Trim(path.Clean(name), "/")
}
