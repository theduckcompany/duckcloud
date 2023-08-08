package oauthsessions

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
)

const tableName = "oauth_sessions"

type sqlStorage struct {
	db *sql.DB
}

func newSqlStorage(db *sql.DB) *sqlStorage {
	return &sqlStorage{db}
}

func (t *sqlStorage) Save(ctx context.Context, session *Session) error {
	_, err := sq.
		Insert(tableName).
		Columns(
			"access_token",
			"access_created_at",
			"access_expires_at",
			"refresh_token",
			"refresh_created_at",
			"refresh_expires_at",
			"client_id",
			"user_id",
			"scope",
		).
		Values(
			session.accessToken,
			session.accessCreatedAt,
			session.accessExpiresAt,
			session.refreshToken,
			session.refreshCreatedAt,
			session.refreshExpiresAt,
			session.clientID,
			session.userID,
			session.scope,
		).
		RunWith(t.db).
		ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("sql error: %w", err)
	}

	return nil
}

func (t *sqlStorage) RemoveByAccessToken(ctx context.Context, access string) error {
	return t.removeByKey(ctx, "access_token", access)
}

func (t *sqlStorage) RemoveByRefreshToken(ctx context.Context, refresh string) error {
	return t.removeByKey(ctx, "refresh_token", refresh)
}

func (t *sqlStorage) removeByKey(ctx context.Context, key, expected string) error {
	_, err := sq.
		Delete(tableName).
		Where(sq.Eq{key: expected}).
		RunWith(t.db).
		ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("sql error: %w", err)
	}

	return nil
}

func (t *sqlStorage) GetByAccessToken(ctx context.Context, access string) (*Session, error) {
	return t.getByKey(ctx, "access_token", access)
}

func (t *sqlStorage) GetByRefreshToken(ctx context.Context, refresh string) (*Session, error) {
	return t.getByKey(ctx, "refresh_token", refresh)
}

func (t *sqlStorage) getByKey(ctx context.Context, key, expected string) (*Session, error) {
	res := Session{}

	err := sq.
		Select(
			"access_token",
			"access_created_at",
			"access_expires_at",
			"refresh_token",
			"refresh_created_at",
			"refresh_expires_at",
			"client_id",
			"user_id",
			"scope",
		).
		From(tableName).
		Where(sq.Eq{key: expected}).
		RunWith(t.db).
		ScanContext(ctx,
			&res.accessToken,
			&res.accessCreatedAt,
			&res.accessExpiresAt,
			&res.refreshToken,
			&res.refreshCreatedAt,
			&res.refreshExpiresAt,
			&res.clientID,
			&res.userID,
			&res.scope,
		)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("sql error: %w", err)
	}

	return &res, nil
}
