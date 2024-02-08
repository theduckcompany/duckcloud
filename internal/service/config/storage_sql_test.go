package config

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/tools/secret"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
)

func TestSQLStorage(t *testing.T) {
	ctx := context.Background()

	db := storage.NewTestStorage(t)
	store := newSqlStorage(db)

	mk, err := secret.NewKey()
	require.NoError(t, err)

	key, err := secret.NewKey()
	require.NoError(t, err)

	sealedKey, err := secret.SealKey(mk, key)
	require.NoError(t, err)

	t.Run("Save success", func(t *testing.T) {
		err := store.Save(ctx, masterKey, sealedKey)
		require.NoError(t, err)
	})

	t.Run("Get success", func(t *testing.T) {
		res, err := store.GetKey(ctx, masterKey)
		require.NoError(t, err)
		assert.True(t, sealedKey.Equals(res))
	})
}
