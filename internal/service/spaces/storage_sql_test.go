package spaces

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/sqlstorage"
)

func TestSpaceSqlstore(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tools := tools.NewMock(t)
	db := sqlstorage.NewTestStorage(t)
	store := newSqlStorage(db, tools)

	user := users.NewFakeUser(t).Build()
	space := NewFakeSpace(t).WithOwners(*user).Build()
	space2 := NewFakeSpace(t).Build()

	t.Run("Create success", func(t *testing.T) {
		// Run
		err := store.Save(ctx, space)

		// Asserts
		require.NoError(t, err)
	})

	t.Run("GetByID success", func(t *testing.T) {
		// Run
		res, err := store.GetByID(ctx, space.ID())

		// Asserts
		require.NoError(t, err)
		assert.EqualValues(t, space, res)
	})

	t.Run("GetByID not found", func(t *testing.T) {
		// Run
		res, err := store.GetByID(ctx, "some-invalid-uuid")

		// Asserts
		assert.Nil(t, res)
		require.ErrorIs(t, err, errNotFound)
	})

	t.Run("Create success", func(t *testing.T) {
		// Run
		err := store.Save(ctx, space2)

		// Asserts
		require.NoError(t, err)
	})

	t.Run("GetAllSpaces success", func(t *testing.T) {
		// Run
		res, err := store.GetAllSpaces(ctx, nil)

		// Asserts
		require.NoError(t, err)
		assert.EqualValues(t, []Space{*space, *space2}, res)
	})

	t.Run("GetAllUserSpaces with only personal success", func(t *testing.T) {
		// Run
		res, err := store.GetAllUserSpaces(ctx, user.ID(), nil)

		// Asserts
		require.NoError(t, err)
		assert.EqualValues(t, []Space{*space}, res)
	})

	t.Run("Patch success", func(t *testing.T) {
		// Run
		err := store.Patch(ctx, space.id, map[string]any{"name": "foo"})
		require.NoError(t, err)

		// Asserts
		res, err := store.GetByID(ctx, space.id)
		require.NoError(t, err)
		assert.Equal(t, "foo", res.name)
	})

	t.Run("Delete success", func(t *testing.T) {
		// Run
		err := store.Delete(ctx, space.ID())

		// Asserts
		require.NoError(t, err)

		res, err := store.GetAllSpaces(ctx, nil)
		require.NoError(t, err)
		assert.Equal(t, []Space{*space2}, res)
	})
}
