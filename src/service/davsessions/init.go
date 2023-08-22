package davsessions

import (
	"context"
	"database/sql"

	"github.com/theduckcompany/duckcloud/src/service/inodes"
	"github.com/theduckcompany/duckcloud/src/service/users"
	"github.com/theduckcompany/duckcloud/src/tools"
	"github.com/theduckcompany/duckcloud/src/tools/storage"
	"github.com/theduckcompany/duckcloud/src/tools/uuid"
)

//go:generate mockery --name Service
type Service interface {
	GetAllForUser(ctx context.Context, userID uuid.UUID, paginateCmd *storage.PaginateCmd) ([]DavSession, error)
}

func Init(db *sql.DB, inodes inodes.Service, users users.Service, tools tools.Tools) Service {
	storage := newSqlStorage(db)

	return NewService(storage, inodes, users, tools)
}
