package oauthsessions

import (
	"context"
	"database/sql"

	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/secret"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

//go:generate mockery --name Service
type Service interface {
	Create(ctx context.Context, input *CreateCmd) (*Session, error)
	RemoveByAccessToken(ctx context.Context, access secret.Text) error
	RemoveByRefreshToken(ctx context.Context, refresh secret.Text) error
	GetByAccessToken(ctx context.Context, access secret.Text) (*Session, error)
	GetByRefreshToken(ctx context.Context, refresh secret.Text) (*Session, error)
	DeleteAllForUser(ctx context.Context, userID uuid.UUID) error
}

func Init(tools tools.Tools, db *sql.DB) Service {
	storage := newSqlStorage(db)

	return newService(tools, storage)
}
