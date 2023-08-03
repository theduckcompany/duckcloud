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
		return nil, fmt.Errorf("failed to count the nunber of inodes: %w", err)
	}

	if nb > 0 {
		return nil, ErrAlreadyBootstraped
	}

	node := INode{
		ID:             s.uuid.New(),
		UserID:         userID,
		Parent:         NoParent,
		Type:           Directory,
		CreatedAt:      s.clock.Now(),
		LastModifiedAt: s.clock.Now(),
	}

	err = s.storage.Save(ctx, &node)
	if err != nil {
		return nil, fmt.Errorf("failed to save the node into the storage: %w", err)
	}

	return &node, nil
}

func (s *INodeService) CreateDirectory(ctx context.Context, cmd *CreateDirectoryCmd) (*INode, error) {
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

	node := INode{
		ID:             s.uuid.New(),
		UserID:         cmd.UserID,
		Parent:         cmd.Parent,
		Type:           Directory,
		CreatedAt:      s.clock.Now(),
		LastModifiedAt: s.clock.Now(),
	}

	err = s.storage.Save(ctx, &node)
	if err != nil {
		return nil, fmt.Errorf("failed to save the node into the storage: %w", err)
	}

	return &node, nil
}
