package oauthcodes

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/theduckcompany/duckcloud/internal/tools/ptr"
	"github.com/theduckcompany/duckcloud/internal/tools/secret"
	"github.com/theduckcompany/duckcloud/internal/tools/sqlstorage"
)

const tableName = "oauth_codes"

var errNotFound = errors.New("not found")

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
			ptr.To(sqlstorage.SQLTime(code.createdAt)),
			ptr.To(sqlstorage.SQLTime(code.expiresAt)),
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

func (t *sqlStorage) RemoveByCode(ctx context.Context, code secret.Text) error {
	_, err := sq.
		Delete(tableName).
		Where(sq.Eq{"code": code.Raw()}).
		RunWith(t.db).
		ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("sql error: %w", err)
	}

	return nil
}

func (t *sqlStorage) GetByCode(ctx context.Context, code secret.Text) (*Code, error) {
	var res Code
	var sqlCreatedAt sqlstorage.SQLTime
	var sqlExpiresAt sqlstorage.SQLTime

	err := sq.
		Select(allFields...).
		From(tableName).
		Where(sq.Eq{"code": code.Raw()}).
		RunWith(t.db).
		ScanContext(ctx,
			&res.code,
			&sqlCreatedAt,
			&sqlExpiresAt,
			&res.clientID,
			&res.userID,
			&res.redirectURI,
			&res.scope,
			&res.challenge,
			&res.challengeMethod,
		)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("sql error: %w", err)
	}

	res.expiresAt = sqlExpiresAt.Time()
	res.createdAt = sqlCreatedAt.Time()

	return &res, nil
}
