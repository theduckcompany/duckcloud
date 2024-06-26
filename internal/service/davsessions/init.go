package davsessions

import (
	"context"

	"github.com/theduckcompany/duckcloud/internal/service/spaces"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/secret"
	"github.com/theduckcompany/duckcloud/internal/tools/sqlstorage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

//go:generate mockery --name Service
type Service interface {
	GetAllForUser(ctx context.Context, userID uuid.UUID, paginateCmd *sqlstorage.PaginateCmd) ([]DavSession, error)
	Create(ctx context.Context, cmd *CreateCmd) (*DavSession, string, error)
	Authenticate(ctx context.Context, username string, password secret.Text) (*DavSession, error)
	Delete(ctx context.Context, cmd *DeleteCmd) error
	DeleteAll(ctx context.Context, userID uuid.UUID) error
}

func Init(db sqlstorage.Querier, spaces spaces.Service, tools tools.Tools) Service {
	storage := newSqlStorage(db)

	return newService(storage, spaces, tools)
}
