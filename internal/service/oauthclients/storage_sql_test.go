package oauthclients

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/tools/sqlstorage"
)

func TestOauthClientsSQLStorage(t *testing.T) {
	ctx := context.Background()

	db := sqlstorage.NewTestStorage(t)
	storage := newSqlStorage(db)

	// Data
	user := users.NewFakeUser(t).BuildAndStore(ctx, db)
	client := NewFakeClient(t).
		CreatedBy(user).
		Build()

	t.Run("GetByID not found", func(t *testing.T) {
		res, err := storage.GetByID(ctx, "some-invalid-id")

		assert.Nil(t, res)
		require.ErrorIs(t, err, errNotFound)
	})

	t.Run("Create", func(t *testing.T) {
		err := storage.Save(ctx, client)

		require.NoError(t, err)
	})

	t.Run("GetByID success", func(t *testing.T) {
		res, err := storage.GetByID(ctx, client.id)

		require.NoError(t, err)
		assert.EqualValues(t, client, res)
	})
}
