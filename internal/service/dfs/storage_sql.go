package dfs

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

const tableName = "fs_inodes"

var errNotFound = errors.New("not found")

var allFiels = []string{"id", "name", "parent", "space_id", "size", "last_modified_at", "created_at", "created_by", "file_id"}

type sqlStorage struct {
	db    *sql.DB
	clock clock.Clock
}

func newSqlStorage(db *sql.DB, tools tools.Tools) *sqlStorage {
	return &sqlStorage{db, tools.Clock()}
}

func (s *sqlStorage) Save(ctx context.Context, i *INode) error {
	_, err := sq.
		Insert(tableName).
		Columns(allFiels...).
		Values(i.id, i.name, i.parent, i.spaceID, i.size, i.lastModifiedAt, i.createdAt, i.createdBy, i.fileID).
		RunWith(s.db).
		ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("sql error: %w", err)
	}

	return nil
}

func (s *sqlStorage) GetSpaceRoot(ctx context.Context, spaceID uuid.UUID) (*INode, error) {
	return s.getByKeys(ctx, sq.Eq{"space_id": spaceID, "parent": nil})
}

func (s *sqlStorage) Patch(ctx context.Context, inode uuid.UUID, fields map[string]any) error {
	_, err := sq.Update(tableName).
		SetMap(fields).
		Where(sq.Eq{"id": inode}).
		RunWith(s.db).
		ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("sql error: %w", err)
	}

	return nil
}

func (s *sqlStorage) GetByID(ctx context.Context, id uuid.UUID) (*INode, error) {
	return s.getByKeys(ctx, sq.Eq{"id": id})
}

func (s *sqlStorage) HardDelete(ctx context.Context, id uuid.UUID) error {
	_, err := sq.
		Delete(tableName).
		Where(sq.Eq{"id": string(id)}).
		RunWith(s.db).
		ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("sql error: %w", err)
	}

	return nil
}

func (s *sqlStorage) GetAllChildrens(ctx context.Context, parent uuid.UUID, cmd *storage.PaginateCmd) ([]INode, error) {
	rows, err := storage.PaginateSelection(sq.
		Select(allFiels...).
		Where(sq.Eq{"parent": string(parent), "deleted_at": nil}).
		From(tableName), cmd).
		RunWith(s.db).
		QueryContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("sql error: %w", err)
	}

	defer rows.Close()

	return s.scanRows(rows)
}

func (s *sqlStorage) GetDeleted(ctx context.Context, id uuid.UUID) (*INode, error) {
	return s.getByKeys(ctx, sq.Eq{"id": id}, sq.NotEq{"deleted_at": nil})
}

func (s *sqlStorage) GetAllDeleted(ctx context.Context, limit int) ([]INode, error) {
	rows, err := sq.
		Select(allFiels...).
		From(tableName).
		Where(sq.NotEq{"deleted_at": nil}).
		Limit(uint64(limit)).
		RunWith(s.db).
		QueryContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("sql error: %w", err)
	}

	defer rows.Close()

	return s.scanRows(rows)
}

func (s *sqlStorage) scanRows(rows *sql.Rows) ([]INode, error) {
	inodes := []INode{}

	for rows.Next() {
		var res INode

		err := rows.Scan(&res.id, &res.name, &res.parent, &res.spaceID, &res.size, &res.lastModifiedAt, &res.createdAt, &res.createdBy, &res.fileID)
		if err != nil {
			return nil, fmt.Errorf("failed to scan a row: %w", err)
		}

		inodes = append(inodes, res)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("scan error: %w", err)
	}

	return inodes, nil
}

func (s *sqlStorage) GetSumChildsSize(ctx context.Context, parent uuid.UUID) (uint64, error) {
	var size *uint64

	err := sq.
		Select("SUM(size)").
		From(tableName).
		Where(sq.Eq{"parent": string(parent), "deleted_at": nil}).
		RunWith(s.db).
		ScanContext(ctx, &size)
	if err != nil {
		return 0, fmt.Errorf("sql error: %w", err)
	}

	if size == nil {
		return 0, nil
	}

	return *size, nil
}

func (s *sqlStorage) GetSumRootsSize(ctx context.Context) (uint64, error) {
	var size *uint64

	err := sq.
		Select("SUM(size)").
		From(tableName).
		Where(sq.Eq{"parent": nil, "deleted_at": nil}).
		RunWith(s.db).
		ScanContext(ctx, &size)
	if err != nil {
		return 0, fmt.Errorf("sql error: %w", err)
	}

	if size == nil {
		return 0, nil
	}

	return *size, nil
}

func (s *sqlStorage) GetAllInodesWithFileID(ctx context.Context, fileID uuid.UUID) ([]INode, error) {
	rows, err := sq.
		Select(allFiels...).
		From(tableName).
		Where(sq.Eq{"file_id": fileID}).
		RunWith(s.db).
		QueryContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("sql error: %w", err)
	}

	defer rows.Close()

	return s.scanRows(rows)
}

func (s *sqlStorage) GetByNameAndParent(ctx context.Context, name string, parent uuid.UUID) (*INode, error) {
	res := INode{}

	err := sq.
		Select(allFiels...).
		From(tableName).
		Where(sq.Eq{"parent": string(parent), "name": name, "deleted_at": nil}).
		RunWith(s.db).
		ScanContext(ctx, &res.id, &res.name, &res.parent, &res.spaceID, &res.size, &res.lastModifiedAt, &res.createdAt, &res.createdBy, &res.fileID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("sql error: %w", err)
	}

	return &res, nil
}

func (s *sqlStorage) getByKeys(ctx context.Context, wheres ...any) (*INode, error) {
	res := INode{}

	query := sq.
		Select(allFiels...).
		From(tableName)

	for _, where := range wheres {
		query = query.Where(where)
	}

	err := query.
		RunWith(s.db).
		ScanContext(ctx, &res.id, &res.name, &res.parent, &res.spaceID, &res.size, &res.lastModifiedAt, &res.createdAt, &res.createdBy, &res.fileID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("sql error: %w", err)
	}

	return &res, nil
}
