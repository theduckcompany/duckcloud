package davsessions

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/src/tools/storage"
	"github.com/theduckcompany/duckcloud/src/tools/uuid"
)

func TestDavSessionSqlStorage(t *testing.T) {
	nowData := time.Now().UTC()

	userData := DavSession{
		id:        uuid.UUID("some-uuid"),
		userID:    uuid.UUID("some-user-uuid"),
		username:  "some-username",
		password:  "some-hashed-password",
		fsRoot:    uuid.UUID("some-inode-uuid"),
		createdAt: nowData,
	}

	db := storage.NewTestStorage(t)
	storage := newSqlStorage(db)

	t.Run("Create success", func(t *testing.T) {
		err := storage.Save(context.Background(), &userData)

		assert.NoError(t, err)
	})

	t.Run("GetByUsernamePassword success", func(t *testing.T) {
		res, err := storage.GetByUsernamePassword(context.Background(), "some-username", "some-hashed-password")

		require.NotNil(t, res)
		res.createdAt = res.createdAt.UTC()

		assert.NoError(t, err)
		assert.Equal(t, &userData, res)
	})

	t.Run("GetByUsernamePassword not found", func(t *testing.T) {
		res, err := storage.GetByUsernamePassword(context.Background(), "some-invalid-username", "some-hashed-password")

		assert.NoError(t, err)
		assert.Nil(t, res)
	})
}
