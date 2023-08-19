package fs

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/theduckcompany/duckcloud/src/service/files"
	"github.com/theduckcompany/duckcloud/src/service/inodes"
	"github.com/theduckcompany/duckcloud/src/tools"
	"github.com/theduckcompany/duckcloud/src/tools/uuid"
)

var ErrNotImplemented = errors.New("not implemented")

type FSService struct {
	userID uuid.UUID
	root   uuid.UUID
	inodes inodes.Service
	files  files.Service
	uuid   uuid.Service
}

func NewFSService(
	userID uuid.UUID,
	root uuid.UUID,
	inodes inodes.Service,
	files files.Service,
	tools tools.Tools,
) *FSService {
	return &FSService{userID, root, inodes, files, tools.UUID()}
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

func (s *FSService) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (*File, error) {
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
		// We try to open witout creating a non existing file.
		return nil, fs.ErrNotExist
	}

	if res != nil && res.Mode().IsDir() {
		return &File{res, s.inodes, &pathCmd, perm, nil}, nil
	}

	if flag&(os.O_APPEND) != 0 {
		// We doesn't support these flags yet.
		return nil, fmt.Errorf("%w: O_APPEND not supported", os.ErrInvalid)
	}

	if flag&os.O_EXCL != 0 && res != nil {
		// The flag require that the file doesn't exists but we found one.
		return nil, os.ErrExist
	}

	var fileID uuid.UUID
	if res != nil && res.FileID() != nil {
		fileID = *res.FileID()
	} else {
		fileID = s.uuid.New()
	}

	file, err := s.files.Open(ctx, fileID)
	if err != nil {
		return nil, fmt.Errorf("failed to open %q: %w", fileID, err)
	}

	return &File{res, s.inodes, &pathCmd, perm, file}, nil
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
