package davsessions

import (
	"database/sql"

	"github.com/theduckcompany/duckcloud/src/service/inodes"
	"github.com/theduckcompany/duckcloud/src/service/users"
	"github.com/theduckcompany/duckcloud/src/tools"
)

//go:generate mockery --name Service
type Service interface{}

func Init(db *sql.DB, inodes inodes.Service, users users.Service, tools tools.Tools) Service {
	storage := newSqlStorage(db)

	return NewService(storage, inodes, users, tools)
}
