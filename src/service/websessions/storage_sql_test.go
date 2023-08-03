package websessions

import (
	"context"
	"testing"
	"time"

	"github.com/Peltoche/neurone/src/tools/storage"
	"github.com/Peltoche/neurone/src/tools/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSessionSqlStorage(t *testing.T) {
	nowData := time.Now().UTC()

	sessionData := Session{
		Token:     "some-token",
		UserID:    uuid.UUID("some-user-id"),
		ClientID:  "some-id",
		Device:    "IOS - Firefox",
		CreatedAt: nowData,
	}

	db := storage.NewTestStorage(t)
	storage := newSQLStorage(db)

	t.Run("Create success", func(t *testing.T) {
		err := storage.Save(context.Background(), &sessionData)

		assert.NoError(t, err)
	})

	t.Run("GetByToken success", func(t *testing.T) {
		res, err := storage.GetByToken(context.Background(), "some-token")

		require.NotNil(t, res)
		res.CreatedAt = res.CreatedAt.UTC()

		assert.NoError(t, err)
		assert.Equal(t, &sessionData, res)
	})

	t.Run("GeAllForUser success", func(t *testing.T) {
		res, err := storage.GetAllForUser(context.Background(), "some-user-id")

		require.NotNil(t, res)
		for i, r := range res {
			res[i].CreatedAt = r.CreatedAt.UTC()
		}

		assert.NoError(t, err)
		assert.Equal(t, []Session{sessionData}, res)
	})

	t.Run("GetByToken not found", func(t *testing.T) {
		res, err := storage.GetByToken(context.Background(), "some-invalid-token")

		assert.NoError(t, err)
		assert.Nil(t, res)
	})

	t.Run("RemoveByToken ", func(t *testing.T) {
		err := storage.RemoveByToken(context.Background(), "some-token")
		assert.NoError(t, err)

		res, err := storage.GetByToken(context.Background(), "some-token")
		assert.NoError(t, err)
		assert.Nil(t, res)
	})
}