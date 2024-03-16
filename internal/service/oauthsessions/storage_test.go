package oauthsessions

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/service/oauthclients"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/tools/secret"
	"github.com/theduckcompany/duckcloud/internal/tools/sqlstorage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

func TestSessionStorageStorage(t *testing.T) {
	ctx := context.Background()

	db := sqlstorage.NewTestStorage(t)
	store := newSqlStorage(db)

	user := users.NewFakeUser(t).BuildAndStore(ctx, db)
	client := oauthclients.NewFakeClient(t).CreatedBy(user).BuildAndStore(ctx, db)

	session := NewFakeSession(t).CreatedBy(user).WithClient(client).Build()
	session2 := NewFakeSession(t).CreatedBy(user).WithClient(client).Build()

	t.Run("Save success", func(t *testing.T) {
		// Run
		err := store.Save(ctx, session)

		// Asserts
		require.NoError(t, err)
	})

	t.Run("GetAllForUser success", func(t *testing.T) {
		// Run
		res, err := store.GetAllForUser(ctx, session.UserID(), nil)

		// Asserts
		require.NoError(t, err)
		assert.Equal(t, []Session{*session}, res)
	})

	t.Run("GetAllForUser with an unknown user", func(t *testing.T) {
		// Run
		res, err := store.GetAllForUser(ctx, uuid.UUID("some-invalid-id"), nil)

		// Asserts
		require.NoError(t, err)
		assert.Equal(t, []Session{}, res)
	})

	t.Run("GetByAccessToken success", func(t *testing.T) {
		// Run
		res, err := store.GetByAccessToken(ctx, session.AccessToken())

		// Asserts
		require.NoError(t, err)
		assert.Equal(t, session, res)
	})

	t.Run("GetByAccessToken with an invalid access", func(t *testing.T) {
		// Run
		res, err := store.GetByAccessToken(ctx, secret.NewText("some-invalid-token"))

		// Asserts
		assert.Nil(t, res)
		require.ErrorIs(t, err, errNotFound)
	})

	t.Run("GetByRefreshToken success", func(t *testing.T) {
		// Run
		res, err := store.GetByRefreshToken(ctx, session.RefreshToken())

		// Asserts
		require.NoError(t, err)
		assert.Equal(t, session, res)
	})

	t.Run("GetByRefreshToken with an invalid access", func(t *testing.T) {
		// Run
		res, err := store.GetByRefreshToken(ctx, secret.NewText("some-invalid-token"))

		// Asserts
		assert.Nil(t, res)
		require.ErrorIs(t, err, errNotFound)
	})

	t.Run("RemoveByAccessToken success", func(t *testing.T) {
		// Run
		err := store.RemoveByAccessToken(ctx, session.AccessToken())

		// Asserts
		require.NoError(t, err)
		res, err := store.GetByAccessToken(ctx, session.AccessToken())
		assert.Nil(t, res)
		require.ErrorIs(t, err, errNotFound)
	})

	t.Run("RemoveByAccessToken with an invalid token", func(t *testing.T) {
		// Run
		err := store.RemoveByAccessToken(ctx, secret.NewText("some-invalid-token"))

		// Asserts
		require.NoError(t, err)
	})

	t.Run("Save success 2", func(t *testing.T) {
		err := store.Save(ctx, session2)

		// Asserts
		require.NoError(t, err)
	})

	t.Run("RemoveByRefreshToken success", func(t *testing.T) {
		// Run
		err := store.RemoveByRefreshToken(ctx, session2.RefreshToken())

		// Asserts
		require.NoError(t, err)
		res, err := store.GetByRefreshToken(ctx, session2.RefreshToken())
		assert.Nil(t, res)
		require.ErrorIs(t, err, errNotFound)
	})

	t.Run("RemoveByRefreshToken with an invalid token", func(t *testing.T) {
		// Run
		err := store.RemoveByRefreshToken(ctx, secret.NewText("some-invalid-token"))

		// Asserts
		require.NoError(t, err)
	})
}
