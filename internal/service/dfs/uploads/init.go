package uploads

import (
	context "context"
	"database/sql"

	"github.com/theduckcompany/duckcloud/internal/tools"
)

//go:generate mockery --name Service
type Service interface {
	Register(ctx context.Context, cmd *RegisterUploadCmd) error
	GetOldest(ctx context.Context) (*Upload, error)
	Delete(ctx context.Context, upload *Upload) error
}

func Init(db *sql.DB, tools tools.Tools) Service {
	storage := newSqlStorage(db)

	return NewService(storage, tools)
}
