package inodes

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/service/spaces"
	"github.com/theduckcompany/duckcloud/internal/tools"
	"github.com/theduckcompany/duckcloud/internal/tools/ptr"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

func TestINodeSqlstore(t *testing.T) {
	nowData := time.Now().UTC()
	ctx := context.Background()

	tools := tools.NewMock(t)
	db := storage.NewTestStorage(t)
	store := newSqlStorage(db, tools)

	t.Run("Create success", func(t *testing.T) {
		err := store.Save(ctx, &ExampleBobRoot)

		assert.NoError(t, err)
	})

	t.Run("GetByID success", func(t *testing.T) {
		res, err := store.GetByID(ctx, ExampleBobRoot.SpaceID(), ExampleBobRoot.ID())

		require.NoError(t, err)
		require.NotNil(t, res)
		assert.Equal(t, &ExampleBobRoot, res)
	})

	t.Run("GetSpaceRoot success", func(t *testing.T) {
		res, err := store.GetSpaceRoot(ctx, ExampleBobRoot.spaceID)
		assert.NoError(t, err)
		assert.Equal(t, &ExampleBobRoot, res)
	})

	t.Run("GetSpaceRoot with an unknown space", func(t *testing.T) {
		res, err := store.GetSpaceRoot(ctx, uuid.UUID("some-invalid-space"))
		assert.ErrorIs(t, err, errNotFound)
		assert.Nil(t, res)
	})

	t.Run("Create 10 childes", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			err := store.Save(ctx, &INode{
				id:             uuid.UUID(fmt.Sprintf("some-child-id-%d", i)),
				parent:         ptr.To(ExampleBobRoot.ID()),
				size:           10,
				name:           fmt.Sprintf("child-%d", i),
				spaceID:        ExampleBobRoot.spaceID,
				lastModifiedAt: nowData,
				createdAt:      nowData,
				fileID:         nil,
			})
			require.NoError(t, err)
		}
	})

	t.Run("GetAllChildrens success", func(t *testing.T) {
		res, err := store.GetAllChildrens(ctx, ExampleBobRoot.ID(), &storage.PaginateCmd{
			StartAfter: map[string]string{"id": "some-child-id-4"},
			Limit:      5,
		})

		assert.NotNil(t, res)
		for i, r := range res {
			res[i].createdAt = r.createdAt.UTC()
		}

		assert.NoError(t, err)
		assert.Len(t, res, 5)
		assert.Equal(t, res[0], INode{
			id:             uuid.UUID("some-child-id-5"),
			parent:         ptr.To(ExampleBobRoot.ID()),
			name:           "child-5",
			spaceID:        ExampleBobRoot.spaceID,
			lastModifiedAt: nowData,
			size:           10,
			createdAt:      nowData,
			fileID:         nil,
		}, res)
	})

	t.Run("GetSumChildsSize success", func(t *testing.T) {
		totalSize, err := store.GetSumChildsSize(ctx, ExampleBobRoot.ID())
		assert.NoError(t, err)
		// 10 bytes x 10 files ==  100 bytes
		assert.Equal(t, uint64(100), totalSize)
	})

	t.Run("GetSumChildsSize with an invalid space", func(t *testing.T) {
		totalSize, err := store.GetSumChildsSize(ctx, uuid.UUID("some-invalid-id"))
		assert.Equal(t, uint64(0), totalSize)
		assert.NoError(t, err)
	})

	t.Run("GetByID not found", func(t *testing.T) {
		res, err := store.GetByID(ctx, spaces.ExampleBobPersonalSpace.ID(), "some-invalid-id")

		assert.Nil(t, res)
		assert.ErrorIs(t, err, errNotFound)
	})

	t.Run("GetByNameAndParent success", func(t *testing.T) {
		res, err := store.GetByNameAndParent(ctx, "child-5", ExampleBobRoot.ID())
		assert.NoError(t, err)
		assert.EqualValues(t, &INode{
			id:             uuid.UUID("some-child-id-5"),
			parent:         ptr.To(ExampleBobRoot.ID()),
			name:           "child-5",
			spaceID:        ExampleBobRoot.spaceID,
			lastModifiedAt: nowData,
			size:           10,
			createdAt:      nowData,
			fileID:         nil,
		}, res)
	})

	t.Run("GetByNameAndParent not matching", func(t *testing.T) {
		res, err := store.GetByNameAndParent(ctx, "invalid-name", ExampleBobRoot.ID())

		assert.Nil(t, res)
		assert.ErrorIs(t, err, errNotFound)
	})

	t.Run("Patch success", func(t *testing.T) {
		err := store.Save(ctx, &ExampleAliceFile)
		require.NoError(t, err)

		modifiedINode := ExampleAliceFile
		modifiedINode.name = "new-name"

		err = store.Patch(ctx, ExampleAliceFile.id, map[string]any{"name": "new-name"})
		assert.NoError(t, err)

		res, err := store.GetByID(ctx, spaces.ExampleBobPersonalSpace.ID(), ExampleAliceFile.ID())
		assert.NoError(t, err)
		assert.EqualValues(t, &modifiedINode, res)
	})

	t.Run("Delete via a Patch", func(t *testing.T) {
		err := store.Patch(ctx, uuid.UUID("some-child-id-5"), map[string]any{"deleted_at": time.Now().UTC()})
		assert.NoError(t, err)
	})

	t.Run("GetByID a soft deleted inode success", func(t *testing.T) {
		res, err := store.GetByID(ctx, spaces.ExampleBobPersonalSpace.ID(), uuid.UUID("some-child-id-5"))

		require.NoError(t, err)
		require.NotNil(t, res)
		assert.EqualValues(t, &INode{
			id:             uuid.UUID("some-child-id-5"),
			parent:         ptr.To(ExampleBobRoot.ID()),
			name:           "child-5",
			spaceID:        ExampleBobRoot.spaceID,
			size:           10,
			lastModifiedAt: nowData,
			createdAt:      nowData,
			fileID:         nil,
		}, res)
	})

	t.Run("GetAllDeleted", func(t *testing.T) {
		res, err := store.GetAllDeleted(ctx, 10)

		assert.NoError(t, err)
		assert.Len(t, res, 1)
		assert.Equal(t, INode{
			id:             uuid.UUID("some-child-id-5"),
			parent:         ptr.To(ExampleBobRoot.ID()),
			name:           "child-5",
			spaceID:        ExampleBobRoot.spaceID,
			size:           10,
			lastModifiedAt: nowData,
			createdAt:      nowData,
			fileID:         nil,
		}, res[0])
	})

	t.Run("HardDelete success", func(t *testing.T) {
		err := store.HardDelete(ctx, uuid.UUID("some-child-id-5"))
		assert.NoError(t, err)

		// Check that the node is no more available even as a soft deleted one
		res, err := store.GetAllDeleted(ctx, 10)
		assert.NoError(t, err)
		assert.Len(t, res, 0)
	})

	t.Run("GetAllInodesWithFileID success", func(t *testing.T) {
		err := store.Save(ctx, &ExampleAliceFile2)
		assert.NoError(t, err)

		res, err := store.GetAllInodesWithFileID(ctx, *ExampleAliceFile2.fileID)
		assert.NoError(t, err)
		assert.Equal(t, []INode{ExampleAliceFile2}, res)
	})
}
