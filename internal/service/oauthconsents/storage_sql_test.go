package oauthconsents

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/service/oauthclients"
	"github.com/theduckcompany/duckcloud/internal/service/users"
	"github.com/theduckcompany/duckcloud/internal/tools/sqlstorage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

func TestConsentSqlStorage(t *testing.T) {
	ctx := context.Background()

	db := sqlstorage.NewTestStorage(t)
	storage := newSQLStorage(db)

	// Data
	user := users.NewFakeUser(t).BuildAndStore(ctx, db)
	oauthClient := oauthclients.NewFakeClient(t).CreatedBy(user).BuildAndStore(ctx, db)
	consent := NewFakeConsent(t).CreatedBy(user).WithClient(oauthClient).Build()

	t.Run("Create success", func(t *testing.T) {
		// Run
		err := storage.Save(ctx, consent)

		// Asserts
		require.NoError(t, err)
	})

	t.Run("GetByID success", func(t *testing.T) {
		// Run
		res, err := storage.GetByID(ctx, consent.ID())

		// Asserts
		require.NoError(t, err)
		assert.Equal(t, consent, res)
	})

	t.Run("GetByID not found", func(t *testing.T) {
		// Run
		res, err := storage.GetByID(ctx, "some-invalid-token")

		// Asserts
		assert.Nil(t, res)
		require.ErrorIs(t, err, errNotFound)
	})

	t.Run("GetAllForUser success", func(t *testing.T) {
		// Run
		res, err := storage.GetAllForUser(ctx, user.ID(), nil)

		// Asserts
		require.NoError(t, err)
		assert.Equal(t, []Consent{*consent}, res)
	})

	t.Run("Delete success", func(t *testing.T) {
		// Run
		err := storage.Delete(ctx, consent.ID())

		// Asserts
		require.NoError(t, err)

		res, err := storage.GetByID(ctx, consent.ID())
		assert.Nil(t, res)
		require.ErrorIs(t, err, errNotFound)
	})

	t.Run("Delete with an invalid id", func(t *testing.T) {
		// Run
		err := storage.Delete(ctx, uuid.UUID("some-invalid-id"))

		// Asserts
		require.NoError(t, err)
	})
}
