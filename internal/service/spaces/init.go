package spaces

import (
	"context"
	"database/sql"

	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

//go:generate mockery --name Service
type Service interface {
	Create(ctx context.Context, cmd *CreateCmd) (*Space, error)
	GetAllUserSpaces(ctx context.Context, userID uuid.UUID, cmd *storage.PaginateCmd) ([]Space, error)
	GetUserSpace(ctx context.Context, userID, spaceID uuid.UUID) (*Space, error)
	GetByID(ctx context.Context, spaceID uuid.UUID) (*Space, error)
	Delete(ctx context.Context, spaceID uuid.UUID) error
}

func Init(tools tools.Tools, db *sql.DB) Service {
	storage := newSqlStorage(db, tools)

	return NewService(tools, storage)
}
