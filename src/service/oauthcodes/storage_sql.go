package oauthcodes

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
)

const tableName = "oauth_codes"

type sqlStorage struct {
	db *sql.DB
}

func newSqlStorage(db *sql.DB) *sqlStorage {
	return &sqlStorage{db}
}

func (t *sqlStorage) Save(ctx context.Context, code *Code) error {
	_, err := sq.
		Insert(tableName).
		Columns(
			"code",
			"created_at",
			"expires_at",
			"client_id",
			"user_id",
			"redirect_uri",
			"scope",
			"challenge",
			"challenge_method",
		).
		Values(
			code.Code,
			code.CreatedAt,
			code.ExpiresAt,
			code.ClientID,
			code.UserID,
			code.RedirectURI,
			code.Scope,
			code.Challenge,
			code.ChallengeMethod,
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
		Select(
			"code",
			"created_at",
			"expires_at",
			"client_id",
			"user_id",
			"redirect_uri",
			"scope",
			"challenge",
			"challenge_method",
		).
		From(tableName).
		Where(sq.Eq{"code": code}).
		RunWith(t.db).
		ScanContext(ctx,
			&res.Code,
			&res.CreatedAt,
			&res.ExpiresAt,
			&res.ClientID,
			&res.UserID,
			&res.RedirectURI,
			&res.Scope,
			&res.Challenge,
			&res.ChallengeMethod,
		)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("sql error: %w", err)
	}

	return &res, nil
}
