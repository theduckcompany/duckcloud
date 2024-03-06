package config

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
)

func TestSQLStorage(t *testing.T) {
	ctx := context.Background()

	db := storage.NewTestStorage(t)
	store := newSqlStorage(db)

	t.Run("Save success", func(t *testing.T) {
		err := store.Save(ctx, masterKey, "some-content")
		require.NoError(t, err)
	})

	t.Run("Get success", func(t *testing.T) {
		res, err := store.Get(ctx, masterKey)
		require.NoError(t, err)
		assert.Equal(t, "some-content", res)
	})
}
