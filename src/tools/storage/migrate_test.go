package storage

import (
	"testing"

	"github.com/Peltoche/neurone/src/tools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunMigration(t *testing.T) {
	cfg := Config{DSN: "sqlite3://" + t.TempDir() + "/db.sqlite"}

	tools := tools.NewMock(t)
	err := RunMigrations(cfg, tools)
	require.NoError(t, err)

	client, err := NewSQliteClient(cfg)
	require.NoError(t, err)

	row := client.QueryRow(`SELECT COUNT(*) FROM sqlite_schema 
  where type='table' AND name NOT LIKE 'sqlite_%'`)

	require.NoError(t, row.Err())
	var res int
	row.Scan(&res)

	// There is more than 3 tables
	assert.Greater(t, res, 3)
}

func TestRunMigrationWithAnInvalidDSN(t *testing.T) {
	cfg := Config{DSN: "sqlite3:///foo/some-invali-path"}

	tools := tools.NewMock(t)
	err := RunMigrations(cfg, tools)
	require.EqualError(t, err, "failed to create a migrate manager: unable to open database file: no such file or directory")
}

func TestRunMigrationTwice(t *testing.T) {
	cfg := Config{DSN: "sqlite3://" + t.TempDir() + "/db.sqlite"}

	tools := tools.NewMock(t)
	err := RunMigrations(cfg, tools)
	require.NoError(t, err)

	err = RunMigrations(cfg, tools)
	require.NoError(t, err)
}
