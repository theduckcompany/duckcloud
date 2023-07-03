package oauthclients

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
)

const tableName = "oauth_clients"

type sqlStorage struct {
	db *sql.DB
}

func newSqlStorage(db *sql.DB) *sqlStorage {
	return &sqlStorage{db}
}

func (t *sqlStorage) Save(ctx context.Context, account *Client) error {
	_, err := sq.
		Insert(tableName).
		Columns("id",
			"secret",
			"redirect_uri",
			"user_id",
			"scopes",
			"is_public",
			"skip_validation",
			"created_at").
		Values(account.ID,
			account.Secret,
			account.RedirectURI,
			account.UserID,
			account.Scopes,
			account.Public,
			account.SkipValidation,
			account.CreatedAt).
		RunWith(t.db).
		ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("sql error: %w", err)
	}

	return nil
}

func (t *sqlStorage) GetByID(ctx context.Context, id string) (*Client, error) {
	var res Client

	err := sq.
		Select("id",
			"secret",
			"redirect_uri",
			"user_id",
			"scopes",
			"is_public",
			"skip_validation",
			"created_at").
		From(tableName).
		Where(sq.Eq{"id": id}).
		RunWith(t.db).
		ScanContext(ctx, &res.ID,
			&res.Secret,
			&res.RedirectURI,
			&res.UserID,
			&res.Scopes,
			&res.Public,
			&res.SkipValidation,
			&res.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("sql error: %w", err)
	}

	return &res, nil
}
