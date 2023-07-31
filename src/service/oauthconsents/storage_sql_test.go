package oauthconsents

import (
	"context"
	"testing"
	"time"

	"github.com/Peltoche/neurone/src/tools/storage"
	"github.com/Peltoche/neurone/src/tools/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConsentSqlStorage(t *testing.T) {
	nowData := time.Now().UTC()

	sessionData := Consent{
		ID:        uuid.UUID("some-consent-id"),
		UserID:    uuid.UUID("some-user-id"),
		ClientID:  "some-id",
		Scopes:    []string{"scope-a", "scope-b"},
		CreatedAt: nowData,
	}

	db := storage.NewTestStorage(t)
	storage := newSQLStorage(db)

	t.Run("Create success", func(t *testing.T) {
		err := storage.Save(context.Background(), &sessionData)

		assert.NoError(t, err)
	})

	t.Run("GetByID success", func(t *testing.T) {
		res, err := storage.GetByID(context.Background(), "some-consent-id")

		require.NotNil(t, res)
		res.CreatedAt = res.CreatedAt.UTC()

		assert.NoError(t, err)
		assert.Equal(t, &sessionData, res)
	})

	t.Run("GetByID not found", func(t *testing.T) {
		res, err := storage.GetByID(context.Background(), "some-invalid-token")

		assert.NoError(t, err)
		assert.Nil(t, res)
	})
}