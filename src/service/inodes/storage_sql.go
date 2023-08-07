package inodes

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/Peltoche/neurone/src/tools/storage"
	"github.com/Peltoche/neurone/src/tools/uuid"
)

const tableName = "fs_inodes"

type sqlStorage struct {
	db *sql.DB
}

func newSqlStorage(db *sql.DB) *sqlStorage {
	return &sqlStorage{db}
}

func (t *sqlStorage) Save(ctx context.Context, dir *INode) error {
	_, err := sq.
		Insert(tableName).
		Columns("id", "user_id", "name", "parent", "last_modified_at", "created_at").
		Values(dir.ID, dir.UserID, dir.name, dir.Parent, dir.LastModifiedAt, dir.CreatedAt).
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
		Select("id", "user_id", "name", "parent", "last_modified_at", "created_at").
		From(tableName).
		Where(sq.Eq{"id": string(id)}).
		RunWith(t.db).
		ScanContext(ctx, &res.ID, &res.UserID, &res.name, &res.Parent, &res.LastModifiedAt, &res.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		fmt.Printf("err no rows\n\n\n")
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("sql error: %w", err)
	}

	return &res, nil
}

func (t *sqlStorage) GetAllChildrens(ctx context.Context, userID, parent uuid.UUID, cmd *storage.PaginateCmd) ([]INode, error) {
	rows, err := storage.PaginateSelection(sq.
		Select("id", "user_id", "name", "parent", "last_modified_at", "created_at").
		Where(sq.Eq{"user_id": string(userID), "parent": string(parent)}).
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

		err = rows.Scan(&res.ID, &res.UserID, &res.name, &res.Parent, &res.LastModifiedAt, &res.CreatedAt)
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
		Where(sq.Eq{"user_id": string(userID)}).
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
		Select("id", "user_id", "name", "parent", "last_modified_at", "created_at").
		From(tableName).
		Where(sq.Eq{"user_id": string(userID), "parent": string(parent), "name": name}).
		RunWith(t.db).
		ScanContext(ctx, &res.ID, &res.UserID, &res.name, &res.Parent, &res.LastModifiedAt, &res.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("sql error: %w", err)
	}

	return &res, nil
}
