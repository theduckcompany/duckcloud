package oauthcodes

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/theduckcompany/duckcloud/src/tools/storage"
)

func TestOauthCodeSQLStorage(t *testing.T) {
	ctx := context.Background()
	now := time.Now().UTC()

	codeExample := Code{
		code:            "some-code",
		createdAt:       now,
		expiresAt:       now.Add(time.Hour),
		clientID:        "some-client-id",
		userID:          "some-user-id",
		redirectURI:     "http://some-redirect.com/uri",
		scope:           "some-scope",
		challenge:       "some-challenge",
		challengeMethod: "plain",
	}

	db := storage.NewTestStorage(t)
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
