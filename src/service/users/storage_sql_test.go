package users

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/theduckcompany/duckcloud/src/tools/storage"
	"github.com/theduckcompany/duckcloud/src/tools/uuid"
)

func TestUserSqlStorage(t *testing.T) {
	nowData := time.Now().UTC()

	userData := User{
		id:        uuid.UUID("some-uuid"),
		username:  "some-username",
		email:     "some-email",
		fsRoot:    uuid.UUID("some-inode-uuid"),
		password:  "some-password",
		createdAt: nowData,
	}

	db := storage.NewTestStorage(t)
	storage := newSqlStorage(db)

	t.Run("Create success", func(t *testing.T) {
		err := storage.Save(context.Background(), &userData)

		assert.NoError(t, err)
	})

	t.Run("GetByID success", func(t *testing.T) {
		res, err := storage.GetByID(context.Background(), "some-uuid")

		assert.NotNil(t, res)
		res.createdAt = res.createdAt.UTC()

		assert.NoError(t, err)
		assert.Equal(t, &userData, res)
	})

	t.Run("GetByID not found", func(t *testing.T) {
		res, err := storage.GetByID(context.Background(), "some-invalid-id")

		assert.NoError(t, err)
		assert.Nil(t, res)
	})

	t.Run("GetByEmail success", func(t *testing.T) {
		res, err := storage.GetByEmail(context.Background(), "some-email")

		assert.NotNil(t, res)
		res.createdAt = res.createdAt.UTC()

		assert.NoError(t, err)
		assert.Equal(t, &userData, res)
	})

	t.Run("GetByEmail not found", func(t *testing.T) {
		res, err := storage.GetByEmail(context.Background(), "some-invalid-email")

		assert.NoError(t, err)
		assert.Nil(t, res)
	})

	t.Run("GetByUsername success", func(t *testing.T) {
		res, err := storage.GetByUsername(context.Background(), "some-username")

		assert.NotNil(t, res)
		res.createdAt = res.createdAt.UTC()

		assert.NoError(t, err)
		assert.Equal(t, &userData, res)
	})

	t.Run("GetByUsername not found", func(t *testing.T) {
		res, err := storage.GetByUsername(context.Background(), "some-invalid-username")

		assert.NoError(t, err)
		assert.Nil(t, res)
	})
}
