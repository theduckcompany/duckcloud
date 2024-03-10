package websessions

import (
	"context"
	"database/sql"
	"errors"
	"net/http"

	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/secret"
	"github.com/theduckcompany/duckcloud/internal/tools/sqlstorage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

var (
	ErrMissingSessionToken = errors.New("missing session token")
	ErrSessionNotFound     = errors.New("session not found")
)

//go:generate mockery --name Service
type Service interface {
	Create(ctx context.Context, cmd *CreateCmd) (*Session, error)
	GetByToken(ctx context.Context, token secret.Text) (*Session, error)
	GetFromReq(r *http.Request) (*Session, error)
	Logout(r *http.Request, w http.ResponseWriter) error
	GetAllForUser(ctx context.Context, userID uuid.UUID, cmd *sqlstorage.PaginateCmd) ([]Session, error)
	Delete(ctx context.Context, cmd *DeleteCmd) error
	DeleteAll(ctx context.Context, userID uuid.UUID) error
}

func Init(tools tools.Tools, db *sql.DB) Service {
	storage := newSQLStorage(db)

	return newService(storage, tools)
}
