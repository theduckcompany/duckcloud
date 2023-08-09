package dav

import (
	"context"
	"fmt"
	"io/fs"
	"os"

	"github.com/Peltoche/neurone/src/service/inodes"
	"github.com/Peltoche/neurone/src/tools/uuid"
	"golang.org/x/net/webdav"
)

const (
	currentUser = uuid.UUID("e09c8c33-2cdc-44d9-8802-01126ae50fa1")
	root        = uuid.UUID("e677de9f-c08a-4a90-a780-63d419cd1218")
)

type davKeyCtx string

var (
	usernameKeyCtx davKeyCtx = "username"
	passwordKeyCtx davKeyCtx = "password"
)

type FSService struct {
	inodes inodes.Service
}

func NewFSService(inodes inodes.Service) *FSService {
	return &FSService{inodes}
}

func (s *FSService) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	if name == "" {
		name = "/"
	}

	_, err := s.inodes.Mkdir(ctx, &inodes.PathCmd{
		Root:     root,
		UserID:   currentUser,
		FullName: name,
	})
	if err != nil {
		return fmt.Errorf("inodes mkdir error: %w", err)
	}

	return nil
}

func (s *FSService) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	if name == "" {
		name = "/"
	}

	res, err := s.inodes.Open(ctx, &inodes.PathCmd{
		Root:     root,
		UserID:   currentUser,
		FullName: name,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open inodes: %w", err)
	}

	if res == nil {
		return nil, fs.ErrNotExist
	}

	if res.Type() == inodes.Directory {
		return &Directory{res, s.inodes}, nil
	}

	return nil, webdav.ErrNotImplemented
}

func (s *FSService) RemoveAll(ctx context.Context, name string) error {
	fmt.Printf("Remove All: %q\n\n", name)
	return webdav.ErrNotImplemented
}

func (s *FSService) Rename(ctx context.Context, oldName, newName string) error {
	fmt.Printf("Rename %q -> %q: \n\n", oldName, newName)
	return webdav.ErrNotImplemented
}

func (s *FSService) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	if name == "" {
		name = "/"
	}

	res, err := s.inodes.Open(ctx, &inodes.PathCmd{
		Root:     root,
		UserID:   currentUser,
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
