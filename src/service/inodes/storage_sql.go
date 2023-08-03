package inodes

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
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
		Columns("id", "user_id", "name", "last_modified_at", "created_at").
		Values(dir.ID, dir.UserID, dir.name, &dir.LastModifiedAt, dir.CreatedAt).
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
		Select("id", "user_id", "name", "last_modified_at", "created_at").
		From(tableName).
		Where(sq.Eq{"id": string(id)}).
		RunWith(t.db).
		ScanContext(ctx, &res.ID, &res.UserID, &res.name, &res.LastModifiedAt, &res.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("sql error: %w", err)
	}

	return &res, nil
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
