package oauthconsents

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"
)

const tableName = "oauth_consents"

type sqlStorage struct {
	db *sql.DB
}

func newSQLStorage(db *sql.DB) *sqlStorage {
	return &sqlStorage{db}
}

func (s *sqlStorage) Save(ctx context.Context, consent *Consent) error {
	scopes := strings.Join(consent.Scopes, ",")

	_, err := sq.
		Insert(tableName).
		Columns("id", "user_id", "client_id", "scopes", "session_token", "created_at").
		Values(consent.ID, consent.UserID, consent.ClientID, scopes, consent.SessionToken, consent.CreatedAt).
		RunWith(s.db).
		ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("sql error: %w", err)
	}

	return nil
}

func (s *sqlStorage) GetByID(ctx context.Context, id string) (*Consent, error) {
	res := Consent{}

	var rawScopes string

	err := sq.
		Select("id", "user_id", "client_id", "scopes", "session_token", "created_at").
		From(tableName).
		Where(sq.Eq{"id": id}).
		RunWith(s.db).
		ScanContext(ctx, &res.ID, &res.UserID, &res.ClientID, &rawScopes, &res.SessionToken, &res.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("sql error: %w", err)
	}

	res.Scopes = strings.Split(rawScopes, ",")

	return &res, nil
}
