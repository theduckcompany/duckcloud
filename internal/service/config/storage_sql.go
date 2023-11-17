package config

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/theduckcompany/duckcloud/internal/tools/secret"
)

const tableName = "config"

var errNotfound = errors.New("not found")

type sqlStorage struct {
	db *sql.DB
}

func newSqlStorage(db *sql.DB) *sqlStorage {
	return &sqlStorage{db}
}

func (s *sqlStorage) Save(ctx context.Context, key ConfigKey, value any) error {
	_, err := sq.
		Insert(tableName).
		Columns("key", "value").
		Values(key, value).
		Suffix("ON CONFLICT DO UPDATE SET value = ?", value).
		RunWith(s.db).
		ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("sql error: %w", err)
	}

	return nil
}

func (s *sqlStorage) GetKey(ctx context.Context, key ConfigKey) (*secret.Key, error) {
	var res secret.Key

	err := s.get(ctx, key, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (s *sqlStorage) get(ctx context.Context, key ConfigKey, val any) error {
	err := sq.
		Select("value").
		From(tableName).
		Where(sq.Eq{"key": key}).
		RunWith(s.db).
		ScanContext(ctx, val)
	if errors.Is(err, sql.ErrNoRows) {
		return errNotfound
	}

	if err != nil {
		return fmt.Errorf("sql error: %w", err)
	}

	return nil
}
