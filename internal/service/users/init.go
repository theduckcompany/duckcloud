package users

import (
	"context"
	"database/sql"

	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/secret"
	"github.com/theduckcompany/duckcloud/internal/tools/sqlstorage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

const (
	BoostrapUsername = "admin"
	BoostrapPassword = "duckcloud"
)

//go:generate mockery --name Service
type Service interface {
	Create(ctx context.Context, user *CreateCmd) (*User, error)
	Bootstrap(ctx context.Context) (*User, error)
	GetByID(ctx context.Context, userID uuid.UUID) (*User, error)
	Authenticate(ctx context.Context, username string, password secret.Text) (*User, error)
	GetAll(ctx context.Context, paginateCmd *sqlstorage.PaginateCmd) ([]User, error)
	AddToDeletion(ctx context.Context, userID uuid.UUID) error
	HardDelete(ctx context.Context, userID uuid.UUID) error
	GetAllWithStatus(ctx context.Context, status Status, cmd *sqlstorage.PaginateCmd) ([]User, error)
	MarkInitAsFinished(ctx context.Context, userID uuid.UUID) (*User, error)
	UpdateUserPassword(ctx context.Context, cmd *UpdatePasswordCmd) error
}

func Init(
	tools tools.Tools,
	db *sql.DB,
	scheduler scheduler.Service,
) Service {
	store := newSqlStorage(db, tools)

	return newService(tools, store, scheduler)
}
