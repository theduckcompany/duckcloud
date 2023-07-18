package users

import (
	"context"
	"testing"
	"time"

	"github.com/Peltoche/neurone/src/tools"
	"github.com/Peltoche/neurone/src/tools/storage"
	"github.com/Peltoche/neurone/src/tools/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserSqlStorage(t *testing.T) {
	nowData := time.Now().UTC()
	tools := tools.NewMock(t)

	userData := User{
		ID:        uuid.UUID("some-uuid"),
		Username:  "some-username",
		Email:     "some-email",
		password:  "some-password",
		CreatedAt: nowData,
	}

	db, err := storage.NewSQliteDBWithMigrate(storage.Config{Path: t.TempDir() + "/test.db"}, tools)
	require.NoError(t, err)

	storage := newSqlStorage(db)

	t.Run("Create success", func(t *testing.T) {
		err := storage.Save(context.Background(), &userData)

		assert.NoError(t, err)
	})

	t.Run("GetByID success", func(t *testing.T) {
		res, err := storage.GetByID(context.Background(), "some-uuid")

		assert.NotNil(t, res)
		res.CreatedAt = res.CreatedAt.UTC()

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
		res.CreatedAt = res.CreatedAt.UTC()

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
		res.CreatedAt = res.CreatedAt.UTC()

		assert.NoError(t, err)
		assert.Equal(t, &userData, res)
	})

	t.Run("GetByUsername not found", func(t *testing.T) {
		res, err := storage.GetByUsername(context.Background(), "some-invalid-username")

		assert.NoError(t, err)
		assert.Nil(t, res)
	})
}
