package fs

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/Peltoche/neurone/src/service/inodes"
	"github.com/Peltoche/neurone/src/tools/uuid"
)

var ErrNotImplemented = errors.New("not implemented")

type FSService struct {
	userID uuid.UUID
	root   uuid.UUID
	inodes inodes.Service
}

func NewFSService(userID uuid.UUID, root uuid.UUID, inodes inodes.Service) *FSService {
	return &FSService{userID, root, inodes}
}

func (s *FSService) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	if name == "" {
		name = "/"
	}

	_, err := s.inodes.Mkdir(ctx, &inodes.PathCmd{
		Root:     s.root,
		UserID:   s.userID,
		FullName: name,
	})
	if err != nil {
		return fmt.Errorf("inodes mkdir error: %w", err)
	}

	return nil
}

func (s *FSService) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (File, error) {
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

	if res == nil {
		return nil, fs.ErrNotExist
	}

	if res.Mode().IsDir() {
		return &Directory{res, s.inodes, &pathCmd}, nil
	}

	return nil, ErrNotImplemented
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
	fmt.Printf("Rename %q -> %q: \n\n", oldName, newName)
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
