package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTestStorage(t *testing.T) {
	client := NewTestStorage(t)

	row := client.QueryRow(`SELECT COUNT(*) FROM sqlite_schema 
  where type='table' AND name NOT LIKE 'sqlite_%'`)

	require.NoError(t, row.Err())
	var res int
	row.Scan(&res)

	// There is more than 3 tables
	assert.Greater(t, res, 3)
}
