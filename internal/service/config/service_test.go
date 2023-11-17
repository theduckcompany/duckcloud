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

	key, err := secret.NewKey()
	require.NoError(t, err)

	t.Run("SetMasterKey success", func(t *testing.T) {
		err := svc.SetMasterKey(ctx, key)
		require.NoError(t, err)
	})

	t.Run("GetMasterKey success", func(t *testing.T) {
		res, err := svc.GetMasterKey(ctx)
		require.NoError(t, err)

		assert.True(t, res.Equals(key))
	})
}
