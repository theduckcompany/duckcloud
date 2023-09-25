package oauthcodes

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
)

const tableName = "oauth_codes"

var allFields = []string{"code", "created_at", "expires_at", "client_id", "user_id", "redirect_uri", "scope", "challenge", "challenge_method"}

type sqlStorage struct {
	db *sql.DB
}

func newSqlStorage(db *sql.DB) *sqlStorage {
	return &sqlStorage{db}
}

func (t *sqlStorage) Save(ctx context.Context, code *Code) error {
	_, err := sq.
		Insert(tableName).
		Columns(allFields...).
		Values(
			code.code,
			code.createdAt,
			code.expiresAt,
			code.clientID,
			code.userID,
			code.redirectURI,
			code.scope,
			code.challenge,
			code.challengeMethod,
		).
		RunWith(t.db).
		ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("sql error: %w", err)
	}

	return nil
}

func (t *sqlStorage) RemoveByCode(ctx context.Context, code string) error {
	_, err := sq.
		Delete(tableName).
		Where(sq.Eq{"code": code}).
		RunWith(t.db).
		ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("sql error: %w", err)
	}

	return nil
}

func (t *sqlStorage) GetByCode(ctx context.Context, code string) (*Code, error) {
	res := Code{}

	err := sq.
		Select(allFields...).
		From(tableName).
		Where(sq.Eq{"code": code}).
		RunWith(t.db).
		ScanContext(ctx,
			&res.code,
			&res.createdAt,
			&res.expiresAt,
			&res.clientID,
			&res.userID,
			&res.redirectURI,
			&res.scope,
			&res.challenge,
			&res.challengeMethod,
		)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("sql error: %w", err)
	}

	return &res, nil
}
