package fs

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"

	"github.com/myminicloud/myminicloud/src/service/blocks"
	"github.com/myminicloud/myminicloud/src/service/inodes"
	"github.com/myminicloud/myminicloud/src/tools/uuid"
)

var ErrNotImplemented = errors.New("not implemented")

type FSService struct {
	userID uuid.UUID
	root   uuid.UUID
	inodes inodes.Service
	blocks blocks.Service
}

func NewFSService(userID uuid.UUID, root uuid.UUID, inodes inodes.Service, blocks blocks.Service) *FSService {
	return &FSService{userID, root, inodes, blocks}
}

func (s *FSService) CreateDir(ctx context.Context, name string, perm os.FileMode) error {
	if name == "" {
		name = "/"
	}

	_, err := s.inodes.CreateDir(ctx, &inodes.PathCmd{
		Root:     s.root,
		UserID:   s.userID,
		FullName: name,
	})
	if err != nil {
		return fmt.Errorf("inodes mkdir error: %w", err)
	}

	return nil
}

func (s *FSService) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (*Dir, error) {
	if name == "" {
		name = "/"
	}

	pathCmd := inodes.PathCmd{
		Root:     s.root,
		UserID:   s.userID,
		FullName: name,
	}

	res, err := s.inodes.Get(ctx, &pathCmd)
	if err != nil {
		return nil, fmt.Errorf("failed to open inodes: %w", err)
	}

	if res == nil && flag&os.O_CREATE == 0 {
		// We try to open a non existing file without creation permission.
		return nil, fs.ErrNotExist
	}

	if res != nil && res.Mode().IsDir() {
		return &Dir{res, s.inodes, &pathCmd}, nil
	}

	if flag&(os.O_SYNC|os.O_APPEND) != 0 {
		// We doesn't support these flags yet.
		return nil, fmt.Errorf("%w: O_SYNC and O_APPEND not supported", os.ErrInvalid)
	}

	if flag&os.O_EXCL != 0 && res != nil {
		// The flag require that the file doesn't exists but we found one.
		return nil, os.ErrExist
	}

	var file Dir
	if res == nil {
		inode, err := s.createFile(ctx, &pathCmd, perm)
		if err != nil {
			return nil, fmt.Errorf("failed to createFile: %w", err)
		}

		// The file doesnt exists but we have the create flag.
		file = Dir{inode, s.inodes, &pathCmd}

		return &file, nil
	}

	return &Dir{res, s.inodes, &pathCmd}, nil
}

func (s *FSService) RemoveAll(ctx context.Context, name string) error {
	if name == "" {
		name = "/"
	}

	err := s.inodes.RemoveAll(ctx, &inodes.PathCmd{
		Root:     s.root,
		UserID:   s.userID,
		FullName: name,
	})
	if err != nil {
		return fmt.Errorf("failed to RemoveAll: %w", err)
	}

	return nil
}

func (s *FSService) Rename(ctx context.Context, oldName, newName string) error {
	return ErrNotImplemented
}

func (s *FSService) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	if name == "" {
		name = "/"
	}

	res, err := s.inodes.Get(ctx, &inodes.PathCmd{
		Root:     s.root,
		UserID:   s.userID,
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
		UserID:   cmd.UserID,
		FullName: dir,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to Get: %w", err)
	}

	file, err := s.inodes.CreateFile(ctx, &inodes.CreateFileCmd{
		Parent: parent.ID(),
		UserID: cmd.UserID,
		Mode:   perm,
		Name:   fileName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to CreateFile: %w", err)
	}

	return file, nil
}
