package inodes

import (
	"context"
	"fmt"
	"io/fs"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/src/tools"
	"github.com/theduckcompany/duckcloud/src/tools/ptr"
	"github.com/theduckcompany/duckcloud/src/tools/storage"
	"github.com/theduckcompany/duckcloud/src/tools/uuid"
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
		res, err := store.GetByID(ctx, ExampleBobRoot.ID())

		require.NoError(t, err)
		require.NotNil(t, res)
		assert.Equal(t, &ExampleBobRoot, res)
	})

	t.Run("Create 10 childes", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			err := store.Save(ctx, &INode{
				id:             uuid.UUID(fmt.Sprintf("some-child-id-%d", i)),
				userID:         ExampleBobRoot.UserID(),
				parent:         ptr.To(ExampleBobRoot.ID()),
				name:           fmt.Sprintf("child-%d", i),
				mode:           0o660 | fs.ModeDir,
				lastModifiedAt: nowData,
				createdAt:      nowData,
			})
			require.NoError(t, err)
		}
	})

	t.Run("GetAllChildrens success", func(t *testing.T) {
		res, err := store.GetAllChildrens(ctx, ExampleBobRoot.UserID(), ExampleBobRoot.ID(), &storage.PaginateCmd{
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
			userID:         ExampleBobRoot.UserID(),
			parent:         ptr.To(ExampleBobRoot.ID()),
			name:           "child-5",
			mode:           0o660 | fs.ModeDir,
			lastModifiedAt: nowData,
			createdAt:      nowData,
		}, res)
	})

	t.Run("GetByID not found", func(t *testing.T) {
		res, err := store.GetByID(ctx, "some-invalid-id")

		assert.NoError(t, err)
		assert.Nil(t, res)
	})

	t.Run("GetByNameAndParent success", func(t *testing.T) {
		res, err := store.GetByNameAndParent(ctx, ExampleBobRoot.UserID(), "child-5", ExampleBobRoot.ID())
		assert.NoError(t, err)
		assert.EqualValues(t, &INode{
			id:             uuid.UUID("some-child-id-5"),
			userID:         ExampleBobRoot.UserID(),
			parent:         ptr.To(ExampleBobRoot.ID()),
			name:           "child-5",
			mode:           0o660 | fs.ModeDir,
			lastModifiedAt: nowData,
			createdAt:      nowData,
		}, res)
	})

	t.Run("GetByNameAndParent not matching", func(t *testing.T) {
		res, err := store.GetByNameAndParent(ctx, ExampleBobRoot.UserID(), "invalid-name", ExampleBobRoot.ID())
		assert.NoError(t, err)
		assert.Nil(t, res)
	})

	t.Run("CountUserINodes success", func(t *testing.T) {
		res, err := store.CountUserINodes(ctx, ExampleBobRoot.UserID())

		assert.NoError(t, err)
		// the 10 files + ExampleAliceFile
		assert.EqualValues(t, uint(11), res)
	})

	t.Run("CountUserINodes not found", func(t *testing.T) {
		res, err := store.CountUserINodes(ctx, uuid.UUID("some-invalid-uuid"))

		assert.NoError(t, err)
		assert.Equal(t, uint(0), res)
	})

	t.Run("UpdateModifiedSizeAndChecksum success", func(t *testing.T) {
		nowUpdated := time.Now().UTC()

		err := store.Save(ctx, &ExampleAliceFile)
		require.NoError(t, err)

		modifiedINode := ExampleAliceFile
		modifiedINode.checksum = "some-new-checksum"
		modifiedINode.lastModifiedAt = nowUpdated
		modifiedINode.size = 84

		err = store.UpdateModifiedSizeAndChecksum(ctx, &modifiedINode)
		assert.NoError(t, err)

		res, err := store.GetByID(ctx, ExampleAliceFile.ID())
		assert.NoError(t, err)
		assert.EqualValues(t, &modifiedINode, res)
	})

	t.Run("Delete success", func(t *testing.T) {
		tools.ClockMock.On("Now").Return(time.Now()).Once()

		err := store.Delete(ctx, uuid.UUID("some-child-id-5"))
		assert.NoError(t, err)

		// Check that the node is no more available
		res, err := store.GetAllDeleted(ctx, 1)
		assert.NoError(t, err)
		assert.Equal(t, res[0].ID(), uuid.UUID("some-child-id-5"))
	})

	t.Run("GetAllDeleted", func(t *testing.T) {
		res, err := store.GetAllDeleted(ctx, 10)

		assert.NoError(t, err)
		assert.Len(t, res, 1)
		assert.Equal(t, INode{
			id:             uuid.UUID("some-child-id-5"),
			userID:         ExampleBobRoot.UserID(),
			parent:         ptr.To(ExampleBobRoot.ID()),
			name:           "child-5",
			mode:           0o660 | fs.ModeDir,
			lastModifiedAt: nowData,
			createdAt:      nowData,
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
}
