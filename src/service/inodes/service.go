package inodes

import (
	"context"
	"errors"
	"fmt"

	"github.com/Peltoche/neurone/src/tools"
	"github.com/Peltoche/neurone/src/tools/clock"
	"github.com/Peltoche/neurone/src/tools/errs"
	"github.com/Peltoche/neurone/src/tools/uuid"
)

var (
	ErrInvalidParent      = errors.New("invalid parent")
	ErrAlreadyBootstraped = errors.New("this user is already bootstraped")
)

//go:generate mockery --name Storage
type Storage interface {
	Save(ctx context.Context, dir *INode) error
	GetByID(ctx context.Context, id uuid.UUID) (*INode, error)
	CountUserINodes(ctx context.Context, userID uuid.UUID) (uint, error)
}

type INodeService struct {
	storage Storage
	clock   clock.Clock
	uuid    uuid.Service
}

func NewService(tools tools.Tools, storage Storage) *INodeService {
	return &INodeService{storage, tools.Clock(), tools.UUID()}
}

func (s *INodeService) BootstrapUser(ctx context.Context, userID uuid.UUID) (*INode, error) {
	nb, err := s.storage.CountUserINodes(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to count the number of inodes: %w", err)
	}

	if nb > 0 {
		return nil, errs.BadRequest(ErrAlreadyBootstraped, "user alread bootstraped")
	}

	now := s.clock.Now()

	node := INode{
		ID:             s.uuid.New(),
		UserID:         userID,
		Parent:         NoParent,
		Type:           Directory,
		CreatedAt:      now,
		LastModifiedAt: now,
	}

	err = s.storage.Save(ctx, &node)
	if err != nil {
		return nil, fmt.Errorf("failed to save the node into the storage: %w", err)
	}

	return &node, nil
}

func (s *INodeService) GetByUserAndID(ctx context.Context, userID uuid.UUID, inodeID uuid.UUID) (*INode, error) {
	res, err := s.storage.GetByID(ctx, inodeID)
	if err != nil {
		return nil, err
	}

	if res.UserID != userID {
		return nil, errs.NotFound(fmt.Errorf("file %q is not owned by %q", inodeID, userID), "not found")
	}

	return res, nil
}

func (s *INodeService) CreateDirectory(ctx context.Context, cmd *CreateDirectoryCmd) (*INode, error) {
	err := cmd.Validate()
	if err != nil {
		return nil, errs.ValidationError(err)
	}

	parentNode, err := s.storage.GetByID(ctx, cmd.Parent)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch the parent node: %w", err)
	}

	if parentNode == nil {
		return nil, errs.BadRequest(fmt.Errorf("%w: parent doesn't exists", ErrInvalidParent), "invalid parent")
	}

	if parentNode.UserID != cmd.UserID {
		return nil, errs.BadRequest(fmt.Errorf("%w: parent not authorized", ErrInvalidParent), "invalid parent")
	}

	now := s.clock.Now()

	node := INode{
		ID:             s.uuid.New(),
		name:           cmd.Name,
		UserID:         cmd.UserID,
		Parent:         cmd.Parent,
		Type:           Directory,
		CreatedAt:      now,
		LastModifiedAt: now,
	}

	err = s.storage.Save(ctx, &node)
	if err != nil {
		return nil, fmt.Errorf("failed to save the node into the storage: %w", err)
	}

	return &node, nil
}
