package stats

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/tools/sqlstorage"
)

func TestConfig(t *testing.T) {
	ctx := context.Background()

	db := storage.NewTestStorage(t)
	store := newSqlStorage(db)
	svc := newService(store)

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
