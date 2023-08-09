package inodes

import (
	"context"
	"database/sql"

	"github.com/Peltoche/neurone/src/tools"
	"github.com/Peltoche/neurone/src/tools/storage"
	"github.com/Peltoche/neurone/src/tools/uuid"
)

//go:generate mockery --name Service
type Service interface {
	BootstrapUser(ctx context.Context, userID uuid.UUID) (*INode, error)
	Mkdir(ctx context.Context, cmd *PathCmd) (*INode, error)
	Open(ctx context.Context, cmd *PathCmd) (*INode, error)
	Readdir(ctx context.Context, cmd *PathCmd, paginateCmd *storage.PaginateCmd) ([]INode, error)
}

func Init(tools tools.Tools, db *sql.DB) Service {
	storage := newSqlStorage(db)

	return NewService(tools, storage)
}
