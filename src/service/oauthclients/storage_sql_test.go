package oauthclients

import (
	"context"
	"testing"
	"time"

	"github.com/Peltoche/neurone/src/tools"
	"github.com/Peltoche/neurone/src/tools/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOauthClientsSQLStorage(t *testing.T) {
	ctx := context.Background()
	tools := tools.NewMock(t)

	now := time.Now().UTC()
	clientExample := Client{
		ID:             "some-client-id",
		Secret:         "some-secret",
		RedirectURI:    "some-url",
		UserID:         "some-user-id",
		CreatedAt:      now,
		Scopes:         []string{"scope-a"},
		Public:         true,
		SkipValidation: true,
	}

	db, err := storage.NewSQliteDBWithMigrate(storage.Config{Path: t.TempDir() + "/test.db"}, tools)
	require.NoError(t, err)

	storage := newSqlStorage(db)

	t.Run("GetByID not found", func(t *testing.T) {
		res, err := storage.GetByID(ctx, "some-invalid-id")
		assert.NoError(t, err)
		assert.Nil(t, res)
	})

	t.Run("Create", func(t *testing.T) {
		err := storage.Save(context.Background(), &clientExample)
		assert.NoError(t, err)
	})

	t.Run("GetByID success", func(t *testing.T) {
		res, err := storage.GetByID(ctx, "some-client-id")
		assert.NoError(t, err)
		assert.EqualValues(t, &clientExample, res)
	})
}
