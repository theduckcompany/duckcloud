package users

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/secret"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

const (
	BoostrapUsername = "admin"
	BoostrapPassword = "duckcloud"
)

//go:generate mockery --name Service
type Service interface {
	Create(ctx context.Context, user *CreateCmd) (*User, error)
	GetByID(ctx context.Context, userID uuid.UUID) (*User, error)
	Authenticate(ctx context.Context, username string, password secret.Text) (*User, error)
	GetAll(ctx context.Context, paginateCmd *storage.PaginateCmd) ([]User, error)
	AddToDeletion(ctx context.Context, userID uuid.UUID) error
	HardDelete(ctx context.Context, userID uuid.UUID) error
	GetAllWithStatus(ctx context.Context, status Status, cmd *storage.PaginateCmd) ([]User, error)
	MarkInitAsFinished(ctx context.Context, userID uuid.UUID) (*User, error)
	UpdateUserPassword(ctx context.Context, cmd *UpdatePasswordCmd) error
}

func Init(
	ctx context.Context,
	tools tools.Tools,
	db *sql.DB,
	scheduler scheduler.Service,
) (Service, error) {
	store := newSqlStorage(db, tools)

	svc := NewService(tools, store, scheduler)

	res, err := svc.GetAll(ctx, &storage.PaginateCmd{Limit: 4})
	if err != nil {
		return nil, fmt.Errorf("failed to GetAll users: %w", err)
	}

	if len(res) == 0 {
		_, err = svc.bootstrap(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to create the first user: %w", err)
		}
	}

	return svc, nil
}
