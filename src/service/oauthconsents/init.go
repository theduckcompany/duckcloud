package oauthconsents

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/Peltoche/neurone/src/service/oauthclients"
	"github.com/Peltoche/neurone/src/service/websessions"
	"github.com/Peltoche/neurone/src/tools"
)

//go:generate mockery --name Service
type Service interface {
	Create(ctx context.Context, cmd *CreateCmd) (*Consent, error)
	Check(r *http.Request, client *oauthclients.Client, session *websessions.Session) error
}

func Init(tools tools.Tools, db *sql.DB) Service {
	storage := newSQLStorage(db)

	return NewService(storage, tools)
}
