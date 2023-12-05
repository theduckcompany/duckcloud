package dfs

import (
	"context"
	"errors"
	"fmt"
	"path"

	"github.com/theduckcompany/duckcloud/internal/service/files"
	"github.com/theduckcompany/duckcloud/internal/service/spaces"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

const DefaultSpaceName = "My files"

var (
	ErrNotImplemented  = errors.New("not implemented")
	ErrInvalidPath     = errors.New("invalid path")
	ErrInvalidRoot     = errors.New("invalid root")
	ErrInvalidParent   = errors.New("invalid parent")
	ErrInvalidMimeType = errors.New("invalid mime type")
	ErrIsNotDir        = errors.New("not a directory")
	ErrIsADir          = errors.New("is a directory")
	ErrNotFound        = errors.New("inode not found")
	ErrAlreadyExists   = errors.New("already exists")
)

type FSService struct {
	storage   Storage
	files     files.Service
	spaces    spaces.Service
	scheduler scheduler.Service
	tools     tools.Tools
}

func NewFSService(storage Storage, files files.Service, spaces spaces.Service, tasks scheduler.Service, tools tools.Tools) *FSService {
	return &FSService{storage, files, spaces, tasks, tools}
}

func (s *FSService) GetSpaceFS(space *spaces.Space) FS {
	return newLocalFS(s.storage, s.files, s.spaces, s.scheduler, s.tools)
}

func (s *FSService) RemoveFS(ctx context.Context, space *spaces.Space) error {
	fs := s.GetSpaceFS(space)

	err := fs.Remove(ctx, &PathCmd{Space: space, Path: "/"})
	if err != nil {
		return fmt.Errorf("failed to remove the fs: %w", err)
	}

	// XXX:MULTI-WRITE
	//
	err = s.spaces.Delete(ctx, space.ID())
	if err != nil {
		return fmt.Errorf("failed to delete the space %q: %w", space.ID(), err)
	}

	return nil
}

func (s *FSService) CreateFS(ctx context.Context, user *users.User, owners []uuid.UUID) (*spaces.Space, error) {
	// XXX:MULTI-WRITE
	space, err := s.spaces.Create(ctx, &spaces.CreateCmd{
		User:   user,
		Name:   DefaultSpaceName,
		Owners: owners,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create the space: %w", err)
	}

	now := s.tools.Clock().Now()
	node := INode{
		id:             s.tools.UUID().New(),
		parent:         nil,
		name:           "",
		spaceID:        space.ID(),
		createdAt:      now,
		createdBy:      user.ID(),
		lastModifiedAt: now,
		fileID:         nil,
	}

	err = s.storage.Save(ctx, &node)
	if err != nil {
		return nil, errs.Internal(fmt.Errorf("failed to Save: %w", err))
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
