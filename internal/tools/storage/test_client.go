package storage

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/migrations"
)

func NewTestStorage(t *testing.T) *sql.DB {
	cfg := Config{Path: ":memory:"}

	client, err := NewSQliteClient(&cfg)
	require.NoError(t, err)

	err = client.Ping()
	require.NoError(t, err)

	err = migrations.Run(client, nil)
	require.NoError(t, err)

	return client
}
