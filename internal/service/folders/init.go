package folders

import (
	"context"
	"database/sql"

	"github.com/theduckcompany/duckcloud/internal/service/inodes"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

//go:generate mockery --name Service
type Service interface {
	CreatePersonalFolder(ctx context.Context, cmd *CreatePersonalFolderCmd) (*Folder, error)
	GetAllUserFolders(ctx context.Context, userID uuid.UUID, cmd *storage.PaginateCmd) ([]Folder, error)
	GetUserFolder(ctx context.Context, userID, folderID uuid.UUID) (*Folder, error)
	GetByID(ctx context.Context, folderID uuid.UUID) (*Folder, error)
	Delete(ctx context.Context, folderID uuid.UUID) error
	RegisterWrite(ctx context.Context, folderID uuid.UUID, size uint64) (*Folder, error)
	RegisterDeletion(ctx context.Context, folderID uuid.UUID, size uint64) (*Folder, error)
	GetAllFoldersWithRoot(ctx context.Context, rootID uuid.UUID, cmd *storage.PaginateCmd) ([]Folder, error)
}

func Init(tools tools.Tools, db *sql.DB, inodes inodes.Service) Service {
	storage := newSqlStorage(db, tools)

	return NewService(tools, storage, inodes)
}
