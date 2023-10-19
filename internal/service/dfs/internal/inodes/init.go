package inodes

import (
	"context"
	"database/sql"
	"time"

	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

//go:generate mockery --name Service
type Service interface {
	CreateRootDir(ctx context.Context) (*INode, error)
	Get(ctx context.Context, cmd *PathCmd) (*INode, error)
	GetByID(ctx context.Context, inodeID uuid.UUID) (*INode, error)
	Readdir(ctx context.Context, inode *INode, paginateCmd *storage.PaginateCmd) ([]INode, error)
	Remove(ctx context.Context, inode *INode) error
	GetAllDeleted(ctx context.Context, limit int) ([]INode, error)
	HardDelete(ctx context.Context, inode uuid.UUID) error
	GetByNameAndParent(ctx context.Context, name string, parent uuid.UUID) (*INode, error)
	CreateDir(ctx context.Context, parent *INode, name string) (*INode, error)
	CreateFile(ctx context.Context, cmd *CreateFileCmd) (*INode, error)
	RegisterWrite(ctx context.Context, inode *INode, sizeWrite int64, modeTime time.Time) error
	MkdirAll(ctx context.Context, cmd *PathCmd) (*INode, error)
	Move(ctx context.Context, source *INode, into *PathCmd) (*INode, error)
}

func Init(tools tools.Tools, db *sql.DB) Service {
	storage := newSqlStorage(db, tools)

	return NewService(tools, storage)
}
