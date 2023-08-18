package websessions

import (
	"context"
	"database/sql"
	"errors"
	"net/http"

	"github.com/myminicloud/myminicloud/src/tools"
	"github.com/myminicloud/myminicloud/src/tools/uuid"
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
	GetUserSessions(ctx context.Context, userID uuid.UUID) ([]Session, error)
}

func Init(tools tools.Tools, db *sql.DB) Service {
	storage := newSQLStorage(db)

	return NewService(storage, tools)
}
