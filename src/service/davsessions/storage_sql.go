package davsessions

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/theduckcompany/duckcloud/src/tools/storage"
	"github.com/theduckcompany/duckcloud/src/tools/uuid"
)

const tableName = "dav_sessions"

var allFields = []string{"id", "username", "name", "password", "user_id", "folders", "created_at"}

type sqlStorage struct {
	db *sql.DB
}

func newSqlStorage(db *sql.DB) *sqlStorage {
	return &sqlStorage{db}
}

func (t *sqlStorage) Save(ctx context.Context, session *DavSession) error {
	_, err := sq.
		Insert(tableName).
		Columns(allFields...).
		Values(session.id, session.username, session.name, session.password, session.userID, session.folders, session.createdAt).
		RunWith(t.db).
		ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("sql error: %w", err)
	}

	return nil
}

func (t *sqlStorage) GetByID(ctx context.Context, sessionID uuid.UUID) (*DavSession, error) {
	res := DavSession{}

	err := sq.
		Select(allFields...).
		From(tableName).
		Where(sq.Eq{"id": sessionID}).
		RunWith(t.db).
		ScanContext(ctx, &res.id, &res.username, &res.name, &res.password, &res.userID, &res.folders, &res.createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("sql error: %w", err)
	}

	return &res, nil
}

func (t *sqlStorage) RemoveByID(ctx context.Context, sessionID uuid.UUID) error {
	_, err := sq.
		Delete(tableName).
		Where(sq.Eq{"id": sessionID}).
		RunWith(t.db).
		ExecContext(ctx)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("sql error: %w", err)
	}

	return nil
}

func (t *sqlStorage) GetByUsernameAndPassHash(ctx context.Context, username, password string) (*DavSession, error) {
	res := DavSession{}

	err := sq.
		Select(allFields...).
		From(tableName).
		Where(sq.Eq{"username": username, "password": password}).
		RunWith(t.db).
		ScanContext(ctx, &res.id, &res.username, &res.name, &res.password, &res.userID, &res.folders, &res.createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("sql error: %w", err)
	}

	return &res, nil
}

func (s *sqlStorage) GetAllForUser(ctx context.Context, userID uuid.UUID, cmd *storage.PaginateCmd) ([]DavSession, error) {
	rows, err := storage.PaginateSelection(sq.
		Select(allFields...).
		Where(sq.Eq{"user_id": string(userID)}).
		From(tableName), cmd).
		RunWith(s.db).
		QueryContext(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return []DavSession{}, nil
	}

	if err != nil {
		return nil, fmt.Errorf("sql error: %w", err)
	}

	defer rows.Close()

	return s.scanRows(rows)
}

func (s *sqlStorage) scanRows(rows *sql.Rows) ([]DavSession, error) {
	inodes := []DavSession{}

	for rows.Next() {
		var res DavSession

		err := rows.Scan(&res.id, &res.username, &res.name, &res.password, &res.userID, &res.folders, &res.createdAt)
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
