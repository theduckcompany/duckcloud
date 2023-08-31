package folders

import (
	"context"
	"errors"
	"fmt"

	"github.com/theduckcompany/duckcloud/src/service/inodes"
	"github.com/theduckcompany/duckcloud/src/tools"
	"github.com/theduckcompany/duckcloud/src/tools/clock"
	"github.com/theduckcompany/duckcloud/src/tools/errs"
	"github.com/theduckcompany/duckcloud/src/tools/storage"
	"github.com/theduckcompany/duckcloud/src/tools/uuid"
)

var (
	ErrRootFSNotFound = errors.New("rootFS not found")
	ErrRootFSExist    = errors.New("rootFS exists")
	ErrInvalidRootFS  = errors.New("invalid rootFS")
)

//go:generate mockery --name Storage
type Storage interface {
	Save(ctx context.Context, folder *Folder) error
	GetByID(ctx context.Context, id uuid.UUID) (*Folder, error)
	GetAllUserFolders(ctx context.Context, userID uuid.UUID, cmd *storage.PaginateCmd) ([]Folder, error)
	Delete(ctx context.Context, folderID uuid.UUID) error
}

type FolderService struct {
	storage Storage
	inodes  inodes.Service
	clock   clock.Clock
	uuid    uuid.Service
}

func NewService(tools tools.Tools, storage Storage, inodes inodes.Service) *FolderService {
	return &FolderService{storage, inodes, tools.Clock(), tools.UUID()}
}

func (s *FolderService) CreatePersonalFolder(ctx context.Context, cmd *CreatePersonalFolderCmd) (*Folder, error) {
	err := cmd.Validate()
	if err != nil {
		return nil, errs.ValidationError(err)
	}

	// TODO: This action is not idempotent and could lead to orphan root dirs.
	inode, err := s.inodes.CreateRootDir(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to CreateRootDir: %w", err)
	}

	if !inode.IsDir() || inode.Parent() != nil {
		return nil, errs.BadRequest(ErrInvalidRootFS, "invalid rootFS")
	}

	folder := Folder{
		id:        s.uuid.New(),
		name:      cmd.Name,
		isPublic:  false,
		owners:    Owners{cmd.Owner},
		rootFS:    inode.ID(),
		createdAt: s.clock.Now(),
	}

	err = s.storage.Save(ctx, &folder)
	if err != nil {
		return nil, fmt.Errorf("failed to Save the folder: %w", err)
	}

	return &folder, nil
}

func (s *FolderService) Delete(ctx context.Context, folderID uuid.UUID) error {
	folder, err := s.storage.GetByID(ctx, folderID)
	if err != nil {
		return fmt.Errorf("failed to GetByID: %w", err)
	}

	if folder == nil {
		return nil
	}

	res, err := s.inodes.GetByID(ctx, folder.RootFS())
	if err != nil {
		return fmt.Errorf("failed to GetByID: %w", err)
	}

	if res != nil {
		return ErrRootFSExist
	}

	err = s.storage.Delete(ctx, folderID)
	if err != nil {
		return fmt.Errorf("failed to Delete: %w", err)
	}

	return nil
}

func (s *FolderService) GetByID(ctx context.Context, folderID uuid.UUID) (*Folder, error) {
	return s.storage.GetByID(ctx, folderID)
}

func (s *FolderService) GetAllUserFolders(ctx context.Context, userID uuid.UUID, cmd *storage.PaginateCmd) ([]Folder, error) {
	return s.storage.GetAllUserFolders(ctx, userID, cmd)
}
