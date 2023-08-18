package oauthsessions

import (
	"context"
	"testing"
	"time"

	"github.com/myminicloud/myminicloud/src/tools/storage"
	"github.com/stretchr/testify/assert"
)

func TestSessionStorageStorage(t *testing.T) {
	db := storage.NewTestStorage(t)
	storage := newSqlStorage(db)

	nowData := time.Now().UTC()

	sessionData := Session{
		accessToken:      "some-access-token",
		accessCreatedAt:  nowData,
		accessExpiresAt:  nowData.Add(time.Hour),
		refreshToken:     "some-refresh-token",
		refreshCreatedAt: nowData,
		refreshExpiresAt: nowData.Add(10 * time.Hour),
		clientID:         "some-client-id",
		userID:           "some-user-id",
		scope:            "some-scope",
	}

	t.Run("Save success", func(t *testing.T) {
		err := storage.Save(context.Background(), &sessionData)
		assert.NoError(t, err)
	})
}
