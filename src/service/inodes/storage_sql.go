package inodes

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

const tableName = "fs_inodes"

type sqlStorage struct {
	db    *sql.DB
	clock clock.Clock
}

func newSqlStorage(db *sql.DB, tools tools.Tools) *sqlStorage {
	return &sqlStorage{db, tools.Clock()}
}

func (s *sqlStorage) Save(ctx context.Context, dir *INode) error {
	_, err := sq.
		Insert(tableName).
		Columns("id", "user_id", "name", "parent", "mode", "last_modified_at", "created_at").
		Values(dir.id, dir.userID, dir.name, dir.parent, dir.mode, dir.lastModifiedAt, dir.createdAt).
		RunWith(s.db).
		ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("sql error: %w", err)
	}

	return nil
}

func (s *sqlStorage) GetByID(ctx context.Context, id uuid.UUID) (*INode, error) {
	res := INode{}

	err := sq.
		Select("id", "user_id", "name", "parent", "mode", "last_modified_at", "created_at").
		From(tableName).
		Where(sq.Eq{"id": string(id)}).
		RunWith(s.db).
		ScanContext(ctx, &res.id, &res.userID, &res.name, &res.parent, &res.mode, &res.lastModifiedAt, &res.createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("sql error: %w", err)
	}

	return &res, nil
}

func (s *sqlStorage) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := sq.
		Update(tableName).
		Where(sq.Eq{"id": string(id)}).
		Set("deleted_at", s.clock.Now()).
		RunWith(s.db).
		ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("sql error: %w", err)
	}

	return nil
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

func (s *sqlStorage) GetAllChildrens(ctx context.Context, userID, parent uuid.UUID, cmd *storage.PaginateCmd) ([]INode, error) {
	rows, err := storage.PaginateSelection(sq.
		Select("id", "user_id", "name", "parent", "mode", "last_modified_at", "created_at").
		Where(sq.Eq{"user_id": string(userID), "parent": string(parent), "deleted_at": nil}).
		From(tableName), cmd).
		RunWith(s.db).
		QueryContext(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return []INode{}, nil
	}

	if err != nil {
		return nil, fmt.Errorf("sql error: %w", err)
	}

	defer rows.Close()

	return s.scanRows(rows)
}

func (s *sqlStorage) GetDeletedINodes(ctx context.Context, limit int) ([]INode, error) {
	rows, err := sq.
		Select("id", "user_id", "name", "parent", "mode", "last_modified_at", "created_at").
		From(tableName).
		Where(sq.NotEq{"deleted_at": nil}).
		Limit(uint64(limit)).
		RunWith(s.db).
		QueryContext(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return []INode{}, nil
	}

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

		err := rows.Scan(&res.id, &res.userID, &res.name, &res.parent, &res.mode, &res.lastModifiedAt, &res.createdAt)
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

func (s *sqlStorage) CountUserINodes(ctx context.Context, userID uuid.UUID) (uint, error) {
	var count uint

	err := sq.
		Select("count(*)").
		From(tableName).
		Where(sq.Eq{"user_id": string(userID), "deleted_at": nil}).
		RunWith(s.db).
		ScanContext(ctx, &count)
	if err != nil {
		return 0, fmt.Errorf("sql error: %w", err)
	}

	return count, nil
}

func (s *sqlStorage) GetByNameAndParent(ctx context.Context, userID uuid.UUID, name string, parent uuid.UUID) (*INode, error) {
	res := INode{}

	err := sq.
		Select("id", "user_id", "name", "parent", "mode", "last_modified_at", "created_at").
		From(tableName).
		Where(sq.Eq{"user_id": string(userID), "parent": string(parent), "name": name, "deleted_at": nil}).
		RunWith(s.db).
		ScanContext(ctx, &res.id, &res.userID, &res.name, &res.parent, &res.mode, &res.lastModifiedAt, &res.createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("sql error: %w", err)
	}

	return &res, nil
}
