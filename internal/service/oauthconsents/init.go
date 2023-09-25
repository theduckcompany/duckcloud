package oauthconsents

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/theduckcompany/duckcloud/internal/service/oauthclients"
	"github.com/theduckcompany/duckcloud/internal/service/websessions"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

//go:generate mockery --name Service
type Service interface {
	Create(ctx context.Context, cmd *CreateCmd) (*Consent, error)
	Check(r *http.Request, client *oauthclients.Client, session *websessions.Session) error
	Delete(ctx context.Context, consentID uuid.UUID) error
	GetAll(ctx context.Context, userID uuid.UUID, cmd *storage.PaginateCmd) ([]Consent, error)
	DeleteAll(ctx context.Context, userID uuid.UUID) error
}

func Init(tools tools.Tools, db *sql.DB) Service {
	storage := newSQLStorage(db)

	return NewService(storage, tools)
}
