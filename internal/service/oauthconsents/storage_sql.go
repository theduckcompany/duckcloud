package oauthconsents

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/theduckcompany/duckcloud/internal/tools/ptr"
	"github.com/theduckcompany/duckcloud/internal/tools/sqlstorage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

const tableName = "oauth_consents"

var errNotFound = errors.New("not found")

var allFields = []string{"id", "user_id", "client_id", "scopes", "session_token", "created_at"}

type sqlStorage struct {
	db sqlstorage.Querier
}

func newSQLStorage(db sqlstorage.Querier) *sqlStorage {
	return &sqlStorage{db}
}

func (s *sqlStorage) Save(ctx context.Context, consent *Consent) error {
	scopes := strings.Join(consent.Scopes(), ",")

	_, err := sq.
		Insert(tableName).
		Columns(allFields...).
		Values(consent.id, consent.userID, consent.clientID, scopes, consent.sessionToken, ptr.To(sqlstorage.SQLTime(consent.createdAt))).
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

	var sqlCreatedAt sqlstorage.SQLTime
	err := sq.
		Select(allFields...).
		From(tableName).
		Where(sq.Eq{"id": id}).
		RunWith(s.db).
		ScanContext(ctx, &res.id, &res.userID, &res.clientID, &rawScopes, &res.sessionToken, &sqlCreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("sql error: %w", err)
	}

	res.createdAt = sqlCreatedAt.Time()
	res.scopes = strings.Split(rawScopes, ",")

	return &res, nil
}

func (s *sqlStorage) GetAllForUser(ctx context.Context, userID uuid.UUID, cmd *sqlstorage.PaginateCmd) ([]Consent, error) {
	rows, err := sqlstorage.PaginateSelection(sq.
		Select(allFields...).
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

func (s *sqlStorage) Delete(ctx context.Context, consentID uuid.UUID) error {
	_, err := sq.
		Delete(tableName).
		Where(sq.Eq{"id": string(consentID)}).
		RunWith(s.db).
		ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("sql error: %w", err)
	}

	return nil
}

func (s *sqlStorage) scanRows(rows *sql.Rows) ([]Consent, error) {
	consents := []Consent{}

	for rows.Next() {
		var res Consent

		var rawScopes string
		var sqlCreatedAt sqlstorage.SQLTime

		err := rows.Scan(&res.id, &res.userID, &res.clientID, &rawScopes, &res.sessionToken, &sqlCreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan a row: %w", err)
		}

		res.scopes = strings.Split(rawScopes, ",")

		res.createdAt = sqlCreatedAt.Time()
		consents = append(consents, res)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("scan error: %w", err)
	}

	return consents, nil
}
