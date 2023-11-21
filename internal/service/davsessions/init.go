package davsessions

import (
	"context"
	"database/sql"

	"github.com/theduckcompany/duckcloud/internal/service/spaces"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/secret"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

//go:generate mockery --name Service
type Service interface {
	GetAllForUser(ctx context.Context, userID uuid.UUID, paginateCmd *storage.PaginateCmd) ([]DavSession, error)
	Create(ctx context.Context, cmd *CreateCmd) (*DavSession, string, error)
	Authenticate(ctx context.Context, username string, password secret.Text) (*DavSession, error)
	Delete(ctx context.Context, cmd *DeleteCmd) error
	DeleteAll(ctx context.Context, userID uuid.UUID) error
}

func Init(db *sql.DB, spaces spaces.Service, tools tools.Tools) Service {
	storage := newSqlStorage(db)

	return NewService(storage, spaces, tools)
}
