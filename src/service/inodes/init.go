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
	Get(ctx context.Context, cmd *PathCmd) (*INode, error)
	Readdir(ctx context.Context, cmd *PathCmd, paginateCmd *storage.PaginateCmd) ([]INode, error)
	RemoveAll(ctx context.Context, cmd *PathCmd) error
	GetDeletedINodes(ctx context.Context, limit int) ([]INode, error)
	HardDelete(ctx context.Context, inode uuid.UUID) error
	CreateFile(ctx context.Context, cmd *CreateFileCmd) (*INode, error)
}

func Init(tools tools.Tools, db *sql.DB) Service {
	storage := newSqlStorage(db, tools)

	return NewService(tools, storage)
}
