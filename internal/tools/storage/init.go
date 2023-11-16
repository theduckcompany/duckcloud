package storage

import (
	"database/sql"
	"fmt"
)

func Init(cfg Config) (*sql.DB, error) {
	db, err := NewSQliteClient(&cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create the sqlite client: %w", err)
	}

	return db, nil
}
