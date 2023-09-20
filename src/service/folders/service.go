package folders

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"

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
	ErrNotFound       = errors.New("folder not found")
)

//go:generate mockery --name Storage
type Storage interface {
	Save(ctx context.Context, folder *Folder) error
	GetByID(ctx context.Context, id uuid.UUID) (*Folder, error)
	GetAllUserFolders(ctx context.Context, userID uuid.UUID, cmd *storage.PaginateCmd) ([]Folder, error)
	Delete(ctx context.Context, folderID uuid.UUID) error
	Patch(ctx context.Context, folderID uuid.UUID, fields map[string]any) error
	GetAllFoldersWithRoot(ctx context.Context, rootID uuid.UUID, cmd *storage.PaginateCmd) ([]Folder, error)
}

type FolderService struct {
	storage Storage
	inodes  inodes.Service
	clock   clock.Clock
	uuid    uuid.Service
	lockMap sync.Map
}

func NewService(tools tools.Tools, storage Storage, inodes inodes.Service) *FolderService {
	return &FolderService{storage, inodes, tools.Clock(), tools.UUID(), sync.Map{}}
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

	now := s.clock.Now()
	folder := Folder{
		id:             s.uuid.New(),
		name:           cmd.Name,
		isPublic:       false,
		owners:         Owners{cmd.Owner},
		rootFS:         inode.ID(),
		size:           0,
		createdAt:      now,
		lastModifiedAt: now,
	}

	err = s.storage.Save(ctx, &folder)
	if err != nil {
		return nil, fmt.Errorf("failed to Save the folder: %w", err)
	}

	return &folder, nil
}

func (s *FolderService) RegisterWrite(ctx context.Context, folderID uuid.UUID, size uint64) (*Folder, error) {
	val, _ := s.lockMap.LoadOrStore(folderID, new(sync.Mutex))
	lock := val.(*sync.Mutex)

	lock.Lock()
	defer lock.Unlock()

	return s.registerSizeChange(ctx, folderID, size, true)
}

func (s *FolderService) RegisterDeletion(ctx context.Context, folderID uuid.UUID, size uint64) (*Folder, error) {
	return s.registerSizeChange(ctx, folderID, size, false)
}

func (s *FolderService) registerSizeChange(ctx context.Context, folderID uuid.UUID, size uint64, positif bool) (*Folder, error) {
	folder, err := s.GetByID(ctx, folderID)
	if err != nil {
		return nil, fmt.Errorf("failed to GetByID: %w", err)
	}

	if folder == nil {
		return nil, ErrNotFound
	}

	folder.lastModifiedAt = s.clock.Now()
	switch {
	case positif:
		folder.size += size
	case !positif && size < folder.size:
		folder.size -= size
	default:
		// We try to remove more bytes than there is.
		folder.size = 0
	}

	err = s.storage.Patch(ctx, folderID, map[string]any{
		"last_modified_at": folder.lastModifiedAt,
		"size":             folder.size,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to Patch: %w", err)
	}

	return folder, nil
}

func (s *FolderService) Delete(ctx context.Context, folderID uuid.UUID) error {
	val, _ := s.lockMap.LoadOrStore(folderID, new(sync.Mutex))
	lock := val.(*sync.Mutex)

	lock.Lock()
	defer lock.Unlock()

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

func (s *FolderService) GetAllFoldersWithRoot(ctx context.Context, rootID uuid.UUID, cmd *storage.PaginateCmd) ([]Folder, error) {
	return s.storage.GetAllFoldersWithRoot(ctx, rootID, cmd)
}

func (s *FolderService) GetByID(ctx context.Context, folderID uuid.UUID) (*Folder, error) {
	return s.storage.GetByID(ctx, folderID)
}

func (s *FolderService) GetAllUserFolders(ctx context.Context, userID uuid.UUID, cmd *storage.PaginateCmd) ([]Folder, error) {
	return s.storage.GetAllUserFolders(ctx, userID, cmd)
}

func (s *FolderService) GetUserFolder(ctx context.Context, userID, folderID uuid.UUID) (*Folder, error) {
	folder, err := s.storage.GetByID(ctx, folderID)
	if err != nil {
		return nil, err
	}

	if !slices.Contains[[]uuid.UUID, uuid.UUID](folder.Owners(), userID) {
		return nil, nil
	}

	return folder, nil
}
