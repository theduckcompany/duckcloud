package oauthconsents

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/theduckcompany/duckcloud/src/tools/storage"
	"github.com/theduckcompany/duckcloud/src/tools/uuid"
)

const tableName = "oauth_consents"

type sqlStorage struct {
	db *sql.DB
}

func newSQLStorage(db *sql.DB) *sqlStorage {
	return &sqlStorage{db}
}

func (s *sqlStorage) Save(ctx context.Context, consent *Consent) error {
	scopes := strings.Join(consent.Scopes(), ",")

	_, err := sq.
		Insert(tableName).
		Columns("id", "user_id", "client_id", "scopes", "session_token", "created_at").
		Values(consent.id, consent.userID, consent.clientID, scopes, consent.sessionToken, consent.createdAt).
		RunWith(s.db).
		ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("sql error: %w", err)
	}

	return nil
}

func (s *sqlStorage) GetByID(ctx context.Context, id uuid.UUID) (*Consent, error) {
	res := Consent{}

	var rawScopes string

	err := sq.
		Select("id", "user_id", "client_id", "scopes", "session_token", "created_at").
		From(tableName).
		Where(sq.Eq{"id": id}).
		RunWith(s.db).
		ScanContext(ctx, &res.id, &res.userID, &res.clientID, &rawScopes, &res.sessionToken, &res.createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("sql error: %w", err)
	}

	res.scopes = strings.Split(rawScopes, ",")

	return &res, nil
}

func (s *sqlStorage) GetAllForUser(ctx context.Context, userID uuid.UUID, cmd *storage.PaginateCmd) ([]Consent, error) {
	rows, err := storage.PaginateSelection(sq.
		Select("id", "user_id", "client_id", "scopes", "session_token", "created_at").
		Where(sq.Eq{"user_id": string(userID)}).
		From(tableName), cmd).
		RunWith(s.db).
		QueryContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("sql error: %w", err)
	}

	defer rows.Close()

	return s.scanRows(rows)
}

func (s *sqlStorage) scanRows(rows *sql.Rows) ([]Consent, error) {
	consents := []Consent{}

	for rows.Next() {
		var res Consent

		var rawScopes string

		err := rows.Scan(&res.id, &res.userID, &res.clientID, &rawScopes, &res.sessionToken, &res.createdAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan a row: %w", err)
		}

		res.scopes = strings.Split(rawScopes, ",")

		consents = append(consents, res)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("scan error: %w", err)
	}

	return consents, nil
}
