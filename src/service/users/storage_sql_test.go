package users

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/theduckcompany/duckcloud/src/tools"
	"github.com/theduckcompany/duckcloud/src/tools/storage"
)

func TestUserSqlStorage(t *testing.T) {
	ctx := context.Background()

	tools := tools.NewMock(t)
	db := storage.NewTestStorage(t)
	store := newSqlStorage(db, tools)

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

	t.Run("Delete success", func(t *testing.T) {
		tools.ClockMock.On("Now").Return(time.Now()).Once()

		err := store.Delete(ctx, ExampleAlice.ID())
		assert.NoError(t, err)

		// Check that the node is no more available
		res, err := store.GetDeletedUsers(ctx, 1)
		assert.NoError(t, err)
		assert.Equal(t, res[0].ID(), ExampleAlice.ID())
	})

	t.Run("GetDeletedINodes", func(t *testing.T) {
		res, err := store.GetDeletedUsers(ctx, 10)

		assert.NoError(t, err)
		assert.Equal(t, []User{ExampleAlice}, res)
	})

	t.Run("HardDelete success", func(t *testing.T) {
		err := store.HardDelete(ctx, ExampleAlice.ID())
		assert.NoError(t, err)

		// Check that the node is no more available even as a soft deleted one
		res, err := store.GetDeletedUsers(ctx, 10)
		assert.NoError(t, err)
		assert.Len(t, res, 0)
	})
}
