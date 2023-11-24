package users

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/clock"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

const tableName = "users"

var errNotFound = errors.New("not found")

var allFields = []string{"id", "username", "admin", "status", "space", "password", "created_at", "created_by"}

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
func (s *sqlStorage) Save(ctx context.Context, u *User) error {
	_, err := sq.
		Insert(tableName).
		Columns(allFields...).
		Values(u.id, u.username, u.isAdmin, u.status, u.defaultSpaceID, u.password, u.createdAt, u.createdBy).
		RunWith(s.db).
		ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("sql error: %w", err)
	}

	return nil
}

func (s *sqlStorage) GetAll(ctx context.Context, cmd *storage.PaginateCmd) ([]User, error) {
	rows, err := storage.PaginateSelection(sq.
		Select(allFields...).
		From(tableName), cmd).
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

func (s *sqlStorage) Patch(ctx context.Context, userID uuid.UUID, fields map[string]any) error {
	_, err := sq.Update(tableName).
		SetMap(fields).
		Where(sq.Eq{"id": userID}).
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

func (s *sqlStorage) getByKeys(ctx context.Context, wheres ...any) (*User, error) {
	res := User{}

	query := sq.
		Select(allFields...).
		From(tableName)

	for _, where := range wheres {
		query = query.Where(where)
	}

	err := query.
		RunWith(s.db).
		ScanContext(ctx, &res.id, &res.username, &res.isAdmin, &res.status, &res.defaultSpaceID, &res.password, &res.createdAt, &res.createdBy)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errNotFound
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

		err := rows.Scan(&res.id, &res.username, &res.isAdmin, &res.status, &res.defaultSpaceID, &res.password, &res.createdAt, &res.createdBy)
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
