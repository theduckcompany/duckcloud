package websessions

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

func TestSessionSqlStorage(t *testing.T) {
	nowData := time.Now().UTC()

	sessionData := Session{
		token:     "some-token",
		userID:    uuid.UUID("some-user-id"),
		device:    "IOS - Firefox",
		createdAt: nowData,
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
		res.createdAt = res.createdAt.UTC()

		assert.NoError(t, err)
		assert.Equal(t, &sessionData, res)
	})

	t.Run("GeAllForUser success", func(t *testing.T) {
		res, err := storage.GetAllForUser(context.Background(), "some-user-id", nil)

		require.NotNil(t, res)
		for i, r := range res {
			res[i].createdAt = r.createdAt.UTC()
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
