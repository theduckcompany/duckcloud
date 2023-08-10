package inodes

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/Peltoche/neurone/src/tools"
	"github.com/Peltoche/neurone/src/tools/clock"
	"github.com/Peltoche/neurone/src/tools/storage"
	"github.com/Peltoche/neurone/src/tools/uuid"
)

const tableName = "fs_inodes"

type sqlStorage struct {
	db    *sql.DB
	clock clock.Clock
}

func newSqlStorage(db *sql.DB, tools tools.Tools) *sqlStorage {
	return &sqlStorage{db, tools.Clock()}
}

func (t *sqlStorage) Save(ctx context.Context, dir *INode) error {
	_, err := sq.
		Insert(tableName).
		Columns("id", "user_id", "name", "parent", "mode", "last_modified_at", "created_at").
		Values(dir.id, dir.userID, dir.name, dir.parent, dir.mode, dir.lastModifiedAt, dir.createdAt).
		RunWith(t.db).
		ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("sql error: %w", err)
	}

	return nil
}

func (t *sqlStorage) GetByID(ctx context.Context, id uuid.UUID) (*INode, error) {
	res := INode{}

	err := sq.
		Select("id", "user_id", "name", "parent", "mode", "last_modified_at", "created_at").
		From(tableName).
		Where(sq.Eq{"id": string(id), "deleted_at": nil}).
		RunWith(t.db).
		ScanContext(ctx, &res.id, &res.userID, &res.name, &res.parent, &res.mode, &res.lastModifiedAt, &res.createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("sql error: %w", err)
	}

	return &res, nil
}

func (t *sqlStorage) Delete(ctx context.Context, id uuid.UUID) error {
	res, err := sq.
		Update(tableName).
		Where(sq.Eq{"id": string(id)}).
		Set("deleted_at", t.clock.Now()).
		RunWith(t.db).
		ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("sql error: %w", err)
	}

	nb, _ := res.RowsAffected()
	fmt.Printf("row affected: %d\n\n", nb)

	return nil
}

func (t *sqlStorage) HardDelete(ctx context.Context, id uuid.UUID) error {
	_, err := sq.
		Delete(tableName).
		Where(sq.Eq{"id": string(id)}).
		RunWith(t.db).
		ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("sql error: %w", err)
	}

	return nil
}

func (t *sqlStorage) GetAllChildrens(ctx context.Context, userID, parent uuid.UUID, cmd *storage.PaginateCmd) ([]INode, error) {
	rows, err := storage.PaginateSelection(sq.
		Select("id", "user_id", "name", "parent", "mode", "last_modified_at", "created_at").
		Where(sq.Eq{"user_id": string(userID), "parent": string(parent), "deleted_at": nil}).
		From(tableName), cmd).
		RunWith(t.db).
		QueryContext(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	defer rows.Close()

	inodes := []INode{}
	for rows.Next() {
		var res INode

		err = rows.Scan(&res.id, &res.userID, &res.name, &res.parent, &res.mode, &res.lastModifiedAt, &res.createdAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan a row: %w", err)
		}

		inodes = append(inodes, res)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("scan error: %w", err)
	}

	if err != nil {
		return nil, fmt.Errorf("sql error: %w", err)
	}

	return inodes, nil
}

func (t *sqlStorage) CountUserINodes(ctx context.Context, userID uuid.UUID) (uint, error) {
	var count uint

	err := sq.
		Select("count(*)").
		From(tableName).
		Where(sq.Eq{"user_id": string(userID), "deleted_at": nil}).
		RunWith(t.db).
		ScanContext(ctx, &count)
	if err != nil {
		return 0, fmt.Errorf("sql error: %w", err)
	}

	return count, nil
}

func (t *sqlStorage) GetByNameAndParent(ctx context.Context, userID uuid.UUID, name string, parent uuid.UUID) (*INode, error) {
	res := INode{}

	err := sq.
		Select("id", "user_id", "name", "parent", "mode", "last_modified_at", "created_at").
		From(tableName).
		Where(sq.Eq{"user_id": string(userID), "parent": string(parent), "name": name, "deleted_at": nil}).
		RunWith(t.db).
		ScanContext(ctx, &res.id, &res.userID, &res.name, &res.parent, &res.mode, &res.lastModifiedAt, &res.createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("sql error: %w", err)
	}

	return &res, nil
}
