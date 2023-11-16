package migrations

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/tools"
)

func newTestStorage(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", "file::memory:")
	require.NoError(t, err)

	err = db.Ping()
	require.NoError(t, err)

	return db
}

func TestRunMigration(t *testing.T) {
	tools := tools.NewMock(t)
	db := newTestStorage(t)

	err := Run(db, tools)
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
	db := newTestStorage(t)

	err := Run(db, tools)
	require.NoError(t, err)

	err = Run(db, tools)
	require.NoError(t, err)
}
