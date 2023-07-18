package storage

import (
	"database/sql"
	"testing"

	"github.com/Peltoche/neurone/src/tools"
	"github.com/stretchr/testify/require"
)

func NewTestStorage(t *testing.T) *sql.DB {
	cfg := Config{DSN: "sqlite3://" + t.TempDir() + "/db.sqlite"}

	tools := tools.NewMock(t)
	err := RunMigrations(cfg, tools)
	require.NoError(t, err)

	client, err := NewSQliteClient(cfg)
	require.NoError(t, err)

	err = client.Ping()
	require.NoError(t, err)

	return client
}
