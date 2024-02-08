package files

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
)

func TestUserSqlStorage(t *testing.T) {
	ctx := context.Background()

	db := storage.NewTestStorage(t)
	store := newSqlStorage(db)

	t.Run("Save success", func(t *testing.T) {
		err := store.Save(ctx, &ExampleFile1)

		require.NoError(t, err)
	})

	t.Run("GetByID success", func(t *testing.T) {
		res, err := store.GetByID(ctx, ExampleFile1.ID())
		require.NoError(t, err)
		assert.Equal(t, &ExampleFile1, res)
	})

	t.Run("GetByID not found", func(t *testing.T) {
		res, err := store.GetByID(ctx, "some-invalid-id")
		assert.Nil(t, res)
		require.ErrorIs(t, err, errNotFound)
	})

	t.Run("Delete success", func(t *testing.T) {
		err := store.Delete(ctx, ExampleFile1.ID())
		require.NoError(t, err)
	})

	t.Run("GetByID a deleted file", func(t *testing.T) {
		res, err := store.GetByID(ctx, ExampleFile1.ID())
		assert.Nil(t, res)
		require.ErrorIs(t, err, errNotFound)
	})
}
