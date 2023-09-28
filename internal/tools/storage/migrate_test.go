package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/tools"
)

func TestRunMigration(t *testing.T) {
	tools := tools.NewMock(t)
	db := NewTestStorage(t)

	err := RunMigrations(db, tools)
	require.NoError(t, err)

	row := db.QueryRow(`SELECT COUNT(*) FROM sqlite_schema 
  where type='table' AND name NOT LIKE 'sqlite_%'`)

	require.NoError(t, row.Err())
	var res int
	row.Scan(&res)

	// There is more than 3 tables
	assert.Greater(t, res, 3)
}

func TestRunMigrationTwice(t *testing.T) {
	tools := tools.NewMock(t)
	db := NewTestStorage(t)

	err := RunMigrations(db, tools)
	require.NoError(t, err)

	err = RunMigrations(db, tools)
	require.NoError(t, err)
}
