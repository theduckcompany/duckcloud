package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/src/tools"
)

func TestNewSQliteClient(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tools := tools.NewMock(t)
		cfg := Config{Path: t.TempDir() + "/db.sqlite"}

		client, err := NewSQliteClient(&cfg, tools.Logger())
		require.NoError(t, err)

		require.NoError(t, client.Ping())
	})

	t.Run("with an invalid path", func(t *testing.T) {
		tools := tools.NewMock(t)
		cfg := Config{Path: "/foo/some-invalidpath"}

		client, err := NewSQliteClient(&cfg, tools.Logger())
		assert.Nil(t, client)
		assert.EqualError(t, err, "unable to open database file: no such file or directory")
	})

	t.Run("with not specified path", func(t *testing.T) {
		tools := tools.NewMock(t)
		cfg := Config{Path: ""}

		client, err := NewSQliteClient(&cfg, tools.Logger())
		assert.NotNil(t, client)
		assert.NoError(t, err)
	})

	t.Run("with no logger", func(t *testing.T) {
		cfg := Config{Path: ""}

		client, err := NewSQliteClient(&cfg, nil)
		assert.NotNil(t, client)
		assert.NoError(t, err)
	})
}
