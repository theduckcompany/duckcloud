package fs

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path"

	"github.com/theduckcompany/duckcloud/internal/service/files"
	"github.com/theduckcompany/duckcloud/internal/service/folders"
	"github.com/theduckcompany/duckcloud/internal/service/inodes"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
)

type LocalFS struct {
	inodes  inodes.Service
	files   files.Service
	folder  *folders.Folder
	folders folders.Service
}

func newLocalFS(
	inodes inodes.Service,
	files files.Service,
	folder *folders.Folder,
	folders folders.Service,
) *LocalFS {
	return &LocalFS{inodes, files, folder, folders}
}

func (s *LocalFS) CreateDir(ctx context.Context, name string) (*inodes.INode, error) {
	name, err := validatePath(name)
	if err != nil {
		return nil, err
	}

	inode, err := s.inodes.MkdirAll(ctx, &inodes.PathCmd{
		Root:     s.folder.RootFS(),
		FullName: name,
	})
	if err != nil {
		return nil, fmt.Errorf("inodes mkdir error: %w", err)
	}

	return inode, nil
}

func (s *LocalFS) ReadDir(ctx context.Context, name string, cmd *storage.PaginateCmd) ([]inodes.INode, error) {
	return s.inodes.Readdir(ctx, &inodes.PathCmd{
		Root:     s.folder.RootFS(),
		FullName: name,
	}, cmd)
}

func (s *LocalFS) OpenFile(ctx context.Context, name string, flag int) (FileOrDirectory, error) {
	name, err := validatePath(name)
	if err != nil {
		return nil, err
	}

	pathCmd := inodes.PathCmd{
		Root:     s.folder.RootFS(),
		FullName: name,
	}

	inode, err := s.inodes.Get(ctx, &pathCmd)
	if err != nil {
		return nil, fmt.Errorf("failed to open inodes: %w", err)
	}

	if inode != nil && inode.Mode().IsDir() {
		return NewDirectory(inode, s.inodes, &pathCmd), nil
	}

	if flag&os.O_EXCL != 0 && inode != nil {
		// The flag require that the file doesn't exists but we found one.
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrExist}
	}

	if inode == nil && flag&os.O_CREATE == 0 {
		// We try to open witout creating a non existing file.
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
	}

	// The APPEND flag is not supported yet.
	if flag&(os.O_SYNC|os.O_APPEND) != 0 {
		return nil, fmt.Errorf("%w: O_SYNC and O_APPEND not supported", os.ErrInvalid)
	}

	// At the moment we are only able to write into new files. This situation apprear only at two occations:
	// - the open command have the CREATE flag set and the file doesn't exists
	// - the open command have the TRUNC flag set and the file already exists
	//
	// For all the other cases like APPEND in a existing file or Seek to a position and then write into the file for example are
	// not authorized yet.
	if (flag&os.O_WRONLY != 0 || flag&os.O_RDWR != 0) && ((inode != nil && flag&os.O_TRUNC == 0) || (inode == nil && flag&os.O_CREATE == 0)) {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrInvalid}
	}

	if flag&os.O_TRUNC != 0 {
		err = s.RemoveAll(ctx, name)
		if err != nil {
			return nil, fmt.Errorf("failed to RemoveAll for TRUNC: %w", err)
		}

		inode = nil
	}

	if inode == nil {
		inode, err = s.createFile(ctx, &pathCmd)
		if err != nil {
			return nil, fmt.Errorf("failed to createFile: %w", err)
		}
	}

	file, err := s.files.Open(ctx, inode.ID())
	if err != nil {
		return nil, fmt.Errorf("failed to Open the file %q: %w", inode.ID(), err)
	}

	return NewFile(inode, s.inodes, s.files, s.folder.ID(), s.folders, &pathCmd, file), nil
}

func (s *LocalFS) RemoveAll(ctx context.Context, name string) error {
	name, err := validatePath(name)
	if err != nil {
		return err
	}

	err = s.inodes.RemoveAll(ctx, &inodes.PathCmd{
		Root:     s.folder.RootFS(),
		FullName: name,
	})
	if err != nil {
		return fmt.Errorf("failed to RemoveAll: %w", err)
	}

	return nil
}

func (s *LocalFS) Rename(ctx context.Context, oldName, newName string) error {
	_, err := validatePath(oldName)
	if err != nil {
		return err
	}

	_, err = validatePath(newName)
	if err != nil {
		return err
	}

	return ErrNotImplemented
}

func (s *LocalFS) Get(ctx context.Context, name string) (*inodes.INode, error) {
	name, err := validatePath(name)
	if err != nil {
		return nil, err
	}

	res, err := s.inodes.Get(ctx, &inodes.PathCmd{
		Root:     s.folder.RootFS(),
		FullName: name,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open inodes: %w", err)
	}

	if res == nil {
		return nil, &fs.PathError{Op: "stat", Path: name, Err: fs.ErrNotExist}
	}

	return res, nil
}

func (s *LocalFS) createFile(ctx context.Context, cmd *inodes.PathCmd) (*inodes.INode, error) {
	dir, fileName := path.Split(cmd.FullName)
	if dir == "" {
		dir = "/"
	}

	parent, err := s.inodes.Get(ctx, &inodes.PathCmd{
		Root:     cmd.Root,
		FullName: dir,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to Get: %w", err)
	}

	file, err := s.inodes.CreateFile(ctx, &inodes.CreateFileCmd{
		Parent: parent.ID(),
		Name:   fileName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to CreateFile: %w", err)
	}

	return file, nil
}

func validatePath(name string) (string, error) {
	if name == "" {
		return ".", nil
	}

	if !fs.ValidPath(name) {
		return "", &fs.PathError{Op: "open", Path: name, Err: fs.ErrInvalid}
	}

	return path.Clean(name), nil
}
