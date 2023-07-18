package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSQliteClient(t *testing.T) {
	cfg := Config{DSN: "sqlite3://" + t.TempDir() + "/db.sqlite"}

	client, err := NewSQliteClient(cfg)
	require.NoError(t, err)

	require.NoError(t, client.Ping())
}

func TestNewSQliteClientWithAnInvalidPath(t *testing.T) {
	cfg := Config{DSN: "sqlite3:///foo/some-invalidpath"}

	client, err := NewSQliteClient(cfg)
	assert.Nil(t, client)
	assert.EqualError(t, err, "unable to open database file: no such file or directory")
}
