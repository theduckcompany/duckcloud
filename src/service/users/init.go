package users

import (
	"context"
	"database/sql"

	"github.com/Peltoche/neurone/src/tools"
	"github.com/Peltoche/neurone/src/tools/uuid"
)

// UserService encapsulates usecase logic for users.
type Service interface {
	Create(ctx context.Context, user *CreateUserRequest) (*User, error)
	GetByID(ctx context.Context, userID uuid.UUID) (*User, error)
	Authenticate(ctx context.Context, username, password string) (*User, error)
}

func Init(tools tools.Tools, db *sql.DB) Service {
	storage := newSqlStorage(db)

	return NewService(tools, storage)
}
