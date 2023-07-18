package oauthsessions

import (
	"context"
	"testing"
	"time"

	"github.com/Peltoche/neurone/src/tools/storage"
	"github.com/stretchr/testify/assert"
)

func TestSessionStorageStorage(t *testing.T) {
	db := storage.NewTestStorage(t)
	storage := newSqlStorage(db)

	nowData := time.Now().UTC()

	sessionData := Session{
		AccessToken:      "some-access-token",
		AccessCreatedAt:  nowData,
		AccessExpiresAt:  nowData.Add(time.Hour),
		RefreshToken:     "some-refresh-token",
		RefreshCreatedAt: nowData,
		RefreshExpiresAt: nowData.Add(10 * time.Hour),
		ClientID:         "some-client-id",
		UserID:           "some-user-id",
		Scope:            "some-scope",
	}

	t.Run("Save success", func(t *testing.T) {
		err := storage.Save(context.Background(), &sessionData)
		assert.NoError(t, err)
	})
}
