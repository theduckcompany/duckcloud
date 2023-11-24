package dfs

import (
	"context"
	"errors"
	"fmt"
	"path"

	"github.com/theduckcompany/duckcloud/internal/service/dfs/internal/inodes"
	"github.com/theduckcompany/duckcloud/internal/service/files"
	"github.com/theduckcompany/duckcloud/internal/service/spaces"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

const DefaultSpaceName = "My files"

var (
	ErrNotImplemented = errors.New("not implemented")
	ErrInvalidPath    = inodes.ErrInvalidPath
	ErrAlreadyExists  = inodes.ErrAlreadyExists
)

type FSService struct {
	inodes    inodes.Service
	files     files.Service
	spaces    spaces.Service
	scheduler scheduler.Service
	tools     tools.Tools
}

func NewFSService(inodes inodes.Service, files files.Service, spaces spaces.Service, tasks scheduler.Service, tools tools.Tools) *FSService {
	return &FSService{inodes, files, spaces, tasks, tools}
}

func (s *FSService) GetSpaceFS(space *spaces.Space) FS {
	return newLocalFS(s.inodes, s.files, space, s.spaces, s.scheduler, s.tools)
}

func (s *FSService) RemoveFS(ctx context.Context, space *spaces.Space) error {
	rootFS, err := s.inodes.GetByID(ctx, space.RootFS())
	if err != nil && !errors.Is(err, errs.ErrNotFound) {
		return fmt.Errorf("failed to Get the rootFS for %q: %w", space.Name(), err)
	}

	// XXX:MULTI-WRITE
	//
	// TODO: Create a spacedelete task
	if rootFS != nil {
		err = s.inodes.Remove(ctx, rootFS)
		if err != nil {
			return fmt.Errorf("failed to remove the rootFS for %q: %w", space.Name(), err)
		}
	}
	err = s.spaces.Delete(ctx, space.ID())
	if err != nil {
		return fmt.Errorf("failed to delete the space %q: %w", space.ID(), err)
	}

	return nil
}

func (s *FSService) CreateFS(ctx context.Context, owners []uuid.UUID) (*spaces.Space, error) {
	root, err := s.inodes.CreateRootDir(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to CreateRootDir: %w", err)
	}

	// XXX:MULTI-WRITE
	space, err := s.spaces.Create(ctx, &spaces.CreateCmd{
		Name:   DefaultSpaceName,
		Owners: owners,
		RootFS: root.ID(),
	})
	if err != nil {
		_ = s.inodes.Remove(ctx, root)

		return nil, fmt.Errorf("failed to create the space: %w", err)
	}

	return space, nil
}

// CleanPath is equivalent to but slightly more efficient than
// path.Clean("/" + name).
func CleanPath(name string) string {
	if name == "" || name[0] != '/' {
		name = "/" + name
	}
	return path.Clean(name)
}
