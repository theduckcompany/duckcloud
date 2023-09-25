package dav

import (
	"context"
	"fmt"
	stdfs "io/fs"
	"os"
	"path"
	"strings"
	"time"

	"github.com/theduckcompany/duckcloud/internal/service/davsessions"
	"github.com/theduckcompany/duckcloud/internal/service/folders"
	"github.com/theduckcompany/duckcloud/internal/service/fs"
	"github.com/theduckcompany/duckcloud/internal/service/inodes"
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

	inode, err := s.fs.GetFolderFS(folder).Get(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to call folderFS.Get: %w", err)
	}

	return &fileInfo{inode}, nil
}

func cleanPath(name string) string {
	return strings.Trim(path.Clean(name), "/")
}

type fileInfo struct {
	inode *inodes.INode
}

func (i *fileInfo) Name() string       { return i.inode.Name() }
func (i *fileInfo) Size() int64        { return i.inode.Size() }
func (i *fileInfo) ModTime() time.Time { return i.inode.ModTime() }
func (i *fileInfo) IsDir() bool        { return i.inode.IsDir() }
func (i *fileInfo) Sys() any           { return nil }
func (i *fileInfo) Mode() os.FileMode {
	if i.inode.IsDir() {
		return 0o660 | stdfs.ModeDir
	}

	return 0o660 // Regular file
}
