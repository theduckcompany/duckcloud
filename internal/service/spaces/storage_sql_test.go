package spaces

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

func TestSpaceSqlstore(t *testing.T) {
	ctx := context.Background()

	// This AliceID is hardcoded in order to avoid dependency cycles
	const AliceID = uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3")

	tools := tools.NewMock(t)
	db := storage.NewTestStorage(t)
	store := newSqlStorage(db, tools)

	t.Run("Create success", func(t *testing.T) {
		err := store.Save(ctx, &ExampleAlicePersonalSpace)

		assert.NoError(t, err)
	})

	t.Run("GetByID success", func(t *testing.T) {
		res, err := store.GetByID(ctx, ExampleAlicePersonalSpace.ID())
		assert.NoError(t, err)
		assert.EqualValues(t, &ExampleAlicePersonalSpace, res)
	})

	t.Run("GetByID not found", func(t *testing.T) {
		res, err := store.GetByID(ctx, "some-invalid-uuid")

		assert.Nil(t, res)
		assert.ErrorIs(t, err, errNotFound)
	})

	t.Run("GetAllUserSpaces with only personal success", func(t *testing.T) {
		res, err := store.GetAllUserSpaces(ctx, AliceID, nil)
		assert.NoError(t, err)
		assert.EqualValues(t, []Space{ExampleAlicePersonalSpace}, res)
	})

	t.Run("Patch success", func(t *testing.T) {
		err := store.Patch(ctx, ExampleAlicePersonalSpace.id, map[string]any{"name": "foo"})
		require.NoError(t, err)

		res, err := store.GetByID(ctx, ExampleAlicePersonalSpace.id)

		expected := ExampleAlicePersonalSpace
		expected.name = "foo"

		assert.NoError(t, err)
		assert.Equal(t, &expected, res)
	})

	t.Run("Delete success", func(t *testing.T) {
		err := store.Delete(ctx, ExampleAlicePersonalSpace.ID())
		assert.NoError(t, err)

		res, err := store.GetAllUserSpaces(ctx, AliceID, nil)
		assert.NoError(t, err)
		assert.Equal(t, []Space{}, res)
	})
}
