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
	"github.com/theduckcompany/duckcloud/internal/tools/sqlstorage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

func TestINodeSqlstore(t *testing.T) {
	ctx := context.Background()

	db := sqlstorage.NewTestStorage(t)
	store := newSqlStorage(db)

	var childs []INode

	// Data
	user := users.NewFakeUser(t).BuildAndStore(ctx, db)
	space := spaces.NewFakeSpace(t).WithOwners(*user).BuildAndStore(ctx, db)
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

			childs = append(childs, *inode)

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

		require.Equal(t, childs[5], res[0])
	})

	t.Run("GetSumChildsSize success", func(t *testing.T) {
		// Run
		totalSize, err := store.GetSumChildsSize(ctx, rootInode.ID())

		// Asserts
		require.NoError(t, err)
		require.Equal(t, file.Size()*10, totalSize)
	})

	t.Run("GetSumRootsSize success", func(t *testing.T) {
		// Data
		anAnotherRoot := NewFakeINode(t).IsRootDirectory().BuildAndStore(ctx, db)

		// Run
		totalSize, err := store.GetSumRootsSize(ctx) // There is two roots: rootInode and  anAnotherRoot

		// Asserts
		require.Equal(t, rootInode.size+anAnotherRoot.size, totalSize)
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
		res, err := store.GetByNameAndParent(ctx, childs[4].name, rootInode.ID())

		// Asserts
		require.NoError(t, err)
		require.EqualValues(t, childs[4], *res)
	})

	t.Run("GetByNameAndParent not matching", func(t *testing.T) {
		// Run
		res, err := store.GetByNameAndParent(ctx, "invalid-name", rootInode.ID())

		// Asserts
		require.Nil(t, res)
		require.ErrorIs(t, err, errNotFound)
	})

	t.Run("Patch success", func(t *testing.T) {
		// Data
		inode := NewFakeINode(t).Build()
		t.Cleanup(func() {
			err := store.HardDelete(ctx, inode.ID())
			require.NoError(t, err)
		})

		err := store.Save(ctx, inode)
		require.NoError(t, err)

		// Run
		err = store.Patch(ctx, inode.id, map[string]any{"name": "new-name"})
		require.NoError(t, err)

		inode.name = "new-name"

		res, err := store.GetByID(ctx, inode.ID())
		require.NoError(t, err)
		require.EqualValues(t, inode, res)
	})

	t.Run("Delete", func(t *testing.T) {
		deletedInode := NewFakeINode(t).BuildAndStore(ctx, db)

		t.Run("Delete via a Patch", func(t *testing.T) {
			// Run
			err := store.Patch(ctx, deletedInode.id, map[string]any{"deleted_at": time.Now().UTC()})

			// Asserts
			require.NoError(t, err)
		})

		t.Run("GetByID a soft deleted inode success", func(t *testing.T) {
			// Run
			res, err := store.GetByID(ctx, deletedInode.id)

			// Asserts
			require.NoError(t, err)
			require.EqualValues(t, deletedInode, res)
		})

		t.Run("GetAllDeleted", func(t *testing.T) {
			// Run
			res, err := store.GetAllDeleted(ctx, 10)

			// Asserts
			require.NoError(t, err)
			require.Equal(t, []INode{*deletedInode}, res)
		})

		t.Run("HardDelete success", func(t *testing.T) {
			// Run
			err := store.HardDelete(ctx, deletedInode.id)
			require.NoError(t, err)

			// Check that the node is no more available even as a soft deleted one
			res, err := store.GetAllDeleted(ctx, 10)
			require.NoError(t, err)
			require.Empty(t, res)
		})
	})

	t.Run("GetAllInodesWithFileID success", func(t *testing.T) {
		// Run
		res, err := store.GetAllInodesWithFileID(ctx, file.ID())
		require.NoError(t, err)
		require.Equal(t, childs, res)
	})
}
