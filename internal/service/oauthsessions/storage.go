package oauthsessions

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/theduckcompany/duckcloud/internal/tools/ptr"
	"github.com/theduckcompany/duckcloud/internal/tools/secret"
	"github.com/theduckcompany/duckcloud/internal/tools/sqlstorage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

const tableName = "oauth_sessions"

var errNotFound = errors.New("not found")

var allFields = []string{"access_token", "access_created_at", "access_expires_at", "refresh_token", "refresh_created_at", "refresh_expires_at", "client_id", "user_id", "scope"}

type sqlStorage struct {
	db *sql.DB
}

func newSqlStorage(db *sql.DB) *sqlStorage {
	return &sqlStorage{db}
}

func (s *sqlStorage) Save(ctx context.Context, session *Session) error {
	_, err := sq.
		Insert(tableName).
		Columns(allFields...).
		Values(
			session.accessToken,
			ptr.To(sqlstorage.SQLTime(session.accessCreatedAt)),
			ptr.To(sqlstorage.SQLTime(session.accessExpiresAt)),
			session.refreshToken,
			ptr.To(sqlstorage.SQLTime(session.refreshCreatedAt)),
			ptr.To(sqlstorage.SQLTime(session.refreshExpiresAt)),
			session.clientID,
			session.userID,
			session.scope,
		).
		RunWith(s.db).
		ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("sql error: %w", err)
	}

	return nil
}

func (s *sqlStorage) RemoveByAccessToken(ctx context.Context, access secret.Text) error {
	return s.remove(ctx, sq.Eq{"access_token": access.Raw()})
}

func (s *sqlStorage) RemoveByRefreshToken(ctx context.Context, refresh secret.Text) error {
	return s.remove(ctx, sq.Eq{"refresh_token": refresh.Raw()})
}

func (s *sqlStorage) remove(ctx context.Context, conditions ...any) error {
	query := sq.
		Delete(tableName)

	for _, condition := range conditions {
		query = query.Where(condition)
	}

	_, err := query.
		RunWith(s.db).
		ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("sql error: %w", err)
	}

	return nil
}

func (s *sqlStorage) GetByAccessToken(ctx context.Context, access secret.Text) (*Session, error) {
	return s.getWithKeys(ctx, sq.Eq{"access_token": access.Raw()})
}

func (s *sqlStorage) GetByRefreshToken(ctx context.Context, refresh secret.Text) (*Session, error) {
	return s.getWithKeys(ctx, sq.Eq{"refresh_token": refresh.Raw()})
}

func (s *sqlStorage) GetAllForUser(ctx context.Context, userID uuid.UUID, cmd *sqlstorage.PaginateCmd) ([]Session, error) {
	rows, err := sqlstorage.PaginateSelection(sq.
		Select(allFields...).
		Where(sq.Eq{"user_id": userID}).
		From(tableName), cmd).
		RunWith(s.db).
		QueryContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("sql error: %w", err)
	}
	defer rows.Close()

	return s.scanRows(rows)
}

func (s *sqlStorage) getWithKeys(ctx context.Context, conditions ...any) (*Session, error) {
	res := Session{}

	query := sq.
		Select(allFields...).
		From(tableName)

	for _, condition := range conditions {
		query = query.Where(condition)
	}

	var sqlAccessCreatedAt sqlstorage.SQLTime
	var sqlAccessExpiresAt sqlstorage.SQLTime
	var sqlRefreshCreatedAt sqlstorage.SQLTime
	var sqlRefresExpiresAt sqlstorage.SQLTime

	err := query.
		RunWith(s.db).
		ScanContext(ctx,
			&res.accessToken,
			&sqlAccessCreatedAt,
			&sqlAccessExpiresAt,
			&res.refreshToken,
			&sqlRefreshCreatedAt,
			&sqlRefresExpiresAt,
			&res.clientID,
			&res.userID,
			&res.scope,
		)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("sql error: %w", err)
	}

	res.accessCreatedAt = sqlAccessCreatedAt.Time()
	res.accessExpiresAt = sqlAccessExpiresAt.Time()
	res.refreshCreatedAt = sqlRefreshCreatedAt.Time()
	res.refreshExpiresAt = sqlRefresExpiresAt.Time()

	return &res, nil
}

func (s *sqlStorage) scanRows(rows *sql.Rows) ([]Session, error) {
	sessions := []Session{}

	for rows.Next() {
		var res Session

		var sqlAccessCreatedAt sqlstorage.SQLTime
		var sqlAccessExpiresAt sqlstorage.SQLTime
		var sqlRefreshCreatedAt sqlstorage.SQLTime
		var sqlRefresExpiresAt sqlstorage.SQLTime

		err := rows.Scan(
			&res.accessToken,
			&sqlAccessCreatedAt,
			&sqlAccessExpiresAt,
			&res.refreshToken,
			&sqlRefreshCreatedAt,
			&sqlRefresExpiresAt,
			&res.clientID,
			&res.userID,
			&res.scope,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan a row: %w", err)
		}

		res.accessCreatedAt = sqlAccessCreatedAt.Time()
		res.accessExpiresAt = sqlAccessExpiresAt.Time()
		res.refreshCreatedAt = sqlRefreshCreatedAt.Time()
		res.refreshExpiresAt = sqlRefresExpiresAt.Time()

		sessions = append(sessions, res)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("scan error: %w", err)
	}

	return sessions, nil
}
