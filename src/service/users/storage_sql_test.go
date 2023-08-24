package users

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/theduckcompany/duckcloud/src/tools/storage"
)

func TestUserSqlStorage(t *testing.T) {
	ctx := context.Background()

	db := storage.NewTestStorage(t)
	store := newSqlStorage(db)

	t.Run("GetAll with nothing", func(t *testing.T) {
		res, err := store.GetAll(ctx, &storage.PaginateCmd{Limit: 10})

		assert.NoError(t, err)
		assert.Len(t, res, 0)
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

		assert.NoError(t, err)
		assert.Nil(t, res)
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

		assert.NoError(t, err)
		assert.Nil(t, res)
	})

	t.Run("GetAll success", func(t *testing.T) {
		res, err := store.GetAll(ctx, &storage.PaginateCmd{Limit: 10})

		assert.NoError(t, err)
		assert.Equal(t, []User{ExampleAlice}, res)
	})
}
