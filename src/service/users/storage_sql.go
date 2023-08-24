package users

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/theduckcompany/duckcloud/src/tools/storage"
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
		Columns("id", "username", "admin", "fs_root", "password", "created_at").
		Values(user.id, user.username, user.isAdmin, user.fsRoot, user.password, user.createdAt).
		RunWith(t.db).
		ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("sql error: %w", err)
	}

	return nil
}

func (t *sqlStorage) GetAll(ctx context.Context, cmd *storage.PaginateCmd) ([]User, error) {
	rows, err := storage.PaginateSelection(sq.
		Select("id", "username", "admin", "fs_root", "password", "created_at").
		From(tableName), cmd).
		RunWith(t.db).
		QueryContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("sql error: %w", err)
	}

	return t.scanRows(rows)
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
		Select("id", "username", "admin", "fs_root", "password", "created_at").
		From(tableName).
		Where(sq.Eq{key: expected}).
		RunWith(t.db).
		ScanContext(ctx, &res.id, &res.username, &res.isAdmin, &res.fsRoot, &res.password, &res.createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("sql error: %w", err)
	}

	return &res, nil
}

func (s *sqlStorage) scanRows(rows *sql.Rows) ([]User, error) {
	users := []User{}

	for rows.Next() {
		var res User

		err := rows.Scan(&res.id, &res.username, &res.isAdmin, &res.fsRoot, &res.password, &res.createdAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan a row: %w", err)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to scan a row: %w", err)
		}

		users = append(users, res)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("scan error: %w", err)
	}

	return users, nil
}
