package oauthcodes

import (
	"context"
	"testing"
	"time"

	"github.com/Peltoche/neurone/src/tools"
	"github.com/Peltoche/neurone/src/tools/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCodeStorage(t *testing.T) {
	ctx := context.Background()
	now := time.Now().UTC()

	codeExample := Code{
		Code:            "some-code",
		CreatedAt:       now,
		ExpiresAt:       now.Add(time.Hour),
		ClientID:        "some-client-id",
		UserID:          "some-user-id",
		RedirectURI:     "http://some-redirect.com/uri",
		Scope:           "some-scope",
		Challenge:       "some-challenge",
		ChallengeMethod: "plain",
	}

	tools := tools.NewMock(t)

	db, err := storage.NewSQliteDBWithMigrate(storage.Config{Path: t.TempDir() + "/test.db"}, tools)
	require.NoError(t, err)

	storage := newSqlStorage(db)

	t.Run("save", func(t *testing.T) {
		err := storage.Save(context.Background(), &codeExample)
		assert.NoError(t, err)
	})

	t.Run("GetByID ok", func(t *testing.T) {
		code, err := storage.GetByCode(context.Background(), "some-code")
		assert.NoError(t, err)
		assert.EqualValues(t, &codeExample, code)
	})

	t.Run("GetByID not found", func(t *testing.T) {
		code, err := storage.GetByCode(ctx, "some-invalid-code")
		assert.NoError(t, err)
		assert.Nil(t, code)
	})

	t.Run("RemoveByCode", func(t *testing.T) {
		err := storage.RemoveByCode(ctx, "some-code")
		assert.NoError(t, err)

		// Check that the code is no more available
		code, err := storage.GetByCode(ctx, "some-code")
		assert.NoError(t, err)
		assert.Nil(t, code)
	})

	t.Run("RemoveByCode invalid code", func(t *testing.T) {
		err := storage.RemoveByCode(ctx, "some-invalid-code")
		assert.NoError(t, err)
	})
}
