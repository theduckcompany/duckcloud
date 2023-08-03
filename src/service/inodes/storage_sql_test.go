package inodes

import (
	"context"
	"testing"
	"time"

	"github.com/Peltoche/neurone/src/tools/storage"
	"github.com/Peltoche/neurone/src/tools/uuid"
	"github.com/stretchr/testify/assert"
)

func TestINodeSqlStorage(t *testing.T) {
	nowData := time.Now().UTC()

	dirData := INode{
		ID:             uuid.UUID("some-uuid"),
		UserID:         uuid.UUID("some-user-uuid"),
		name:           "foo",
		LastModifiedAt: nowData,
		CreatedAt:      nowData,
	}

	db := storage.NewTestStorage(t)
	storage := newSqlStorage(db)

	t.Run("Create success", func(t *testing.T) {
		err := storage.Save(context.Background(), &dirData)

		assert.NoError(t, err)
	})

	t.Run("GetByID success", func(t *testing.T) {
		res, err := storage.GetByID(context.Background(), "some-uuid")

		assert.NotNil(t, res)
		res.CreatedAt = res.CreatedAt.UTC()

		assert.NoError(t, err)
		assert.Equal(t, &dirData, res)
	})

	t.Run("GetByID not found", func(t *testing.T) {
		res, err := storage.GetByID(context.Background(), "some-invalid-id")

		assert.NoError(t, err)
		assert.Nil(t, res)
	})

	t.Run("CountUserINodes success", func(t *testing.T) {
		res, err := storage.CountUserINodes(context.Background(), uuid.UUID("some-user-uuid"))

		assert.NoError(t, err)
		assert.Equal(t, uint(1), res)
	})

	t.Run("CountUserINodes not found", func(t *testing.T) {
		res, err := storage.CountUserINodes(context.Background(), uuid.UUID("some-invalid-uuid"))

		assert.NoError(t, err)
		assert.Equal(t, uint(0), res)
	})
}
