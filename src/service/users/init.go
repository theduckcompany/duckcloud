package users

import (
	"context"
	"database/sql"

	"github.com/Peltoche/neurone/src/service/inodes"
	"github.com/Peltoche/neurone/src/tools"
	"github.com/Peltoche/neurone/src/tools/uuid"
)

//go:generate mockery --name Service
type Service interface {
	Create(ctx context.Context, user *CreateCmd) (*User, error)
	GetByID(ctx context.Context, userID uuid.UUID) (*User, error)
	Authenticate(ctx context.Context, username, password string) (*User, error)
}

func Init(tools tools.Tools, db *sql.DB, inodes inodes.Service) Service {
	storage := newSqlStorage(db)

	return NewService(tools, storage, inodes)
}
