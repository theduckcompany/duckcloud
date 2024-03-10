package sqlstorage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSQliteClient(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		cfg := Config{Path: t.TempDir() + "/db.sqlite"}

		client, err := NewSQliteClient(&cfg)
		require.NoError(t, err)

		require.NoError(t, client.Ping())
	})

	t.Run("with an invalid path", func(t *testing.T) {
		cfg := Config{Path: "/foo/some-invalidpath"}

		client, err := NewSQliteClient(&cfg)
		assert.Nil(t, client)
		require.EqualError(t, err, "unable to open database file: no such file or directory")
	})

	t.Run("with not specified path", func(t *testing.T) {
		cfg := Config{Path: ""}

		client, err := NewSQliteClient(&cfg)
		assert.NotNil(t, client)
		require.NoError(t, err)
	})
}
