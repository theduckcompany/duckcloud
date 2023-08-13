package storage

import (
	"database/sql"
	"testing"

	"github.com/Peltoche/neurone/src/tools"
	"github.com/stretchr/testify/require"
)

func NewTestStorage(t *testing.T) *sql.DB {
	cfg := Config{Path: t.TempDir() + "/db.sqlite"}

	if testing.Verbose() {
		cfg.Debug = true
	}

	tools := tools.NewMock(t)
	err := RunMigrations(cfg, tools)
	require.NoError(t, err)

	client, err := NewSQliteClient(cfg, tools.Logger())
	require.NoError(t, err)

	err = client.Ping()
	require.NoError(t, err)

	return client
}
