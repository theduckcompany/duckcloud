package davsessions

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/tools/secret"
	"github.com/theduckcompany/duckcloud/internal/tools/sqlstorage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

func TestDavSessionSqlStorage(t *testing.T) {
	db := storage.NewTestStorage(t)
	store := newSqlStorage(db)

	t.Run("Create success", func(t *testing.T) {
		err := store.Save(context.Background(), &ExampleAliceSession)

		require.NoError(t, err)
	})

	t.Run("GetByUsernameAndPassword success", func(t *testing.T) {
		res, err := store.GetByUsernameAndPassword(context.Background(), ExampleAliceSession.username, ExampleAliceSession.password)

		require.NoError(t, err)
		assert.Equal(t, &ExampleAliceSession, res)
	})

	t.Run("GetByUsernameAndPassword not found", func(t *testing.T) {
		res, err := store.GetByUsernameAndPassword(context.Background(), "some-invalid-username", secret.NewText("some-hashed-password"))

		assert.Nil(t, res)
		require.ErrorIs(t, err, errNotFound)
	})

	t.Run("GetByID success", func(t *testing.T) {
		res, err := store.GetByID(context.Background(), ExampleAliceSession.id)

		assert.Equal(t, &ExampleAliceSession, res)
		require.NoError(t, err)
	})

	t.Run("GetByID not found", func(t *testing.T) {
		res, err := store.GetByID(context.Background(), "some-invalid-id")

		assert.Nil(t, res)
		require.ErrorIs(t, err, errNotFound)
	})

	t.Run("GetAllForUser success", func(t *testing.T) {
		res, err := store.GetAllForUser(context.Background(), ExampleAliceSession.userID, &storage.PaginateCmd{Limit: 10})

		require.NoError(t, err)
		assert.Equal(t, []DavSession{ExampleAliceSession}, res)
	})

	t.Run("GetAllForUser not found", func(t *testing.T) {
		res, err := store.GetAllForUser(context.Background(), uuid.UUID("unknown-id"), &storage.PaginateCmd{Limit: 10})

		require.NoError(t, err)
		assert.Equal(t, []DavSession{}, res)
	})

	t.Run("RemoveByID success", func(t *testing.T) {
		err := store.RemoveByID(context.Background(), ExampleAliceSession.ID())
		require.NoError(t, err)

		res, err := store.GetByID(context.Background(), ExampleAliceSession.id)
		assert.Nil(t, res)
		require.ErrorIs(t, err, errNotFound)
	})
}
