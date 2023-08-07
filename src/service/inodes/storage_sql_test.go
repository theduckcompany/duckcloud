package inodes

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Peltoche/neurone/src/tools/storage"
	"github.com/Peltoche/neurone/src/tools/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestINodeSqlstore(t *testing.T) {
	nowData := time.Now().UTC()

	dirData := INode{
		ID:             uuid.UUID("some-dir-uuid"),
		UserID:         uuid.UUID("some-user-uuid"),
		Parent:         NoParent,
		name:           "foo",
		LastModifiedAt: nowData,
		CreatedAt:      nowData,
	}

	db := storage.NewTestStorage(t)
	store := newSqlStorage(db)

	t.Run("Create success", func(t *testing.T) {
		err := store.Save(context.Background(), &dirData)

		assert.NoError(t, err)
	})

	t.Run("GetByID success", func(t *testing.T) {
		res, err := store.GetByID(context.Background(), uuid.UUID("some-dir-uuid"))

		require.NoError(t, err)
		require.NotNil(t, res)
		res.CreatedAt = res.CreatedAt.UTC()
		assert.Equal(t, &dirData, res)
	})

	t.Run("Create 10 childes", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			err := store.Save(context.Background(), &INode{
				ID:             uuid.UUID(fmt.Sprintf("some-child-id-%d", i)),
				UserID:         uuid.UUID("some-user-uuid"),
				Parent:         uuid.UUID("some-dir-uuid"),
				name:           fmt.Sprintf("child-%d", i),
				LastModifiedAt: nowData,
				CreatedAt:      nowData,
			})
			require.NoError(t, err)
		}
	})

	t.Run("GetAllChildrens", func(t *testing.T) {
		res, err := store.GetAllChildrens(context.Background(), uuid.UUID("some-user-uuid"), uuid.UUID("some-dir-uuid"), &storage.PaginateCmd{
			StartAfter: map[string]string{"id": ""},
			Limit:      5,
		})

		assert.NotNil(t, res)
		for i, r := range res {
			res[i].CreatedAt = r.CreatedAt.UTC()
		}

		assert.NoError(t, err)
		assert.Len(t, res, 5)
		assert.Equal(t, res[0], INode{
			ID:             uuid.UUID("some-child-id-0"),
			UserID:         uuid.UUID("some-user-uuid"),
			Parent:         uuid.UUID("some-dir-uuid"),
			name:           "child-0",
			LastModifiedAt: nowData,
			CreatedAt:      nowData,
		}, res)
	})

	t.Run("GetAllChildrens", func(t *testing.T) {
		res, err := store.GetAllChildrens(context.Background(), uuid.UUID("some-user-uuid"), uuid.UUID("some-dir-uuid"), &storage.PaginateCmd{
			StartAfter: map[string]string{"id": "some-child-id-4"},
			Limit:      5,
		})

		assert.NotNil(t, res)
		for i, r := range res {
			res[i].CreatedAt = r.CreatedAt.UTC()
		}

		assert.NoError(t, err)
		assert.Len(t, res, 5)
		assert.Equal(t, res[0], INode{
			ID:             uuid.UUID("some-child-id-5"),
			UserID:         uuid.UUID("some-user-uuid"),
			Parent:         uuid.UUID("some-dir-uuid"),
			name:           "child-5",
			LastModifiedAt: nowData,
			CreatedAt:      nowData,
		}, res)
	})

	t.Run("GetByID not found", func(t *testing.T) {
		res, err := store.GetByID(context.Background(), "some-invalid-id")

		assert.NoError(t, err)
		assert.Nil(t, res)
	})

	t.Run("CountUserINodes success", func(t *testing.T) {
		res, err := store.CountUserINodes(context.Background(), uuid.UUID("some-user-uuid"))

		assert.NoError(t, err)
		assert.EqualValues(t, uint(11), res)
	})

	t.Run("CountUserINodes not found", func(t *testing.T) {
		res, err := store.CountUserINodes(context.Background(), uuid.UUID("some-invalid-uuid"))

		assert.NoError(t, err)
		assert.Equal(t, uint(0), res)
	})
}
