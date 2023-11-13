package users

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/theduckcompany/duckcloud/internal/service/davsessions"
	"github.com/theduckcompany/duckcloud/internal/service/dfs"
	"github.com/theduckcompany/duckcloud/internal/service/dfs/folders"
	"github.com/theduckcompany/duckcloud/internal/service/oauthconsents"
	"github.com/theduckcompany/duckcloud/internal/service/oauthsessions"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/runner"
	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
	"github.com/theduckcompany/duckcloud/internal/service/websessions"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/secret"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
	"go.uber.org/fx"
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
	SetDefaultFolder(ctx context.Context, user User, folder *folders.Folder) (*User, error)
}

type Result struct {
	fx.Out
	Service        Service
	UserDeleteTask runner.TaskRunner `group:"tasks"`
	UserCreateTask runner.TaskRunner `group:"tasks"`
}

func Init(
	tools tools.Tools,
	db *sql.DB,
	scheduler scheduler.Service,
	folders folders.Service,
	fs dfs.Service,
	webSessions websessions.Service,
	davSessions davsessions.Service,
	oauthSessions oauthsessions.Service,
	oauthConsents oauthconsents.Service,
) (Result, error) {
	store := newSqlStorage(db, tools)

	svc := NewService(tools, store, scheduler)

	res, err := svc.GetAll(context.Background(), &storage.PaginateCmd{Limit: 4})
	if err != nil {
		return Result{}, fmt.Errorf("failed to GetAll users: %w", err)
	}

	if len(res) == 0 {
		_, err = svc.Create(context.Background(), &CreateCmd{
			Username: BoostrapUsername,
			Password: secret.NewText(BoostrapPassword),
			IsAdmin:  true,
		})
		if err != nil {
			return Result{}, fmt.Errorf("failed to create the first user: %w", err)
		}
	}

	return Result{
		Service:        svc,
		UserCreateTask: NewUserCreateTaskRunner(svc, folders, fs),
		UserDeleteTask: NewUserDeleteTaskRunner(svc, webSessions, davSessions, oauthSessions, oauthConsents, folders, fs),
	}, nil
}
