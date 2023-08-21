package davsessions

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
)

const tableName = "dav_sessions"

type sqlStorage struct {
	db *sql.DB
}

func newSqlStorage(db *sql.DB) *sqlStorage {
	return &sqlStorage{db}
}

func (t *sqlStorage) Save(ctx context.Context, session *DavSession) error {
	_, err := sq.
		Insert(tableName).
		Columns("id", "username", "password", "user_id", "fs_root", "created_at").
		Values(session.id, session.username, session.password, session.userID, session.fsRoot, session.createdAt).
		RunWith(t.db).
		ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("sql error: %w", err)
	}

	return nil
}

func (t *sqlStorage) GetByUsernamePassword(ctx context.Context, username, password string) (*DavSession, error) {
	res := DavSession{}

	err := sq.
		Select("id", "username", "password", "user_id", "fs_root", "created_at").
		From(tableName).
		Where(sq.Eq{"username": username, "password": password}).
		RunWith(t.db).
		ScanContext(ctx, &res.id, &res.username, &res.password, &res.userID, &res.fsRoot, &res.createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("sql error: %w", err)
	}

	return &res, nil
}
