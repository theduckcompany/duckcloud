package users

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/theduckcompany/duckcloud/src/tools"
	"github.com/theduckcompany/duckcloud/src/tools/clock"
	"github.com/theduckcompany/duckcloud/src/tools/storage"
	"github.com/theduckcompany/duckcloud/src/tools/uuid"
)

const tableName = "users"

// sqlStorage use to save/retrieve Users
type sqlStorage struct {
	db    *sql.DB
	clock clock.Clock
}

// newSqlStorage instantiates a new Storage based on sql.
func newSqlStorage(db *sql.DB, tools tools.Tools) *sqlStorage {
	return &sqlStorage{db, tools.Clock()}
}

// Save the given User.
func (s *sqlStorage) Save(ctx context.Context, user *User) error {
	_, err := sq.
		Insert(tableName).
		Columns("id", "username", "admin", "fs_root", "password", "created_at").
		Values(user.id, user.username, user.isAdmin, user.fsRoot, user.password, user.createdAt).
		RunWith(s.db).
		ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("sql error: %w", err)
	}

	return nil
}

func (s *sqlStorage) GetAll(ctx context.Context, cmd *storage.PaginateCmd) ([]User, error) {
	rows, err := storage.PaginateSelection(sq.
		Select("id", "username", "admin", "fs_root", "password", "created_at").
		From(tableName), cmd).
		Where(sq.Eq{"deleted_at": nil}).
		RunWith(s.db).
		QueryContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("sql error: %w", err)
	}

	return s.scanRows(rows)
}

func (s *sqlStorage) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	return s.getByKeys(ctx, sq.Eq{"id": id})
}

func (s *sqlStorage) GetByUsername(ctx context.Context, username string) (*User, error) {
	return s.getByKeys(ctx, sq.Eq{"username": username})
}

func (s *sqlStorage) GetDeleted(ctx context.Context, id uuid.UUID) (*User, error) {
	return s.getByKeys(ctx, sq.Eq{"id": id}, sq.NotEq{"deleted_at": nil})
}

func (s *sqlStorage) Delete(ctx context.Context, userID uuid.UUID) error {
	_, err := sq.
		Update(tableName).
		Where(sq.Eq{"id": userID}).
		Set("deleted_at", s.clock.Now()).
		RunWith(s.db).
		ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("sql error: %w", err)
	}

	return nil
}

func (s *sqlStorage) HardDelete(ctx context.Context, userID uuid.UUID) error {
	_, err := sq.
		Delete(tableName).
		Where(sq.Eq{"id": string(userID)}).
		RunWith(s.db).
		ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("sql error: %w", err)
	}

	return nil
}

func (s *sqlStorage) GetAllDeleted(ctx context.Context, limit int) ([]User, error) {
	rows, err := sq.
		Select("id", "username", "admin", "fs_root", "password", "created_at").
		From(tableName).
		Where(sq.NotEq{"deleted_at": nil}).
		Limit(uint64(limit)).
		RunWith(s.db).
		QueryContext(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return []User{}, nil
	}

	if err != nil {
		return nil, fmt.Errorf("sql error: %w", err)
	}

	defer rows.Close()

	return s.scanRows(rows)
}

func (s *sqlStorage) getByKeys(ctx context.Context, wheres ...any) (*User, error) {
	res := User{}

	query := sq.
		Select("id", "username", "admin", "fs_root", "password", "created_at").
		From(tableName)

	for _, where := range wheres {
		query = query.Where(where)
	}

	err := query.
		RunWith(s.db).
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
