package dfs

import (
	"context"
	"errors"
	"fmt"
	"path"

	"github.com/theduckcompany/duckcloud/internal/service/dfs/internal/inodes"
	"github.com/theduckcompany/duckcloud/internal/service/files"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

var (
	ErrNotImplemented = errors.New("not implemented")
	ErrInvalidPath    = inodes.ErrInvalidPath
	ErrAlreadyExists  = inodes.ErrAlreadyExists
)

type FSService struct {
	inodes    inodes.Service
	files     files.Service
	scheduler scheduler.Service
	tools     tools.Tools
}

func NewFSService(inodes inodes.Service, files files.Service, tasks scheduler.Service, tools tools.Tools) *FSService {
	return &FSService{inodes, files, tasks, tools}
}

func (s *FSService) GetSpaceFS(spaceID uuid.UUID) FS {
	return newLocalFS(s.inodes, s.files, spaceID, s.scheduler, s.tools)
}

func (s *FSService) RemoveSpaceFS(ctx context.Context, spaceID uuid.UUID) error {
	root, err := s.inodes.GetSpaceRoot(ctx, spaceID)
	if errors.Is(err, errs.ErrNotFound) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to GetSpaceRoot: %w", err)
	}

	err = s.inodes.Remove(ctx, root)
	if err != nil {
		return fmt.Errorf("failed to remove the rootFS: %w", err)
	}

	return nil
}

func (s *FSService) CreateSpaceFS(ctx context.Context, spaceID uuid.UUID) (*INode, error) {
	root, err := s.inodes.CreateSpaceRootDir(ctx, spaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to CreateRootDir: %w", err)
	}

	return root, nil
}

// CleanPath is equivalent to but slightly more efficient than
// path.Clean("/" + name).
func CleanPath(name string) string {
	if name == "" || name[0] != '/' {
		name = "/" + name
	}
	return path.Clean(name)
}
