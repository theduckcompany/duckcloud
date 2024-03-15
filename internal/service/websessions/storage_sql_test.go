package websessions

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/tools/secret"
	"github.com/theduckcompany/duckcloud/internal/tools/sqlstorage"
)

func TestSessionSqlStorage(t *testing.T) {
	db := sqlstorage.NewTestStorage(t)
	storage := newSQLStorage(db)
	ctx := context.Background()

	user := users.NewFakeUser(t).WithAdminRole().BuildAndStore(ctx, db)
	sessionToken := "some-token"
	session := NewFakeSession(t).
		CreatedBy(user).
		WithToken(sessionToken).
		Build()

	t.Run("Create success", func(t *testing.T) {
		// Run
		err := storage.Save(context.Background(), session)

		// Asserts
		require.NoError(t, err)
	})

	t.Run("GetByToken success", func(t *testing.T) {
		// Run
		res, err := storage.GetByToken(context.Background(), secret.NewText(sessionToken))

		// Asserts
		require.NoError(t, err)
		assert.Equal(t, session, res)
	})

	t.Run("GeAllForUser success", func(t *testing.T) {
		// Run
		res, err := storage.GetAllForUser(context.Background(), user.ID(), nil)

		// Asserts
		require.NoError(t, err)
		assert.Equal(t, []Session{*session}, res)
	})

	t.Run("GetByToken not found", func(t *testing.T) {
		// Run
		res, err := storage.GetByToken(context.Background(), secret.NewText("some-invalid-token"))

		// Asserts
		assert.Nil(t, res)
		require.ErrorIs(t, err, errNotFound)
	})

	t.Run("RemoveByToken ", func(t *testing.T) {
		// Run
		err := storage.RemoveByToken(context.Background(), secret.NewText(sessionToken))

		// Asserts
		require.NoError(t, err)
		res, err := storage.GetByToken(context.Background(), secret.NewText(sessionToken))
		assert.Nil(t, res)
		require.ErrorIs(t, err, errNotFound)
	})
}
