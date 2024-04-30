package sqlstorage

import (
	"database/sql"
	"fmt"
)

func Init(cfg Config) (*sql.DB, Querier, error) {
	db, err := NewSQliteClient(&cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("sqlite error: %w", err)
	}

	querier := NewSQLQuerier(db)

	return db, querier, nil
}
