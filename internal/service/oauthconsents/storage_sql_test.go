package oauthconsents

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theduckcompany/duckcloud/internal/tools/storage"
	"github.com/theduckcompany/duckcloud/internal/tools/uuid"
)

func TestConsentSqlStorage(t *testing.T) {
	ctx := context.Background()

	db := storage.NewTestStorage(t)
	storage := newSQLStorage(db)

	t.Run("Create success", func(t *testing.T) {
		err := storage.Save(ctx, &ExampleAliceConsent)

		assert.NoError(t, err)
	})

	t.Run("GetByID success", func(t *testing.T) {
		res, err := storage.GetByID(ctx, ExampleAliceConsent.ID())

		require.NotNil(t, res)
		res.createdAt = res.createdAt.UTC()

		assert.NoError(t, err)
		assert.Equal(t, &ExampleAliceConsent, res)
	})

	t.Run("GetByID not found", func(t *testing.T) {
		res, err := storage.GetByID(ctx, "some-invalid-token")

		assert.NoError(t, err)
		assert.Nil(t, res)
	})

	t.Run("GetAllForUser success", func(t *testing.T) {
		res, err := storage.GetAllForUser(ctx, ExampleAliceConsent.UserID(), nil)

		assert.NoError(t, err)
		assert.Equal(t, []Consent{ExampleAliceConsent}, res)
	})

	t.Run("Delete success", func(t *testing.T) {
		err := storage.Delete(ctx, ExampleAliceConsent.ID())
		assert.NoError(t, err)

		res, err := storage.GetByID(ctx, ExampleAliceConsent.ID())
		assert.NoError(t, err)
		assert.Nil(t, res)
	})

	t.Run("Delete with an invalid id", func(t *testing.T) {
		err := storage.Delete(ctx, uuid.UUID("some-invalid-id"))
		assert.NoError(t, err)
	})
}
