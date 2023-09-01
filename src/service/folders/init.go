package folders

import (
	"context"
	"database/sql"

	"github.com/theduckcompany/duckcloud/src/service/inodes"
	"github.com/theduckcompany/duckcloud/src/tools"
	"github.com/theduckcompany/duckcloud/src/tools/storage"
	"github.com/theduckcompany/duckcloud/src/tools/uuid"
)

//go:generate mockery --name Service
type Service interface {
	CreatePersonalFolder(ctx context.Context, cmd *CreatePersonalFolderCmd) (*Folder, error)
	GetAllUserFolders(ctx context.Context, userID uuid.UUID, cmd *storage.PaginateCmd) ([]Folder, error)
	GetByID(ctx context.Context, folderID uuid.UUID) (*Folder, error)
	Delete(ctx context.Context, folderID uuid.UUID) error
}

func Init(tools tools.Tools, db *sql.DB, inodes inodes.Service) Service {
	storage := newSqlStorage(db, tools)

	return NewService(tools, storage, inodes)
}