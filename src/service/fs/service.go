package fs

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"

	"github.com/theduckcompany/duckcloud/src/service/files"
	"github.com/theduckcompany/duckcloud/src/service/folders"
	"github.com/theduckcompany/duckcloud/src/service/inodes"
)

var ErrNotImplemented = errors.New("not implemented")

type FSService struct {
	inodes  inodes.Service
	files   files.Service
	folder  *folders.Folder
	folders folders.Service
}

func NewFSService(
	inodes inodes.Service,
	files files.Service,
	folder *folders.Folder,
	folders folders.Service,
) *FSService {
	return &FSService{inodes, files, folder, folders}
}

func (s *FSService) CreateDir(ctx context.Context, name string, perm os.FileMode) error {
	name, err := validatePath(name)
	if err != nil {
		return err
	}

	_, err = s.inodes.CreateDir(ctx, &inodes.PathCmd{
		Root:     s.folder.RootFS(),
		FullName: name,
	})
	if err != nil {
		return fmt.Errorf("inodes mkdir error: %w", err)
	}

	return nil
}

func (s *FSService) Open(name string) (fs.File, error) {
	return s.OpenFile(context.Background(), name, os.O_RDONLY, 0)
}

func (s *FSService) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (FileOrDirectory, error) {
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

	if inode == nil && flag&os.O_CREATE == 0 {
		// We try to open witout creating a non existing file.
		return nil, fs.ErrNotExist
	}

	if inode != nil && inode.Mode().IsDir() {
		return NewDirectory(inode, s.inodes, &pathCmd), nil
	}

	if flag&(os.O_SYNC|os.O_APPEND) != 0 {
		// We doesn't support these flags yet.
		return nil, fmt.Errorf("%w: O_SYNC and O_APPEND not supported", os.ErrInvalid)
	}

	if flag&os.O_EXCL != 0 && inode != nil {
		// The flag require that the file doesn't exists but we found one.
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrExist}
	}

	if inode == nil {
		inode, err = s.createFile(ctx, &pathCmd, perm)
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

func (s *FSService) RemoveAll(ctx context.Context, name string) error {
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

func (s *FSService) Rename(ctx context.Context, oldName, newName string) error {
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

func (s *FSService) Stat(ctx context.Context, name string) (os.FileInfo, error) {
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

func (s *FSService) createFile(ctx context.Context, cmd *inodes.PathCmd, perm fs.FileMode) (*inodes.INode, error) {
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
		Mode:   perm,
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
