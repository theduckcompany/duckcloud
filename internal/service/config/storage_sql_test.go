package config

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
)

func TestSQLStorage(t *testing.T) {
	ctx := context.Background()

	db := storage.NewTestStorage(t)
	store := newSqlStorage(db)

	t.Run("Save success", func(t *testing.T) {
		err := store.Save(ctx, HostName, "localhost")
		assert.NoError(t, err)
	})

	t.Run("Get success", func(t *testing.T) {
		res, err := store.Get(ctx, HostName)
		assert.NoError(t, err)
		assert.Equal(t, "localhost", res)
	})
}
