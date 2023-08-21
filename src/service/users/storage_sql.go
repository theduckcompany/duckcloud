package users

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/theduckcompany/duckcloud/src/tools/uuid"
)

const tableName = "users"

// sqlStorage use to save/retrieve Users
type sqlStorage struct {
	db *sql.DB
}

// newSqlStorage instantiates a new Storage based on sql.
func newSqlStorage(db *sql.DB) *sqlStorage {
	return &sqlStorage{db}
}

// Save the given User.
func (t *sqlStorage) Save(ctx context.Context, user *User) error {
	_, err := sq.
		Insert(tableName).
		Columns("id", "username", "fs_root", "password", "created_at").
		Values(user.id, user.username, user.fsRoot, user.password, user.createdAt).
		RunWith(t.db).
		ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("sql error: %w", err)
	}

	return nil
}

// GetByID return the user matching the id.
//
// If no user is found, nil and no error will be returned.
func (t *sqlStorage) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	return t.getByKey(ctx, "id", string(id))
}

func (t *sqlStorage) GetByUsername(ctx context.Context, username string) (*User, error) {
	return t.getByKey(ctx, "username", username)
}

func (t *sqlStorage) getByKey(ctx context.Context, key, expected string) (*User, error) {
	res := User{}

	err := sq.
		Select("id", "username", "fs_root", "password", "created_at").
		From(tableName).
		Where(sq.Eq{key: expected}).
		RunWith(t.db).
		ScanContext(ctx, &res.id, &res.username, &res.fsRoot, &res.password, &res.createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("sql error: %w", err)
	}

	return &res, nil
}
