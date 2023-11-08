package folders

import (
	"context"
	"database/sql"

	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

//go:generate mockery --name Service
type Service interface {
	Create(ctx context.Context, cmd *CreateCmd) (*Folder, error)
	GetAllUserFolders(ctx context.Context, userID uuid.UUID, cmd *storage.PaginateCmd) ([]Folder, error)
	GetUserFolder(ctx context.Context, userID, folderID uuid.UUID) (*Folder, error)
	GetByID(ctx context.Context, folderID uuid.UUID) (*Folder, error)
	Delete(ctx context.Context, folderID uuid.UUID) error
}

func Init(tools tools.Tools, db *sql.DB) Service {
	storage := newSqlStorage(db, tools)

	return NewService(tools, storage)
}
