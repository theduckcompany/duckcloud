package spaces

import (
	"context"
	"database/sql"

	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/sqlstorage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

//go:generate mockery --name Service
type Service interface {
	Bootstrap(ctx context.Context, user *users.User) error
	Create(ctx context.Context, cmd *CreateCmd) (*Space, error)
	GetAllUserSpaces(ctx context.Context, userID uuid.UUID, cmd *sqlstorage.PaginateCmd) ([]Space, error)
	GetAllSpaces(ctx context.Context, user *users.User, cmd *sqlstorage.PaginateCmd) ([]Space, error)
	GetUserSpace(ctx context.Context, userID, spaceID uuid.UUID) (*Space, error)
	GetByID(ctx context.Context, spaceID uuid.UUID) (*Space, error)
	AddOwner(ctx context.Context, cmd *AddOwnerCmd) (*Space, error)
	RemoveOwner(ctx context.Context, cmd *RemoveOwnerCmd) (*Space, error)
	Delete(ctx context.Context, user *users.User, spaceID uuid.UUID) error
}

func Init(tools tools.Tools, db *sql.DB, scheduler scheduler.Service) Service {
	storage := newSqlStorage(db, tools)

	return newService(tools, storage, scheduler)
}
