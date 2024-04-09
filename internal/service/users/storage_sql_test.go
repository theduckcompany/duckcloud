package users

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/tools/sqlstorage"
)

func TestUserSqlStorage(t *testing.T) {
	ctx := context.Background()

	db := sqlstorage.NewTestStorage(t)
	store := newSqlStorage(db)

	user := NewFakeUser(t).Build()

	t.Run("GetAll with nothing", func(t *testing.T) {
		// Run
		res, err := store.GetAll(ctx, &sqlstorage.PaginateCmd{Limit: 10})

		// Asserts
		require.NoError(t, err)
		assert.Empty(t, res)
	})

	t.Run("Create success", func(t *testing.T) {
		// Run
		err := store.Save(ctx, user)

		// Asserts
		require.NoError(t, err)
	})

	t.Run("GetByID success", func(t *testing.T) {
		// Run
		res, err := store.GetByID(ctx, user.ID())

		// Asserts
		assert.NotNil(t, res)
		require.NoError(t, err)
		assert.Equal(t, user, res)
	})

	t.Run("GetByID not found", func(t *testing.T) {
		// Run
		res, err := store.GetByID(ctx, "some-invalid-id")

		// Asserts
		assert.Nil(t, res)
		require.ErrorIs(t, err, errNotFound)
	})

	t.Run("Patch success", func(t *testing.T) {
		// Restore the old username
		t.Cleanup(func() {
			err := store.Patch(ctx, user.ID(), map[string]any{"username": user.username})
			require.NoError(t, err)
		})

		// Run
		err := store.Patch(ctx, user.ID(), map[string]any{"username": "new-username"})
		require.NoError(t, err)

		// Asserts
		res, err := store.GetByID(ctx, user.ID())
		require.NoError(t, err)
		assert.Equal(t, "new-username", res.username)
	})

	t.Run("GetByUsername success", func(t *testing.T) {
		// Run
		res, err := store.GetByUsername(ctx, user.Username())

		// Asserts
		require.NoError(t, err)
		assert.Equal(t, user, res)
	})

	t.Run("GetByUsername not found", func(t *testing.T) {
		// Run
		res, err := store.GetByUsername(ctx, "some-invalid-username")

		// Asserts
		assert.Nil(t, res)
		require.ErrorIs(t, err, errNotFound)
	})

	t.Run("GetAll success", func(t *testing.T) {
		// Run
		res, err := store.GetAll(ctx, &sqlstorage.PaginateCmd{Limit: 10})

		// Asserts
		require.NoError(t, err)
		assert.Equal(t, []User{*user}, res)
	})

	t.Run("HardDelete success", func(t *testing.T) {
		// Run
		err := store.HardDelete(ctx, user.ID())
		require.NoError(t, err)

		// Asserts
		res, err := store.GetAll(ctx, nil)
		require.NoError(t, err)
		assert.Empty(t, res)
	})
}
