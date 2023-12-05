package dfs

import (
	"context"
	"errors"
	"fmt"

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
