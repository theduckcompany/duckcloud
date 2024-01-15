package spaces

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
	"github.com/theduckcompany/duckcloud/internal/service/users"
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
	GetAllSpaces(ctx context.Context, cmd *storage.PaginateCmd) ([]Space, error)
	Delete(ctx context.Context, spaceID uuid.UUID) error
	Patch(ctx context.Context, spaceID uuid.UUID, fields map[string]any) error
}

type SpaceService struct {
	storage   Storage
	clock     clock.Clock
	uuid      uuid.Service
	scheduler scheduler.Service
}

func NewService(tools tools.Tools, storage Storage, scheduler scheduler.Service) *SpaceService {
	return &SpaceService{storage, tools.Clock(), tools.UUID(), scheduler}
}

func (s *SpaceService) GetAllSpaces(ctx context.Context, user *users.User, cmd *storage.PaginateCmd) ([]Space, error) {
	if !user.IsAdmin() {
		return nil, errs.ErrUnauthorized
	}

	return s.storage.GetAllSpaces(ctx, cmd)
}

func (s *SpaceService) Create(ctx context.Context, cmd *CreateCmd) (*Space, error) {
	err := cmd.Validate()
	if err != nil {
		return nil, errs.Validation(err)
	}

	if !cmd.User.IsAdmin() {
		return nil, errs.ErrUnauthorized
	}

	// Ensure that the owners are set only once.
	uniqOwnersMap := make(map[uuid.UUID]struct{})
	for _, owner := range cmd.Owners {
		uniqOwnersMap[owner] = struct{}{}
	}

	uniqOwners := []uuid.UUID{}
	for owner := range uniqOwnersMap {
		uniqOwners = append(uniqOwners, owner)
	}

	now := s.clock.Now()
	space := Space{
		id:        s.uuid.New(),
		name:      cmd.Name,
		owners:    uniqOwners,
		createdAt: now,
		createdBy: cmd.User.ID(),
	}

	err = s.storage.Save(context.WithoutCancel(ctx), &space)
	if err != nil {
		return nil, errs.Internal(fmt.Errorf("failed to Save the space: %w", err))
	}

	return &space, nil
}

func (s *SpaceService) Delete(ctx context.Context, user *users.User, spaceID uuid.UUID) error {
	if !user.IsAdmin() {
		return errs.Unauthorized(fmt.Errorf("%q is not an admin", user.Username()))
	}

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

func (s *SpaceService) AddOwner(ctx context.Context, cmd *AddOwnerCmd) (*Space, error) {
	if !cmd.User.IsAdmin() {
		return nil, errs.ErrUnauthorized
	}

	space, err := s.storage.GetByID(ctx, cmd.SpaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get the space: %w", err)
	}

	currentOwners := space.Owners()
	if slices.Contains[[]uuid.UUID, uuid.UUID](currentOwners, cmd.Owner.ID()) {
		return space, nil
	}

	space.owners = append(currentOwners, cmd.Owner.ID())

	err = s.storage.Patch(ctx, space.ID(), map[string]any{"owners": space.owners})
	if err != nil {
		return nil, fmt.Errorf("failed to patch the space's owners field: %w", err)
	}

	return space, nil
}

func (s *SpaceService) RemoveOwner(ctx context.Context, cmd *RemoveOwnerCmd) (*Space, error) {
	// Anyone can remove itself from a space but only the admin can remove an another user.
	if !cmd.User.IsAdmin() && cmd.User != cmd.Owner {
		return nil, errs.ErrUnauthorized
	}

	space, err := s.storage.GetByID(ctx, cmd.SpaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get the space: %w", err)
	}

	if !slices.Contains(space.Owners(), cmd.Owner.ID()) {
		return space, nil
	}

	space.owners = slices.DeleteFunc(space.Owners(), func(id uuid.UUID) bool {
		return id == cmd.Owner.ID()
	})

	err = s.storage.Patch(ctx, space.ID(), map[string]any{"owners": space.owners})
	if err != nil {
		return nil, fmt.Errorf("failed to patch the space's owners field: %w", err)
	}

	return space, nil
}

func (s *SpaceService) Bootstrap(ctx context.Context, user *users.User) error {
	res, err := s.storage.GetAllSpaces(ctx, &storage.PaginateCmd{Limit: 1})
	if err != nil {
		return fmt.Errorf("faile to get all the spaces: %w", err)
	}

	if len(res) > 0 {
		return nil
	}

	err = s.scheduler.RegisterSpaceCreateTask(ctx, &scheduler.SpaceCreateArgs{
		UserID: user.ID(),
		Name:   BootstrapSpaceName,
		Owners: []uuid.UUID{user.ID()},
	})
	if err != nil {
		return fmt.Errorf("failed to RegisterSpaceCreateTask: %w", err)
	}

	return nil
}
