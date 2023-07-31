package websessions

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
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
		Values(session.Token, session.UserID, session.IP, session.ClientID, session.Device, session.CreatedAt).
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
		ScanContext(ctx, &res.Token, &res.UserID, &res.IP, &res.ClientID, &res.Device, &res.CreatedAt)
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