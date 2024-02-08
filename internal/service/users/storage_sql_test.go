package users

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
)

func TestUserSqlStorage(t *testing.T) {
	ctx := context.Background()

	tools := tools.NewMock(t)
	db := storage.NewTestStorage(t)
	store := newSqlStorage(db, tools)

	t.Run("GetAll with nothing", func(t *testing.T) {
		res, err := store.GetAll(ctx, &storage.PaginateCmd{Limit: 10})

		assert.NoError(t, err)
		assert.Empty(t, res)
	})

	t.Run("Create success", func(t *testing.T) {
		err := store.Save(ctx, &ExampleAlice)

		assert.NoError(t, err)
	})

	t.Run("GetByID success", func(t *testing.T) {
		res, err := store.GetByID(ctx, ExampleAlice.ID())

		assert.NotNil(t, res)
		res.createdAt = res.createdAt.UTC()

		assert.NoError(t, err)
		assert.Equal(t, &ExampleAlice, res)
	})

	t.Run("GetByID not found", func(t *testing.T) {
		res, err := store.GetByID(ctx, "some-invalid-id")

		assert.Nil(t, res)
		assert.ErrorIs(t, err, errNotFound)
	})

	t.Run("Patch success", func(t *testing.T) {
		// Restore the old username
		t.Cleanup(func() {
			err := store.Patch(ctx, ExampleAlice.ID(), map[string]any{"username": ExampleAlice.username})
			assert.NoError(t, err)
		})

		err := store.Patch(ctx, ExampleAlice.ID(), map[string]any{"username": "new-username"})
		assert.NoError(t, err)

		res, err := store.GetByID(ctx, ExampleAlice.ID())

		aliceWithNewUsername := ExampleAlice
		aliceWithNewUsername.username = "new-username"

		assert.Equal(t, &aliceWithNewUsername, res)
	})

	t.Run("GetByUsername success", func(t *testing.T) {
		res, err := store.GetByUsername(ctx, ExampleAlice.Username())

		assert.NotNil(t, res)
		res.createdAt = res.createdAt.UTC()

		assert.NoError(t, err)
		assert.Equal(t, &ExampleAlice, res)
	})

	t.Run("GetByUsername not found", func(t *testing.T) {
		res, err := store.GetByUsername(ctx, "some-invalid-username")

		assert.Nil(t, res)
		assert.ErrorIs(t, err, errNotFound)
	})

	t.Run("GetAll success", func(t *testing.T) {
		res, err := store.GetAll(ctx, &storage.PaginateCmd{Limit: 10})

		assert.NoError(t, err)
		assert.Equal(t, []User{ExampleAlice}, res)
	})

	t.Run("HardDelete success", func(t *testing.T) {
		err := store.HardDelete(ctx, ExampleAlice.ID())
		assert.NoError(t, err)

		// Check that the node is no more available even as a soft deleted one
		res, err := store.GetAll(ctx, nil)
		assert.NoError(t, err)
		assert.Empty(t, res)
	})
}
