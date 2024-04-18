package sqlstorage

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/migrations"
)

func NewTestStorage(t *testing.T) Querier {
	cfg := Config{Path: ":memory:"}

	db, err := NewSQliteClient(&cfg)
	require.NoError(t, err)

	err = db.Ping()
	require.NoError(t, err)

	err = migrations.Run(db, nil)
	require.NoError(t, err)

	querier := NewSQLQuerier(db)

	return querier
}
