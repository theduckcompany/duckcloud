package inodes

import (
	"context"
	"database/sql"
	"time"

	"github.com/theduckcompany/duckcloud/internal/service/tasks/scheduler"
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
	HardDelete(ctx context.Context, inode *INode) error
	GetByNameAndParent(ctx context.Context, name string, parent uuid.UUID) (*INode, error)
	CreateDir(ctx context.Context, parent *INode, name string) (*INode, error)
	CreateFile(ctx context.Context, cmd *CreateFileCmd) (*INode, error)
	MkdirAll(ctx context.Context, cmd *PathCmd) (*INode, error)
	PatchMove(ctx context.Context, source, parent *INode, newName string, modeTime time.Time) (*INode, error)
	GetSumChildsSize(ctx context.Context, parent uuid.UUID) (uint64, error)
	RegisterModification(ctx context.Context, inode *INode, newSize uint64, modeTime time.Time) error
}

func Init(scheduler scheduler.Service, tools tools.Tools, db *sql.DB) Service {
	storage := newSqlStorage(db, tools)

	return NewService(scheduler, tools, storage)
}
