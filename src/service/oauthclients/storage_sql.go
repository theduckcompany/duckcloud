package oauthclients

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/theduckcompany/duckcloud/src/tools/uuid"
)

const tableName = "oauth_clients"

var allFields = []string{"id", "name", "secret", "redirect_uri", "user_id", "scopes", "is_public", "skip_validation", "created_at"}

type sqlStorage struct {
	db *sql.DB
}

func newSqlStorage(db *sql.DB) *sqlStorage {
	return &sqlStorage{db}
}

func (t *sqlStorage) Save(ctx context.Context, client *Client) error {
	_, err := sq.
		Insert(tableName).
		Columns(allFields...).
		Values(client.id,
			client.name,
			client.secret,
			client.redirectURI,
			client.userID,
			client.scopes,
			client.public,
			client.skipValidation,
			client.createdAt).
		RunWith(t.db).
		ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("sql error: %w", err)
	}

	return nil
}

func (t *sqlStorage) GetByID(ctx context.Context, id uuid.UUID) (*Client, error) {
	var res Client

	err := sq.
		Select(allFields...).
		From(tableName).
		Where(sq.Eq{"id": id}).
		RunWith(t.db).
		ScanContext(ctx, &res.id,
			&res.name,
			&res.secret,
			&res.redirectURI,
			&res.userID,
			&res.scopes,
			&res.public,
			&res.skipValidation,
			&res.createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("sql error: %w", err)
	}

	return &res, nil
}
