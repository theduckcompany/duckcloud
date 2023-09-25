package dav

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/theduckcompany/duckcloud/internal/service/davsessions"
	"github.com/theduckcompany/duckcloud/internal/service/folders"
	"github.com/theduckcompany/duckcloud/internal/service/fs"
	"golang.org/x/net/webdav"
)

type davFS struct {
	folders folders.Service
	fs      fs.Service
}

func (s *davFS) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	session := ctx.Value(sessionKeyCtx).(*davsessions.DavSession)

	name = cleanPath(name)

	// TODO: Handle several folders
	folder, err := s.folders.GetUserFolder(ctx, session.UserID(), session.FoldersIDs()[0])
	if err != nil {
		return fmt.Errorf("failed to folders.GetByID: %w", err)
	}

	_, err = s.fs.GetFolderFS(folder).CreateDir(ctx, name)

	return err
}

func (s *davFS) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	session := ctx.Value(sessionKeyCtx).(*davsessions.DavSession)

	name = cleanPath(name)

	// TODO: Handle several folders
	folder, err := s.folders.GetUserFolder(ctx, session.UserID(), session.FoldersIDs()[0])
	if err != nil {
		return nil, fmt.Errorf("failed to folders.GetByID: %w", err)
	}

	return s.fs.GetFolderFS(folder).OpenFile(ctx, name, flag)
}

func (s *davFS) RemoveAll(ctx context.Context, name string) error {
	session := ctx.Value(sessionKeyCtx).(*davsessions.DavSession)

	name = cleanPath(name)

	// TODO: Handle several folders
	folder, err := s.folders.GetUserFolder(ctx, session.UserID(), session.FoldersIDs()[0])
	if err != nil {
		return fmt.Errorf("failed to folders.GetByID: %w", err)
	}

	return s.fs.GetFolderFS(folder).RemoveAll(ctx, name)
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

	return s.fs.GetFolderFS(folder).Stat(ctx, name)
}

func cleanPath(name string) string {
	return strings.Trim(path.Clean(name), "/")
}
