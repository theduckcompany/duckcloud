package storage

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/tools"
)

func NewTestStorage(t *testing.T) *sql.DB {
	cfg := Config{Path: ":memory:"}

	if testing.Verbose() {
		cfg.Debug = true
	}

	tools := tools.NewMock(t)
	client, err := NewSQliteClient(&cfg, tools.Logger())
	require.NoError(t, err)

	err = client.Ping()
	require.NoError(t, err)

	err = RunMigrations(client, nil)
	require.NoError(t, err)

	return client
}
