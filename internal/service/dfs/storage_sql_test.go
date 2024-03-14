package dfs

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/service/files"
	"github.com/theduckcompany/duckcloud/internal/service/spaces"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/tools/ptr"
	"github.com/theduckcompany/duckcloud/internal/tools/sqlstorage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

func TestINodeSqlstore(t *testing.T) {
	ctx := context.Background()

	db := sqlstorage.NewTestStorage(t)
	store := newSqlStorage(db)

	// Data
	user := users.NewFakeUser(t).BuildAndStore(db)
	space := spaces.NewFakeSpace(t).WithOwners(*user).BuildAndStore(db)
	file := files.NewFakeFile(t).Build()
	now := time.Now().UTC()

	rootInode := NewFakeINode(t).
		WithSpace(space).
		IsRootDirectory().
		CreatedBy(user).
		CreatedAt(now).
		Build()

	t.Run("Create success", func(t *testing.T) {
		// Run
		err := store.Save(ctx, rootInode)

		// Asserts
		require.NoError(t, err)
	})

	t.Run("GetByID success", func(t *testing.T) {
		// Run
		res, err := store.GetByID(ctx, rootInode.ID())

		// Asserts
		require.NoError(t, err)
		require.NotNil(t, res)
		require.Equal(t, rootInode, res)
	})

	t.Run("GetSpaceRoot success", func(t *testing.T) {
		// Run
		res, err := store.GetSpaceRoot(ctx, space.ID())

		// Asserts
		require.NoError(t, err)
		require.Equal(t, rootInode, res)
	})

	t.Run("GetSpaceRoot with an unknown space", func(t *testing.T) {
		// Run
		res, err := store.GetSpaceRoot(ctx, uuid.UUID("some-invalid-space"))

		// Asserts
		require.ErrorIs(t, err, errNotFound)
		require.Nil(t, res)
	})

	t.Run("Create 10 childes", func(t *testing.T) {
		// Run
		for i := 0; i < 10; i++ {
			inode := NewFakeINode(t).
				WithSpace(space).
				WithFile(file).
				WithParent(rootInode).
				WithID(fmt.Sprintf("some-child-id-%d", i)).
				WithName(fmt.Sprintf("child-%d", i)).
				CreatedBy(user).
				CreatedAt(now).
				Build()

			err := store.Save(ctx, inode)
			require.NoError(t, err)
		}
	})

	t.Run("GetAllChildrens success", func(t *testing.T) {
		// Run
		res, err := store.GetAllChildrens(ctx, rootInode.ID(), &sqlstorage.PaginateCmd{
			StartAfter: map[string]string{"id": "some-child-id-4"},
			Limit:      5,
		})

		// Asserts
		require.NotNil(t, res)
		require.NoError(t, err)
		require.Len(t, res, 5)

		expected := NewFakeINode(t).
			WithSpace(space).
			WithFile(file).
			WithID("some-child-id-5").
			WithParent(rootInode).
			WithName("child-5").
			CreatedBy(user).
			CreatedAt(now).
			Build()

		require.Equal(t, *expected, res[0])
	})

	t.Run("GetSumChildsSize success", func(t *testing.T) {
		// Run
		totalSize, err := store.GetSumChildsSize(ctx, rootInode.ID())

		// Asserts
		require.NoError(t, err)
		require.Equal(t, uint64(file.Size()*10), totalSize)
	})

	t.Run("GetSumRootsSize success", func(t *testing.T) {
		// Data
		anAnotherRoot := NewFakeINode(t).IsRootDirectory().BuildAndStore(db)

		// Run
		totalSize, err := store.GetSumRootsSize(ctx) // There is two roots: rootInode and  anAnotherRoot

		// Asserts
		require.Equal(t, uint64(rootInode.size+anAnotherRoot.size), totalSize)
		require.NoError(t, err)
	})

	t.Run("GetSumChildsSize with an invalid space", func(t *testing.T) {
		// Run
		totalSize, err := store.GetSumChildsSize(ctx, uuid.UUID("some-invalid-id"))

		// Asserts
		require.Equal(t, uint64(0), totalSize)
		require.NoError(t, err)
	})

	t.Run("GetByID not found", func(t *testing.T) {
		// Run
		res, err := store.GetByID(ctx, "some-invalid-id")

		// Asserts
		require.Nil(t, res)
		require.ErrorIs(t, err, errNotFound)
	})

	t.Run("GetByNameAndParent success", func(t *testing.T) {
		// Run
		res, err := store.GetByNameAndParent(ctx, "child-5", ExampleBobRoot.ID())
		require.NoError(t, err)
		require.EqualValues(t, &INode{
			id:             uuid.UUID("some-child-id-5"),
			parent:         ptr.To(ExampleBobRoot.ID()),
			name:           "child-5",
			lastModifiedAt: nowData,
			size:           10,
			createdAt:      nowData,
			fileID:         nil,
		}, res)
	})
	//
	// t.Run("GetByNameAndParent not matching", func(t *testing.T) {
	// 	res, err := store.GetByNameAndParent(ctx, "invalid-name", ExampleBobRoot.ID())
	//
	// 	require.Nil(t, res)
	// 	require.ErrorIs(t, err, errNotFound)
	// })
	//
	// t.Run("Patch success", func(t *testing.T) {
	// 	err := store.Save(ctx, &ExampleAliceFile)
	// 	require.NoError(t, err)
	//
	// 	modifiedINode := ExampleAliceFile
	// 	modifiedINode.name = "new-name"
	//
	// 	err = store.Patch(ctx, ExampleAliceFile.id, map[string]any{"name": "new-name"})
	// 	require.NoError(t, err)
	//
	// 	res, err := store.GetByID(ctx, ExampleAliceFile.ID())
	// 	require.NoError(t, err)
	// 	require.EqualValues(t, &modifiedINode, res)
	// })
	//
	// t.Run("Delete via a Patch", func(t *testing.T) {
	// 	err := store.Patch(ctx, uuid.UUID("some-child-id-5"), map[string]any{"deleted_at": time.Now().UTC()})
	// 	require.NoError(t, err)
	// })
	//
	// t.Run("GetByID a soft deleted inode success", func(t *testing.T) {
	// 	res, err := store.GetByID(ctx, uuid.UUID("some-child-id-5"))
	//
	// 	require.NoError(t, err)
	// 	require.NotNil(t, res)
	// 	require.EqualValues(t, &INode{
	// 		id:             uuid.UUID("some-child-id-5"),
	// 		parent:         ptr.To(ExampleBobRoot.ID()),
	// 		name:           "child-5",
	// 		size:           10,
	// 		lastModifiedAt: nowData,
	// 		createdAt:      nowData,
	// 		fileID:         nil,
	// 	}, res)
	// })
	//
	// t.Run("GetAllDeleted", func(t *testing.T) {
	// 	res, err := store.GetAllDeleted(ctx, 10)
	//
	// 	require.NoError(t, err)
	// 	require.Len(t, res, 1)
	// 	require.Equal(t, INode{
	// 		id:             uuid.UUID("some-child-id-5"),
	// 		parent:         ptr.To(ExampleBobRoot.ID()),
	// 		name:           "child-5",
	// 		size:           10,
	// 		lastModifiedAt: nowData,
	// 		createdAt:      nowData,
	// 		fileID:         nil,
	// 	}, res[0])
	// })
	//
	// t.Run("HardDelete success", func(t *testing.T) {
	// 	err := store.HardDelete(ctx, uuid.UUID("some-child-id-5"))
	// 	require.NoError(t, err)
	//
	// 	// Check that the node is no more available even as a soft deleted one
	// 	res, err := store.GetAllDeleted(ctx, 10)
	// 	require.NoError(t, err)
	// 	require.Empty(t, res)
	// })
	//
	// t.Run("GetAllInodesWithFileID success", func(t *testing.T) {
	// 	err := store.Save(ctx, &ExampleAliceFile2)
	// 	require.NoError(t, err)
	//
	// 	res, err := store.GetAllInodesWithFileID(ctx, *ExampleAliceFile2.fileID)
	// 	require.NoError(t, err)
	// 	require.Equal(t, []INode{ExampleAliceFile2}, res)
	// })
}
