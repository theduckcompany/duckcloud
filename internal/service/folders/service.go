package folders

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"

	"github.com/theduckcompany/duckcloud/internal/service/inodes"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/clock"
	"github.com/theduckcompany/duckcloud/internal/tools/errs"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

var (
	ErrRootFSNotFound      = errors.New("rootFS not found")
	ErrRootFSExist         = errors.New("rootFS exists")
	ErrInvalidRootFS       = errors.New("invalid rootFS")
	ErrNotFound            = errors.New("folder not found")
	ErrInvalidFolderAccess = errors.New("no access to folder")
)

//go:generate mockery --name Storage
type Storage interface {
	Save(ctx context.Context, folder *Folder) error
	GetByID(ctx context.Context, id uuid.UUID) (*Folder, error)
	GetAllUserFolders(ctx context.Context, userID uuid.UUID, cmd *storage.PaginateCmd) ([]Folder, error)
	Delete(ctx context.Context, folderID uuid.UUID) error
	Patch(ctx context.Context, folderID uuid.UUID, fields map[string]any) error
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
		return nil, errs.Validation(err)
	}

	inode, err := s.inodes.CreateRootDir(ctx)
	if err != nil {
		return nil, errs.Internal(fmt.Errorf("failed to CreateRootDir: %w", err))
	}

	now := s.clock.Now()
	folder := Folder{
		id:        s.uuid.New(),
		name:      cmd.Name,
		isPublic:  false,
		owners:    Owners{cmd.Owner},
		rootFS:    inode.ID(),
		createdAt: now,
	}

	// XXX:MULTI-WRITE
	// This method is used inside the task "usercreate". It means it should be
	// idempotent because in case of failure the job should be retriggered.
	//
	// This function is not idempotent and could lead to the creation of orphan
	// inodes. That's not great but that's not terrible.
	err = s.storage.Save(context.WithoutCancel(ctx), &folder)
	if err != nil {
		return nil, errs.Internal(fmt.Errorf("failed to Save the folder: %w", err))
	}

	return &folder, nil
}

func (s *FolderService) Delete(ctx context.Context, folderID uuid.UUID) error {
	val, _ := s.lockMap.LoadOrStore(folderID, new(sync.Mutex))
	lock := val.(*sync.Mutex)

	lock.Lock()
	defer lock.Unlock()

	folder, err := s.storage.GetByID(ctx, folderID)
	if errors.Is(err, errNotFound) {
		return nil
	}

	if err != nil {
		return errs.Internal(fmt.Errorf("failed to GetByID: %w", err))
	}

	res, err := s.inodes.GetByID(ctx, folder.RootFS())
	if err != nil && !errors.Is(err, errs.ErrNotFound) {
		return errs.Internal(fmt.Errorf("failed to GetByID: %w", err))
	}

	if res != nil {
		return ErrRootFSExist
	}

	err = s.storage.Delete(ctx, folderID)
	if err != nil {
		return errs.Internal(fmt.Errorf("failed to Delete: %w", err))
	}

	return nil
}

func (s *FolderService) GetByID(ctx context.Context, folderID uuid.UUID) (*Folder, error) {
	res, err := s.storage.GetByID(ctx, folderID)
	if errors.Is(err, errNotFound) {
		return nil, errs.NotFound(err)
	}
	if err != nil {
		return nil, errs.Internal(err)
	}

	return res, nil
}

func (s *FolderService) GetAllUserFolders(ctx context.Context, userID uuid.UUID, cmd *storage.PaginateCmd) ([]Folder, error) {
	res, err := s.storage.GetAllUserFolders(ctx, userID, cmd)
	if err != nil {
		return nil, errs.Internal(err)
	}

	return res, nil
}

func (s *FolderService) GetUserFolder(ctx context.Context, userID, folderID uuid.UUID) (*Folder, error) {
	folder, err := s.storage.GetByID(ctx, folderID)
	if errors.Is(err, errNotFound) {
		return nil, errs.NotFound(err)
	}

	if err != nil {
		return nil, errs.Internal(fmt.Errorf("failed to GetByID: %w", err))
	}

	if !slices.Contains[[]uuid.UUID, uuid.UUID](folder.Owners(), userID) {
		return nil, errs.Unauthorized(ErrInvalidFolderAccess)
	}

	return folder, nil
}
