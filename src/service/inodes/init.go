package inodes

import (
	"context"
	"database/sql"
	"hash"

	"github.com/theduckcompany/duckcloud/src/tools"
	"github.com/theduckcompany/duckcloud/src/tools/storage"
	"github.com/theduckcompany/duckcloud/src/tools/uuid"
)

//go:generate mockery --name Service
type Service interface {
	CreateRootDir(ctx context.Context) (*INode, error)
	Get(ctx context.Context, cmd *PathCmd) (*INode, error)
	GetByID(ctx context.Context, inodeID uuid.UUID) (*INode, error)
	Readdir(ctx context.Context, cmd *PathCmd, paginateCmd *storage.PaginateCmd) ([]INode, error)
	RemoveAll(ctx context.Context, cmd *PathCmd) error
	GetAllDeleted(ctx context.Context, limit int) ([]INode, error)
	HardDelete(ctx context.Context, inode uuid.UUID) error
	CreateDir(ctx context.Context, cmd *PathCmd) (*INode, error)
	CreateFile(ctx context.Context, cmd *CreateFileCmd) (*INode, error)
	RegisterWrite(ctx context.Context, inode *INode, sizeWrite int, h hash.Hash) error
	GetINodeRoot(ctx context.Context, inode *INode) (*INode, error)
}

func Init(tools tools.Tools, db *sql.DB) Service {
	storage := newSqlStorage(db, tools)

	return NewService(tools, storage)
}
