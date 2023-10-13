package folders

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

func TestFolderSqlstore(t *testing.T) {
	ctx := context.Background()

	// This AliceID is hardcoded in order to avoid dependency cycles
	const AliceID = uuid.UUID("86bffce3-3f53-4631-baf8-8530773884f3")

	tools := tools.NewMock(t)
	db := storage.NewTestStorage(t)
	store := newSqlStorage(db, tools)

	t.Run("Create success", func(t *testing.T) {
		err := store.Save(ctx, &ExampleAlicePersonalFolder)

		assert.NoError(t, err)
	})

	t.Run("GetByID success", func(t *testing.T) {
		res, err := store.GetByID(ctx, ExampleAlicePersonalFolder.ID())
		assert.NoError(t, err)
		assert.EqualValues(t, &ExampleAlicePersonalFolder, res)
	})

	t.Run("GetByID not found", func(t *testing.T) {
		res, err := store.GetByID(ctx, "some-invalid-uuid")

		assert.Nil(t, res)
		assert.ErrorIs(t, err, errNotFound)
	})

	t.Run("GetAllUserFolders with only personal success", func(t *testing.T) {
		res, err := store.GetAllUserFolders(ctx, AliceID, nil)
		assert.NoError(t, err)
		assert.EqualValues(t, []Folder{ExampleAlicePersonalFolder}, res)
	})

	t.Run("GetAllFoldersWithRoot success", func(t *testing.T) {
		res, err := store.GetAllFoldersWithRoot(ctx, ExampleAlicePersonalFolder.rootFS, nil)
		assert.NoError(t, err)
		assert.EqualValues(t, []Folder{ExampleAlicePersonalFolder}, res)
	})

	t.Run("GetAllFoldersWithRoot with an invalid id", func(t *testing.T) {
		res, err := store.GetAllFoldersWithRoot(ctx, "some-invalid-id", nil)
		assert.NoError(t, err)
		assert.EqualValues(t, []Folder{}, res)
	})

	t.Run("Patch success", func(t *testing.T) {
		err := store.Patch(ctx, ExampleAlicePersonalFolder.id, map[string]any{"name": "foo"})
		require.NoError(t, err)

		res, err := store.GetByID(ctx, ExampleAlicePersonalFolder.id)

		expected := ExampleAlicePersonalFolder
		expected.name = "foo"

		assert.NoError(t, err)
		assert.Equal(t, &expected, res)
	})

	t.Run("Delete success", func(t *testing.T) {
		err := store.Delete(ctx, ExampleAlicePersonalFolder.ID())
		assert.NoError(t, err)

		res, err := store.GetAllUserFolders(ctx, AliceID, nil)
		assert.NoError(t, err)
		assert.Equal(t, []Folder{}, res)
	})
}
