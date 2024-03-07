package config

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
)

const tableName = "config"

var errNotfound = errors.New("not found")

type sqlStorage struct {
	db *sql.DB
}

func newSqlStorage(db *sql.DB) *sqlStorage {
	return &sqlStorage{db}
}

func (s *sqlStorage) Save(ctx context.Context, key ConfigKey, value string) error {
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

func (s *sqlStorage) Get(ctx context.Context, key ConfigKey) (string, error) {
	var res string

	err := sq.
		Select("value").
		From(tableName).
		Where(sq.Eq{"key": key}).
		RunWith(s.db).
		ScanContext(ctx, &res)
	if errors.Is(err, sql.ErrNoRows) {
		return res, errNotfound
	}

	if err != nil {
		return res, fmt.Errorf("sql error: %w", err)
	}

	return res, nil
}
