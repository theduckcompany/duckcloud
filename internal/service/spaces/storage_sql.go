package spaces

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

const tableName = "spaces"

var errNotFound = errors.New("not found")

var allFields = []string{"id", "name", "public", "owners", "created_at", "created_by"}

type sqlStorage struct {
	db    *sql.DB
	clock clock.Clock
}

func newSqlStorage(db *sql.DB, tools tools.Tools) *sqlStorage {
	return &sqlStorage{db, tools.Clock()}
}

func (s *sqlStorage) Save(ctx context.Context, space *Space) error {
	_, err := sq.
		Insert(tableName).
		Columns(allFields...).
		Values(space.id, space.name, space.isPublic, space.owners, space.createdAt, space.createdBy).
		RunWith(s.db).
		ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("sql error: %w", err)
	}

	return nil
}

func (s *sqlStorage) GetAllSpaces(ctx context.Context, cmd *storage.PaginateCmd) ([]Space, error) {
	return s.getAllbyKeys(ctx, cmd)
}

func (s *sqlStorage) GetAllUserSpaces(ctx context.Context, userID uuid.UUID, cmd *storage.PaginateCmd) ([]Space, error) {
	return s.getAllbyKeys(ctx, cmd, sq.Like{"owners": fmt.Sprintf("%%%s%%", userID)})
}

func (s *sqlStorage) GetByID(ctx context.Context, id uuid.UUID) (*Space, error) {
	return s.getByKeys(ctx, sq.Eq{"id": id})
}

func (s *sqlStorage) Delete(ctx context.Context, spaceID uuid.UUID) error {
	_, err := sq.
		Delete(tableName).
		Where(sq.Eq{"id": spaceID}).
		RunWith(s.db).
		ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("sql error: %w", err)
	}

	return nil
}

func (s *sqlStorage) Patch(ctx context.Context, spaceID uuid.UUID, fields map[string]any) error {
	_, err := sq.Update(tableName).
		SetMap(fields).
		Where(sq.Eq{"id": spaceID}).
		RunWith(s.db).
		ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("sql error: %w", err)
	}

	return nil
}

func (s *sqlStorage) getAllbyKeys(ctx context.Context, cmd *storage.PaginateCmd, wheres ...any) ([]Space, error) {
	query := sq.
		Select(allFields...).
		From(tableName)

	for _, where := range wheres {
		query = query.Where(where)
	}

	query = storage.PaginateSelection(query, cmd)

	rows, err := query.RunWith(s.db).
		QueryContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("sql error: %w", err)
	}

	defer rows.Close()

	return s.scanRows(rows)
}

func (s *sqlStorage) getByKeys(ctx context.Context, wheres ...any) (*Space, error) {
	res := Space{}

	query := sq.
		Select(allFields...).
		From(tableName)

	for _, where := range wheres {
		query = query.Where(where)
	}

	err := query.
		RunWith(s.db).
		ScanContext(ctx, &res.id, &res.name, &res.isPublic, &res.owners, &res.createdAt, &res.createdBy)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("sql error: %w", err)
	}

	return &res, nil
}

func (s *sqlStorage) scanRows(rows *sql.Rows) ([]Space, error) {
	spaces := []Space{}

	for rows.Next() {
		var res Space

		err := rows.Scan(&res.id, &res.name, &res.isPublic, &res.owners, &res.createdAt, &res.createdBy)
		if err != nil {
			return nil, fmt.Errorf("failed to scan a row: %w", err)
		}

		spaces = append(spaces, res)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("scan error: %w", err)
	}

	return spaces, nil
}
