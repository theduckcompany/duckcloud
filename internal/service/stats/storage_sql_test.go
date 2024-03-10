package stats

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/tools/sqlstorage"
)

func TestSQLStorage(t *testing.T) {
	ctx := context.Background()

	db := storage.NewTestStorage(t)
	store := newSqlStorage(db)

	t.Run("Save success", func(t *testing.T) {
		err := store.Save(ctx, totalSizeKey, uint64(4096))
		require.NoError(t, err)
	})

	t.Run("Get success", func(t *testing.T) {
		var res uint64

		err := store.Get(ctx, totalSizeKey, &res)
		require.NoError(t, err)
		assert.Equal(t, uint64(4096), res)
	})
}
