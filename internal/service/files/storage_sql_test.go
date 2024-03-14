package files

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/tools/sqlstorage"
)

func TestUserSqlStorage(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := sqlstorage.NewTestStorage(t)
	store := newSqlStorage(db)

	// Data
	file := NewFakeFile(t).Build()

	t.Run("Save success", func(t *testing.T) {
		// Run
		err := store.Save(ctx, file)

		// Asserts
		require.NoError(t, err)
	})

	t.Run("GetByID success", func(t *testing.T) {
		// Run
		res, err := store.GetByID(ctx, file.ID())

		// Asserts
		require.NoError(t, err)
		assert.Equal(t, file, res)
	})

	t.Run("GetByID not found", func(t *testing.T) {
		// Run
		res, err := store.GetByID(ctx, "some-invalid-id")

		// Asserts
		assert.Nil(t, res)
		require.ErrorIs(t, err, errNotFound)
	})

	t.Run("Delete success", func(t *testing.T) {
		// Run
		err := store.Delete(ctx, file.ID())

		// Asserts
		require.NoError(t, err)
	})

	t.Run("GetByID a deleted file", func(t *testing.T) {
		// Run
		res, err := store.GetByID(ctx, file.ID())

		// Asserts
		assert.Nil(t, res)
		require.ErrorIs(t, err, errNotFound)
	})
}
