package websessions

import (
	"context"
	"database/sql"
	"errors"
	"net/http"

	"github.com/theduckcompany/duckcloud/src/tools"
	"github.com/theduckcompany/duckcloud/src/tools/storage"
	"github.com/theduckcompany/duckcloud/src/tools/uuid"
)

var (
	ErrMissingSessionToken = errors.New("missing session token")
	ErrSessionNotFound     = errors.New("session not found")
)

//go:generate mockery --name Service
type Service interface {
	Create(ctx context.Context, cmd *CreateCmd) (*Session, error)
	GetByToken(ctx context.Context, token string) (*Session, error)
	GetFromReq(r *http.Request) (*Session, error)
	Logout(r *http.Request, w http.ResponseWriter) error
	GetAllForUser(ctx context.Context, userID uuid.UUID, cmd *storage.PaginateCmd) ([]Session, error)
	Revoke(ctx context.Context, cmd *RevokeCmd) error
	RevokeAll(ctx context.Context, userID uuid.UUID) error
}

func Init(tools tools.Tools, db *sql.DB) Service {
	storage := newSQLStorage(db)

	return NewService(storage, tools)
}
