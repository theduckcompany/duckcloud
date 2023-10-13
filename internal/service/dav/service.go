package dav

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"strings"

	"github.com/theduckcompany/duckcloud/internal/service/dav/webdav"
	"github.com/theduckcompany/duckcloud/internal/service/davsessions"
	"github.com/theduckcompany/duckcloud/internal/service/dfs"
	"github.com/theduckcompany/duckcloud/internal/service/folders"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
)

type davFS struct {
	folders folders.Service
	fs      dfs.Service
}

func (s *davFS) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	name, isOk := validatePath(name)
	if !isOk {
		return &fs.PathError{Op: "mkdir", Path: name, Err: fs.ErrInvalid}
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
	name, isOk := validatePath(name)
	if !isOk {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrInvalid}
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
	if err != nil && !errors.Is(err, errs.ErrNotFound) {
		return nil, fmt.Errorf("failed to fs.Get: %w", err)
	}

	if info != nil && info.IsDir() {
		return NewDirectory(info, ffs, name), nil
	}

	if flag&os.O_EXCL != 0 && info != nil {
		// The flag require that the file doesn't exists but we found one.
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrExist}
	}

	if info == nil && flag&os.O_CREATE == 0 {
		// We try to open witout creating a non existing file.
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
	}

	// The APPEND flag is not supported yet.
	if flag&(os.O_SYNC|os.O_APPEND) != 0 {
		return nil, fmt.Errorf("%w: O_SYNC and O_APPEND not supported", fs.ErrInvalid)
	}

	// At the moment we are only able to write into new files. This situation apprear only at two occations:
	// - the open command have the CREATE flag set and the file doesn't exists
	// - the open command have the TRUNC flag set and the file already exists
	//
	// For all the other cases like APPEND in a existing file or Seek to a position and then write into the file for example are
	// not authorized yet.
	if (flag&os.O_WRONLY != 0 || flag&os.O_RDWR != 0) && info != nil && flag&os.O_TRUNC == 0 {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrInvalid}
	}

	if flag&os.O_TRUNC != 0 {
		err = ffs.RemoveAll(ctx, name)
		if err != nil {
			return nil, fmt.Errorf("failed to RemoveAll for TRUNC: %w", err)
		}

		info = nil
	}

	return NewFile(name, ffs), nil
}

func (s *davFS) RemoveAll(ctx context.Context, name string) error {
	name, isOk := validatePath(name)
	if !isOk {
		return &fs.PathError{Op: "removeAll", Path: name, Err: fs.ErrInvalid}
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
	_, isOk := validatePath(oldName)
	if !isOk {
		return &fs.PathError{Op: "rename", Path: oldName, Err: fs.ErrInvalid}
	}

	_, isOk = validatePath(newName)
	if !isOk {
		return &fs.PathError{Op: "rename", Path: newName, Err: fs.ErrInvalid}
	}

	// session := ctx.Value(sessionKeyCtx).(*davsessions.DavSession)

	// oldName = path.Clean(oldName)
	// newName = path.Clean(newName)

	return webdav.ErrNotImplemented
}

func (s *davFS) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	name, isOk := validatePath(name)
	if !isOk {
		return nil, &fs.PathError{Op: "stat", Path: name, Err: fs.ErrInvalid}
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

func validatePath(name string) (string, bool) {
	newName := strings.Trim(name, "/")

	if newName == "" {
		newName = "."
	}

	if !fs.ValidPath(newName) {
		return name, false
	}

	return path.Clean(newName), true
}
