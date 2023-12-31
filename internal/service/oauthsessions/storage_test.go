package oauthsessions

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/theduckcompany/duckcloud/internal/tools/secret"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

func TestSessionStorageStorage(t *testing.T) {
	ctx := context.Background()
	db := storage.NewTestStorage(t)
	store := newSqlStorage(db)

	t.Run("Save success", func(t *testing.T) {
		err := store.Save(ctx, &ExampleAliceSession)
		assert.NoError(t, err)
	})

	t.Run("GetAllForUser success", func(t *testing.T) {
		res, err := store.GetAllForUser(ctx, ExampleAliceSession.UserID(), nil)

		assert.NoError(t, err)
		assert.Equal(t, []Session{ExampleAliceSession}, res)
	})

	t.Run("GetAllForUser with an unknown user", func(t *testing.T) {
		res, err := store.GetAllForUser(ctx, uuid.UUID("some-invalid-id"), nil)

		assert.NoError(t, err)
		assert.Equal(t, []Session{}, res)
	})

	t.Run("GetByAccessToken success", func(t *testing.T) {
		res, err := store.GetByAccessToken(ctx, ExampleAliceSession.AccessToken())

		assert.NoError(t, err)
		assert.Equal(t, &ExampleAliceSession, res)
	})

	t.Run("GetByAccessToken with an invalid access", func(t *testing.T) {
		res, err := store.GetByAccessToken(ctx, secret.NewText("some-invalid-token"))

		assert.Nil(t, res)
		assert.ErrorIs(t, err, errNotFound)
	})

	t.Run("GetByRefreshToken success", func(t *testing.T) {
		res, err := store.GetByRefreshToken(ctx, ExampleAliceSession.RefreshToken())
		assert.NoError(t, err)
		assert.Equal(t, &ExampleAliceSession, res)
	})

	t.Run("GetByRefreshToken with an invalid access", func(t *testing.T) {
		res, err := store.GetByRefreshToken(ctx, secret.NewText("some-invalid-token"))

		assert.Nil(t, res)
		assert.ErrorIs(t, err, errNotFound)
	})

	t.Run("RemoveByAccessToken success", func(t *testing.T) {
		err := store.RemoveByAccessToken(ctx, ExampleAliceSession.AccessToken())
		assert.NoError(t, err)

		res, err := store.GetByAccessToken(ctx, ExampleAliceSession.AccessToken())
		assert.Nil(t, res)
		assert.ErrorIs(t, err, errNotFound)
	})

	t.Run("RemoveByAccessToken with an invalid token", func(t *testing.T) {
		err := store.RemoveByAccessToken(ctx, secret.NewText("some-invalid-token"))

		assert.NoError(t, err)
	})

	t.Run("Save success 2", func(t *testing.T) {
		err := store.Save(ctx, &ExampleAliceSession)

		assert.NoError(t, err)
	})

	t.Run("RemoveByRefreshToken success", func(t *testing.T) {
		err := store.RemoveByRefreshToken(ctx, ExampleAliceSession.RefreshToken())

		assert.NoError(t, err)

		res, err := store.GetByRefreshToken(ctx, ExampleAliceSession.RefreshToken())

		assert.Nil(t, res)
		assert.ErrorIs(t, err, errNotFound)
	})

	t.Run("RemoveByRefreshToken with an invalid token", func(t *testing.T) {
		err := store.RemoveByRefreshToken(ctx, secret.NewText("some-invalid-token"))

		assert.NoError(t, err)
	})
}
