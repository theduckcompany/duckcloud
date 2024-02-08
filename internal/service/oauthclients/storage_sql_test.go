package oauthclients

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
)

func TestOauthClientsSQLStorage(t *testing.T) {
	ctx := context.Background()

	now := time.Now().UTC()
	clientExample := Client{
		id:             "some-client-id",
		secret:         "some-secret",
		redirectURI:    "some-url",
		userID:         "some-user-id",
		createdAt:      now,
		scopes:         []string{"scope-a"},
		public:         true,
		skipValidation: true,
	}

	db := storage.NewTestStorage(t)

	storage := newSqlStorage(db)

	t.Run("GetByID not found", func(t *testing.T) {
		res, err := storage.GetByID(ctx, "some-invalid-id")

		assert.Nil(t, res)
		require.ErrorIs(t, err, errNotFound)
	})

	t.Run("Create", func(t *testing.T) {
		err := storage.Save(context.Background(), &clientExample)

		require.NoError(t, err)
	})

	t.Run("GetByID success", func(t *testing.T) {
		res, err := storage.GetByID(ctx, "some-client-id")

		require.NoError(t, err)
		assert.EqualValues(t, &clientExample, res)
	})
}
