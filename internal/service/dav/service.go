package dav

import (
	"context"
	"errors"
	"fmt"
	stdfs "io/fs"
	"os"
	"path"

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
	if !stdfs.ValidPath(name) {
		return &stdfs.PathError{Op: "mkdir", Path: name, Err: stdfs.ErrInvalid}
	}

	session := ctx.Value(sessionKeyCtx).(*davsessions.DavSession)

	name = path.Clean(name)

	// TODO: Handle several folders
	folder, err := s.folders.GetUserFolder(ctx, session.UserID(), session.FoldersIDs()[0])
	if err != nil {
		return fmt.Errorf("failed to folders.GetByID: %w", err)
	}

	_, err = s.fs.GetFolderFS(folder).CreateDir(ctx, name)

	return err
}

func (s *davFS) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	if !stdfs.ValidPath(name) {
		return nil, &stdfs.PathError{Op: "open", Path: name, Err: stdfs.ErrInvalid}
	}

	session := ctx.Value(sessionKeyCtx).(*davsessions.DavSession)

	name = path.Clean(name)

	// TODO: Handle several folders
	folder, err := s.folders.GetUserFolder(ctx, session.UserID(), session.FoldersIDs()[0])
	if err != nil {
		return nil, fmt.Errorf("failed to folders.GetByID: %w", err)
	}

	ffs := s.fs.GetFolderFS(folder)

	info, err := ffs.Get(ctx, name)
	if err != nil && !errors.Is(err, stdfs.ErrNotExist) {
		return nil, fmt.Errorf("failed to fs.Get: %w", err)
	}

	if info != nil && info.IsDir() {
		return NewDirectory(info, ffs, name), nil
	}

	if flag&os.O_EXCL != 0 && info != nil {
		// The flag require that the file doesn't exists but we found one.
		return nil, &stdfs.PathError{Op: "open", Path: name, Err: stdfs.ErrExist}
	}

	if info == nil && flag&os.O_CREATE == 0 {
		// We try to open witout creating a non existing file.
		return nil, &stdfs.PathError{Op: "open", Path: name, Err: stdfs.ErrNotExist}
	}

	// The APPEND flag is not supported yet.
	if flag&(os.O_SYNC|os.O_APPEND) != 0 {
		return nil, fmt.Errorf("%w: O_SYNC and O_APPEND not supported", stdfs.ErrInvalid)
	}

	// At the moment we are only able to write into new files. This situation apprear only at two occations:
	// - the open command have the CREATE flag set and the file doesn't exists
	// - the open command have the TRUNC flag set and the file already exists
	//
	// For all the other cases like APPEND in a existing file or Seek to a position and then write into the file for example are
	// not authorized yet.
	if (flag&os.O_WRONLY != 0 || flag&os.O_RDWR != 0) && info != nil && flag&os.O_TRUNC == 0 {
		return nil, &stdfs.PathError{Op: "open", Path: name, Err: stdfs.ErrInvalid}
	}

	if flag&os.O_TRUNC != 0 {
		err = ffs.RemoveAll(ctx, name)
		if err != nil {
			return nil, fmt.Errorf("failed to RemoveAll for TRUNC: %w", err)
		}

		info = nil
	}

	if info == nil {
		info, err = ffs.CreateFile(ctx, name)
		if err != nil {
			return nil, fmt.Errorf("failed to CreateFile: %w", err)
		}
	}

	return NewFile(info, ffs), nil
}

func (s *davFS) RemoveAll(ctx context.Context, name string) error {
	if !stdfs.ValidPath(name) {
		return &stdfs.PathError{Op: "removeAll", Path: name, Err: stdfs.ErrInvalid}
	}

	session := ctx.Value(sessionKeyCtx).(*davsessions.DavSession)

	name = path.Clean(name)

	// TODO: Handle several folders
	folder, err := s.folders.GetUserFolder(ctx, session.UserID(), session.FoldersIDs()[0])
	if err != nil {
		return fmt.Errorf("failed to folders.GetByID: %w", err)
	}

	return s.fs.GetFolderFS(folder).RemoveAll(ctx, name)
}

func (s *davFS) Rename(ctx context.Context, oldName, newName string) error {
	if !stdfs.ValidPath(oldName) {
		return &stdfs.PathError{Op: "rename", Path: oldName, Err: stdfs.ErrInvalid}
	}

	if !stdfs.ValidPath(newName) {
		return &stdfs.PathError{Op: "rename", Path: newName, Err: stdfs.ErrInvalid}
	}

	// session := ctx.Value(sessionKeyCtx).(*davsessions.DavSession)

	// oldName = path.Clean(oldName)
	// newName = path.Clean(newName)

	return webdav.ErrNotImplemented
}

func (s *davFS) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	if !stdfs.ValidPath(name) {
		return nil, &stdfs.PathError{Op: "stat", Path: name, Err: stdfs.ErrInvalid}
	}

	session := ctx.Value(sessionKeyCtx).(*davsessions.DavSession)

	name = path.Clean(name)

	// TODO: Handle several folders
	folder, err := s.folders.GetUserFolder(ctx, session.UserID(), session.FoldersIDs()[0])
	if err != nil {
		return nil, fmt.Errorf("failed to folders.GetByID: %w", err)
	}

	return s.fs.GetFolderFS(folder).Get(ctx, name)
}
