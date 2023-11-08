package folders

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

const tableName = "fs_folders"

var errNotFound = errors.New("not found")

var allFields = []string{"id", "name", "public", "owners", "root_fs", "created_at"}

type sqlStorage struct {
	db    *sql.DB
	clock clock.Clock
}

func newSqlStorage(db *sql.DB, tools tools.Tools) *sqlStorage {
	return &sqlStorage{db, tools.Clock()}
}

func (s *sqlStorage) Save(ctx context.Context, folder *Folder) error {
	_, err := sq.
		Insert(tableName).
		Columns(allFields...).
		Values(folder.id, folder.name, folder.isPublic, folder.owners, folder.rootFS, folder.createdAt).
		RunWith(s.db).
		ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("sql error: %w", err)
	}

	return nil
}

func (s *sqlStorage) GetAllUserFolders(ctx context.Context, userID uuid.UUID, cmd *storage.PaginateCmd) ([]Folder, error) {
	return s.getAllbyKeys(ctx, cmd, sq.Like{"owners": fmt.Sprintf("%%%s%%", userID)})
}

func (s *sqlStorage) GetByID(ctx context.Context, id uuid.UUID) (*Folder, error) {
	return s.getByKeys(ctx, sq.Eq{"id": id})
}

func (s *sqlStorage) Delete(ctx context.Context, folderID uuid.UUID) error {
	_, err := sq.
		Delete(tableName).
		Where(sq.Eq{"id": folderID}).
		RunWith(s.db).
		ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("sql error: %w", err)
	}

	return nil
}

func (s *sqlStorage) Patch(ctx context.Context, folderID uuid.UUID, fields map[string]any) error {
	_, err := sq.Update(tableName).
		SetMap(fields).
		Where(sq.Eq{"id": folderID}).
		RunWith(s.db).
		ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("sql error: %w", err)
	}

	return nil
}

func (s *sqlStorage) getAllbyKeys(ctx context.Context, cmd *storage.PaginateCmd, wheres ...any) ([]Folder, error) {
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

func (s *sqlStorage) getByKeys(ctx context.Context, wheres ...any) (*Folder, error) {
	res := Folder{}

	query := sq.
		Select(allFields...).
		From(tableName)

	for _, where := range wheres {
		query = query.Where(where)
	}

	err := query.
		RunWith(s.db).
		ScanContext(ctx, &res.id, &res.name, &res.isPublic, &res.owners, &res.rootFS, &res.createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("sql error: %w", err)
	}

	return &res, nil
}

func (s *sqlStorage) scanRows(rows *sql.Rows) ([]Folder, error) {
	folders := []Folder{}

	for rows.Next() {
		var res Folder

		err := rows.Scan(&res.id, &res.name, &res.isPublic, &res.owners, &res.rootFS, &res.createdAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan a row: %w", err)
		}

		folders = append(folders, res)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("scan error: %w", err)
	}

	return folders, nil
}