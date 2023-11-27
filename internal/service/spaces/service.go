package spaces

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/clock"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

var (
	ErrRootFSNotFound     = errors.New("rootFS not found")
	ErrRootFSExist        = errors.New("rootFS exists")
	ErrInvalidRootFS      = errors.New("invalid rootFS")
	ErrNotFound           = errors.New("space not found")
	ErrInvalidSpaceAccess = errors.New("no access to space")
)

//go:generate mockery --name Storage
type Storage interface {
	Save(ctx context.Context, space *Space) error
	GetByID(ctx context.Context, id uuid.UUID) (*Space, error)
	GetAllUserSpaces(ctx context.Context, userID uuid.UUID, cmd *storage.PaginateCmd) ([]Space, error)
	Delete(ctx context.Context, spaceID uuid.UUID) error
	Patch(ctx context.Context, spaceID uuid.UUID, fields map[string]any) error
}

type SpaceService struct {
	storage Storage
	clock   clock.Clock
	uuid    uuid.Service
}

func NewService(tools tools.Tools, storage Storage) *SpaceService {
	return &SpaceService{storage, tools.Clock(), tools.UUID()}
}

func (s *SpaceService) Create(ctx context.Context, cmd *CreateCmd) (*Space, error) {
	err := cmd.Validate()
	if err != nil {
		return nil, errs.Validation(err)
	}

	now := s.clock.Now()
	space := Space{
		id:        s.uuid.New(),
		name:      cmd.Name,
		isPublic:  len(cmd.Owners) > 1,
		owners:    cmd.Owners,
		createdAt: now,
		createdBy: cmd.User.ID(),
	}

	err = s.storage.Save(context.WithoutCancel(ctx), &space)
	if err != nil {
		return nil, errs.Internal(fmt.Errorf("failed to Save the space: %w", err))
	}

	return &space, nil
}

func (s *SpaceService) Delete(ctx context.Context, spaceID uuid.UUID) error {
	err := s.storage.Delete(ctx, spaceID)
	if err != nil {
		return errs.Internal(fmt.Errorf("failed to Delete: %w", err))
	}

	return nil
}

func (s *SpaceService) GetByID(ctx context.Context, spaceID uuid.UUID) (*Space, error) {
	res, err := s.storage.GetByID(ctx, spaceID)
	if errors.Is(err, errNotFound) {
		return nil, errs.NotFound(err)
	}
	if err != nil {
		return nil, errs.Internal(err)
	}

	return res, nil
}

func (s *SpaceService) GetAllUserSpaces(ctx context.Context, userID uuid.UUID, cmd *storage.PaginateCmd) ([]Space, error) {
	res, err := s.storage.GetAllUserSpaces(ctx, userID, cmd)
	if err != nil {
		return nil, errs.Internal(err)
	}

	return res, nil
}

func (s *SpaceService) GetUserSpace(ctx context.Context, userID, spaceID uuid.UUID) (*Space, error) {
	space, err := s.storage.GetByID(ctx, spaceID)
	if errors.Is(err, errNotFound) {
		return nil, errs.NotFound(err)
	}

	if err != nil {
		return nil, errs.Internal(fmt.Errorf("failed to GetByID: %w", err))
	}

	if !slices.Contains[[]uuid.UUID, uuid.UUID](space.Owners(), userID) {
		return nil, errs.Unauthorized(ErrInvalidSpaceAccess)
	}

	return space, nil
}
