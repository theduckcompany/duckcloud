package users

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
	Create(ctx context.Context, user *CreateCmd) (*User, error)
	GetByID(ctx context.Context, userID uuid.UUID) (*User, error)
	Authenticate(ctx context.Context, username, password string) (*User, error)
	GetAll(ctx context.Context, paginateCmd *storage.PaginateCmd) ([]User, error)
	Delete(ctx context.Context, userID uuid.UUID) error
	GetDeleted(ctx context.Context, limit int) ([]User, error)
	HardDelete(ctx context.Context, userID uuid.UUID) error
}

func Init(tools tools.Tools, db *sql.DB, inodes inodes.Service) Service {
	storage := newSqlStorage(db, tools)

	return NewService(tools, storage, inodes)
}
