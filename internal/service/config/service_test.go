package config

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/tools/secret"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
)

func TestConfig(t *testing.T) {
	ctx := context.Background()

	db := storage.NewTestStorage(t)
	store := newSqlStorage(db)
	svc := NewService(store)

	masterKey, err := secret.NewKey()
	require.NoError(t, err)

	key, err := secret.NewKey()
	require.NoError(t, err)

	sealedKey, err := secret.SealKey(masterKey, key)
	require.NoError(t, err)

	t.Run("SetMasterKey success", func(t *testing.T) {
		err := svc.SetMasterKey(ctx, sealedKey)
		require.NoError(t, err)
	})

	t.Run("GetMasterKey success", func(t *testing.T) {
		res, err := svc.GetMasterKey(ctx)
		require.NoError(t, err)

		assert.True(t, res.Equals(sealedKey))
	})

	t.Run("SetTotalSize success", func(t *testing.T) {
		err := svc.SetTotalSize(ctx, 4096)
		require.NoError(t, err)
	})

	t.Run("GetTotalSize success", func(t *testing.T) {
		res, err := svc.GetTotalSize(ctx)
		require.NoError(t, err)

		assert.Equal(t, uint64(4096), res)
	})
}
