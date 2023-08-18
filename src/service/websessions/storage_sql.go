package websessions

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/myminicloud/myminicloud/src/tools/uuid"
)

const tableName = "web_sessions"

type sqlStorage struct {
	db *sql.DB
}

func newSQLStorage(db *sql.DB) *sqlStorage {
	return &sqlStorage{db}
}

func (s *sqlStorage) Save(ctx context.Context, session *Session) error {
	_, err := sq.
		Insert(tableName).
		Columns("token", "user_id", "ip", "client_id", "device", "created_at").
		Values(session.token, session.userID, session.ip, session.clientID, session.device, session.createdAt).
		RunWith(s.db).
		ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("sql error: %w", err)
	}

	return nil
}

func (s *sqlStorage) GetByToken(ctx context.Context, token string) (*Session, error) {
	res := Session{}

	err := sq.
		Select("token", "user_id", "ip", "client_id", "device", "created_at").
		From(tableName).
		Where(sq.Eq{"token": token}).
		RunWith(s.db).
		ScanContext(ctx, &res.token, &res.userID, &res.ip, &res.clientID, &res.device, &res.createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("sql error: %w", err)
	}

	return &res, nil
}

func (s *sqlStorage) RemoveByToken(ctx context.Context, token string) error {
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

func (s *sqlStorage) GetAllForUser(ctx context.Context, userID uuid.UUID) ([]Session, error) {
	sessions := []Session{}

	rows, err := sq.
		Select("token", "user_id", "ip", "client_id", "device", "created_at").
		From(tableName).
		Where(sq.Eq{"user_id": userID}).
		RunWith(s.db).
		QueryContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("sql error: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var res Session

		err = rows.Scan(&res.token, &res.userID, &res.ip, &res.clientID, &res.device, &res.createdAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan a row: %w", err)
		}

		sessions = append(sessions, res)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("scan error: %w", err)
	}

	return sessions, nil
}
