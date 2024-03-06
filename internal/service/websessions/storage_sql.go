package websessions

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/theduckcompany/duckcloud/internal/tools/ptr"
	"github.com/theduckcompany/duckcloud/internal/tools/secret"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

const tableName = "web_sessions"

var errNotFound = errors.New("not found")

var allFields = []string{"token", "user_id", "ip", "device", "created_at"}

type sqlStorage struct {
	db *sql.DB
}

func newSQLStorage(db *sql.DB) *sqlStorage {
	return &sqlStorage{db}
}

func (s *sqlStorage) Save(ctx context.Context, session *Session) error {
	_, err := sq.
		Insert(tableName).
		Columns(allFields...).
		Values(session.token, session.userID, session.ip, session.device, ptr.To(storage.SQLTime(session.createdAt))).
		RunWith(s.db).
		ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("sql error: %w", err)
	}

	return nil
}

func (s *sqlStorage) GetByToken(ctx context.Context, token secret.Text) (*Session, error) {
	var res Session
	var sqlCreatedAt storage.SQLTime

	err := sq.
		Select(allFields...).
		From(tableName).
		Where(sq.Eq{"token": token}).
		RunWith(s.db).
		ScanContext(ctx, &res.token, &res.userID, &res.ip, &res.device, &sqlCreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("sql error: %w", err)
	}

	res.createdAt = sqlCreatedAt.Time()

	return &res, nil
}

func (s *sqlStorage) RemoveByToken(ctx context.Context, token secret.Text) error {
	_, err := sq.
		Delete(tableName).
		Where(sq.Eq{"token": token}).
		RunWith(s.db).
		ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("sql error: %w", err)
	}

	return nil
}

func (s *sqlStorage) GetAllForUser(ctx context.Context, userID uuid.UUID, cmd *storage.PaginateCmd) ([]Session, error) {
	sessions := []Session{}

	rows, err := storage.PaginateSelection(sq.
		Select(allFields...).
		From(tableName).
		Where(sq.Eq{"user_id": userID}), cmd).
		RunWith(s.db).
		QueryContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("sql error: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var res Session
		var sqlCreatedAt storage.SQLTime

		err = rows.Scan(&res.token, &res.userID, &res.ip, &res.device, &sqlCreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan a row: %w", err)
		}

		res.createdAt = sqlCreatedAt.Time()

		sessions = append(sessions, res)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("scan error: %w", err)
	}

	return sessions, nil
}
